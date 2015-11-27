package engine

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/yandex/pandora/aggregate"
	"github.com/yandex/pandora/ammo"
	"github.com/yandex/pandora/config"
	"github.com/yandex/pandora/gun"
	"github.com/yandex/pandora/limiter"
	"github.com/yandex/pandora/utils"
)

const (
	MaximumUsers = 100
)

type User struct {
	Name       string
	Ammunition ammo.Provider
	Results    aggregate.ResultListener
	Limiter    limiter.Limiter
	Gun        gun.Gun
}

func (u *User) Run(ctx context.Context) error {
	log.Printf("Starting user: %s\n", u.Name)
	defer func() {
		log.Printf("Exit user: %s\n", u.Name)
	}()
	control := u.Limiter.Control()
	source := u.Ammunition.Source()
	sink := u.Results.Sink()
loop:
	for {
		select {
		case j, more := <-source:
			if !more {
				log.Println("Ammo ended")
				break loop
			}
			_, more = <-control
			if more {
				u.Gun.Shoot(ctx, j, sink)
			} else {
				log.Println("Limiter ended.")
				break loop
			}
		case <-ctx.Done():
			break loop
		}
	}
	return nil
}

type UserPool struct {
	name              string
	userLimiterConfig *config.Limiter
	gunConfig         *config.Gun
	ammunition        ammo.Provider
	results           aggregate.ResultListener
	startupLimiter    limiter.Limiter
	users             []*User
	done              chan bool
}

func NewUserPoolFromConfig(cfg *config.UserPool) (up *UserPool, err error) {
	if cfg == nil {
		return nil, fmt.Errorf("no pool config provided")
	}

	ammunition, err := GetAmmoProvider(cfg.AmmoProvider)
	if err != nil {
		return nil, err
	}
	results, err := GetResultListener(cfg.ResultListener)
	if err != nil {
		return nil, err
	}
	startupLimiter, err := GetLimiter(cfg.StartupLimiter)
	if err != nil {
		return nil, err
	}
	up = &UserPool{
		name:              cfg.Name,
		ammunition:        ammunition,
		results:           results,
		startupLimiter:    startupLimiter,
		gunConfig:         cfg.Gun,
		userLimiterConfig: cfg.UserLimiter,
	}
	return
}

func (up *UserPool) Start(ctx context.Context) error {
	// userCtx will be canceled when all users finished their execution
	userCtx, resultCancel := context.WithCancel(ctx)

	userPromises := utils.Promises{}
	utilsPromises := utils.Promises{
		utils.PromiseCtx(ctx, up.ammunition.Start),
		utils.PromiseCtx(userCtx, up.results.Start),
		utils.PromiseCtx(ctx, up.startupLimiter.Start),
	}

	i := 0
	for range up.startupLimiter.Control() {
		if i > MaximumUsers {
			return fmt.Errorf("Maximum users %d exceeded", MaximumUsers)
		}
		l, err := GetLimiter(up.userLimiterConfig)
		if err != nil {
			return fmt.Errorf("could not make a user limiter from config due to %s", err)
		}
		g, err := GetGun(up.gunConfig)
		if err != nil {
			return fmt.Errorf("could not make a gun from config due to %s", err)
		}
		u := &User{
			Name:       fmt.Sprintf("%s/%d", up.name, i),
			Ammunition: up.ammunition,
			Results:    up.results,
			Limiter:    l,
			Gun:        g,
		}
		utilsPromises = append(utilsPromises, utils.PromiseCtx(userCtx, l.Start))
		userPromises = append(userPromises, utils.PromiseCtx(ctx, u.Run))

		i++
	}
	// FIXME: wrong logic here
	log.Println("Started all users. Waiting for them")
	err := <-userPromises.All()
	resultCancel() // stop result listener when all users finished

	err2 := utilsPromises.All()
	if err2 != nil {
		fmt.Printf("%v", err2)
	}
	return err
}