package postprocessor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/multierr"

	"github.com/PaesslerAG/jsonpath"
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

func (p *VarJsonpathPostprocessor) Process(reqMap map[string]any, _ *http.Response, body []byte) error {
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
		reqMap[k] = val
	}
	return err
}
