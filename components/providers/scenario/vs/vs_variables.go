package vs

import (
	"github.com/yandex/pandora/components/providers/scenario/templater"
)

type VariableSourceVariables struct {
	Name      string
	Variables map[string]any
}

func (v *VariableSourceVariables) GetName() string {
	return v.Name
}

func (v *VariableSourceVariables) GetVariables() any {
	return v.Variables
}

func (v *VariableSourceVariables) Init() error {
	v.recursiveCompute(v.Variables)
	return nil
}

func (v *VariableSourceVariables) recursiveCompute(input map[string]any) {
	for key, val := range input {
		switch value := val.(type) {
		case string:
			input[key] = v.execTemplateFunc(value)
		case map[string]any:
			v.recursiveCompute(value)
		case map[string]string:
			for k, vv := range value {
				value[k] = v.execTemplateFunc(vv)
			}
			input[key] = value
		case []string:
			for i, vv := range value {
				value[i] = v.execTemplateFunc(vv)
			}
			input[key] = value
		}
	}
}

func (v *VariableSourceVariables) execTemplateFunc(in string) string {
	fun, args := templater.ParseFunc(in)
	if fun == nil {
		return in
	}
	value, err := templater.ExecTemplateFunc(fun, args)
	if err != nil {
		return in
	}
	return value
}
