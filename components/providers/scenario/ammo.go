package scenario

import (
	"fmt"
	"io"
	"log"

	"github.com/yandex/pandora/components/guns/scenario"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/yandex/pandora/core/config"
)

type Ammo struct {
	InputParams    []string
	returnedParams []string

	Requests []Request `yaml:"requests"`
	Id       uint64    `yaml:"id"`
	name     string    `yaml:"name"`
}

func (a *Ammo) Steps() []scenario.Step {
	result := make([]scenario.Step, 0)
	for _, req := range a.Requests {
		result = append(result, &req)
	}
	return result
}

func (a *Ammo) ID() uint64 {
	return a.Id
}

func (a *Ammo) VariableStorage() scenario.VariableStorage {
	return map[string]string{}
}

func (a *Ammo) Name() string {
	return a.name
}

var _ scenario.Ammo = (*Ammo)(nil)

type AmmoConfig struct {
	Variables       map[string]string `yaml:"variables"`
	VariableSources []VariableSource  `yaml:"variablesources"`
	Requests        []Request         `yaml:"requests"`
	Scenarios       []Scenario        `yaml:"scenarios"`
}

type Scenario struct {
	Name           string   `yaml:"name"`
	Weight         int64    `yaml:"weight"`
	MinWaitingTime int      `yaml:"minwaitingtime"`
	Shoot          []string `yaml:"shoot"`
}

type Preprocessor interface {
	// TODO
}

type Postprocessor interface {
	ReturnedParams() []string
	// TODO
}

type Request struct {
	Method         string            `yaml:"method"`
	Headers        map[string]string `yaml:"headers"`
	Tag            string            `yaml:"tag"`
	Body           *string           `yaml:"body"`
	Name           string            `yaml:"name"`
	Uri            string            `yaml:"uri"`
	Preprocessors  []Preprocessor    `yaml:"preprocessors"`
	Postprocessors []Postprocessor   `yaml:"postprocessors"`
	returnedParams []string
	expectedParams []string
}

var _ scenario.Step = (*Request)(nil)

func (r *Request) GetMethod() string {
	return r.Method
}

func (r *Request) GetBody() []byte {
	if r.Body == nil {
		return nil
	}
	return []byte(*r.Body)
}

func (r *Request) GetHeaders() map[string]string {
	return r.Headers
}

func (r *Request) GetTag() string {
	return r.Tag
}

func (r *Request) GetURL() string {
	return r.Uri
}

func (r *Request) ReturnedParams() []string {
	return r.returnedParams
}

func parseAmmoConfig(file io.Reader) (AmmoConfig, error) {
	var ammoCfg AmmoConfig
	const op = "scenario/decoder.parseAmmoConfig"
	data := make(map[string]any)
	bytes, err := io.ReadAll(file)
	if err != nil {
		return ammoCfg, fmt.Errorf("%s, io.ReadAll, %w", op, err)
	}
	err = yaml.Unmarshal(bytes, &data)
	if err != nil {
		return ammoCfg, fmt.Errorf("%s, yaml.Unmarshal, %w", op, err)
	}
	err = config.DecodeAndValidate(data, &ammoCfg)
	if err != nil {
		log.Fatal("Config decode failed", zap.Error(err))
	}
	return ammoCfg, nil
}
