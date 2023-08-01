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
	Name           string   `yaml:"name"`
	Weight         int64    `yaml:"weight"`
	MinWaitingTime int      `yaml:"minwaitingtime"`
	Shoot          []string `yaml:"shoot"`
}

type Preprocessor interface {
	// TODO
}

type RequestConfig struct {
	Method         string                        `yaml:"method"`
	Headers        map[string]string             `yaml:"headers"`
	Tag            string                        `yaml:"tag"`
	Body           *string                       `yaml:"body"`
	Name           string                        `yaml:"name"`
	Uri            string                        `yaml:"uri"`
	Preprocessors  []Preprocessor                `yaml:"preprocessors"`
	Postprocessors []postprocessor.Postprocessor `yaml:"postprocessors"`
	Templater      string                        `yaml:"templater"`
}
