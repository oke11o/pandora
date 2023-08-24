package postprocessor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PaesslerAG/jsonpath"
	multierr "github.com/hashicorp/go-multierror"

	httpscenario "github.com/yandex/pandora/components/guns/http_scenario"
)

type VarJsonpathPostprocessor struct {
	Mapping map[string]string
}

func NewVarJsonpathPostprocessor(cfg Config) Postprocessor {
	return &VarJsonpathPostprocessor{
		Mapping: cfg.Mapping,
	}
}

func (p *VarJsonpathPostprocessor) ReturnedParams() []string {
	result := make([]string, len(p.Mapping))
	for k := range p.Mapping {
		result = append(result, k)
	}
	return result
}

func (p *VarJsonpathPostprocessor) Process(request httpscenario.Setter, _ *http.Response, body []byte) error {
	var data any
	err := json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	for k, path := range p.Mapping {
		val, e := jsonpath.Get(path, data)
		if e != nil {
			err = multierr.Append(err, fmt.Errorf("failed to get value by jsonpath %s: %w", path, e))
			continue
		}
		e = request.Set(k, val)
		if e != nil {
			err = multierr.Append(err, fmt.Errorf("failed to set `%s` value %s: %w", k, val, e))
		}
	}
	return err
}
