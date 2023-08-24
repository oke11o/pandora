package httpscenario

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yandex/pandora/lib/mp"

	httpscenario "github.com/yandex/pandora/components/guns/http_scenario"
)

type gunGetSetter struct {
	m    map[string]any
	iter mp.Iterator
}

func (g gunGetSetter) Get(path string) (any, error) {
	return mp.GetMapValue(g.m, path, g.iter)
}

func (g *gunGetSetter) Set(key string, val any) error {
	g.m[key] = val
	return nil
}

type Preprocessor struct {
	Variables map[string]string
	iterator  mp.Iterator
}

func (p *Preprocessor) Process(get map[string]httpscenario.GetSetter) (httpscenario.Getter, error) {
	if get == nil {
		return nil, errors.New("templateVars must not be nil")
	}
	vars := gunGetSetter{m: make(map[string]any), iter: p.iterator}
	for k, v := range p.Variables {
		reqName, path := extractRequestName(v)
		if k == "user_id" && path == "users[next].user_id" { // TODO: remove this hack
			index := p.iterator.Next("users[next].user_id")
			if index >= 10 {
				index %= 10
			}
			_ = vars.Set(k, index)
			continue
		}
		getter, exists := get[reqName]
		if !exists {
			return nil, fmt.Errorf("variable %s not found", k)
		}
		val, err := getter.Get(path)
		if err != nil {
			return vars, fmt.Errorf("failed to get value for %s: %w", k, err)

		}
		err = vars.Set(k, val)
		if err != nil {
			return vars, fmt.Errorf("failed to set value for %s: %w", k, err)
		}
	}
	return vars, nil
}

func extractRequestName(v string) (string, string) {
	result := strings.SplitN(v, ".", 2)
	if len(result) == 1 {
		return result[0], ""
	}
	return result[0], result[1]
}
