package httpscenario

import (
	"errors"
	"fmt"

	"github.com/yandex/pandora/lib/mp"
)

type Preprocessor struct {
	Variables map[string]string
	iterator  mp.Iterator
}

func (p *Preprocessor) Process(templateVars map[string]any) (map[string]any, error) {
	if templateVars == nil {
		return nil, errors.New("templateVars must not be nil")
	}
	result := make(map[string]any, len(p.Variables))
	for k, v := range p.Variables {
		if k == "user_id" && v == "source.users[next].user_id" { // TODO: remove this hack
			index := p.iterator.Next("users[next].user_id")
			if index >= 10 {
				index %= 10
			}
			result[k] = index + 1
			continue
		}
		val, err := mp.GetMapValue(templateVars, v, p.iterator)
		if err != nil {
			return nil, fmt.Errorf("failed to get value for %s: %w", k, err)

		}
		result[k] = val
	}
	return result, nil
}
