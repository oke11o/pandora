package scenario

import (
	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
)

type AmmoConfig struct {
	Variables       map[string]string `yaml:"variables"`
	VariableSources []VariableSource  `yaml:"variablesources"`
	Requests        []RequestConfig   `yaml:"requests"`
	Scenarios       []ScenarioConfig  `yaml:"scenarios"`
}

type ScenarioConfig struct {
	Name           string   `yaml:"name" hcl:"name,label"`
	Weight         int64    `yaml:"weight" hcl:"weight"`
	MinWaitingTime int64    `yaml:"minwaitingtime" hcl:"min_waiting_time"`
	Shoots         []string `yaml:"shoot" hcl:"shoot"`
}

type RequestConfig struct {
	Name           string                        `yaml:"name"`
	Method         string                        `yaml:"method"`
	Headers        map[string]string             `yaml:"headers"`
	Tag            string                        `yaml:"tag"`
	Body           *string                       `yaml:"body"`
	Uri            string                        `yaml:"uri"`
	Preprocessor   Preprocessor                  `yaml:"preprocessor"`
	Postprocessors []postprocessor.Postprocessor `yaml:"postprocessors"`
	Templater      string                        `yaml:"templater"`
}
