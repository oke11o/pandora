package scenario

import (
	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
)

type AmmoConfig struct {
	Variables       map[string]string
	VariableSources []VariableSource `config:"variable_sources"`
	Requests        []RequestConfig
	Scenarios       []ScenarioConfig
}

type ScenarioConfig struct {
	Name           string
	Weight         int64
	MinWaitingTime int64 `config:"min_waiting_time"`
	Shoots         []string
}

type RequestConfig struct {
	Name           string
	Method         string
	Headers        map[string]string
	Tag            string
	Body           *string
	Uri            string
	Preprocessor   Preprocessor
	Postprocessors []postprocessor.Postprocessor
	Templater      string
}
