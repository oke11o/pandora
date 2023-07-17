package scenario

import (
	"fmt"
	"io"
	"log"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/yandex/pandora/core/config"
)

type Step struct {
}

type InputParam string

type OutputParams string

type Ammo struct {
	Steps        []Step
	InputParams  []InputParam
	OutputParams []OutputParams
}

func (a *Ammo) Reset() {
	a.InputParams = []InputParam{}
	a.OutputParams = []OutputParams{}
}

type AmmoConfig struct {
	Variables       map[string]string `yaml:"variables"`
	VariableSources []VariableSource  `yaml:"variablesources"`
	Requests        []Request         `yaml:"requests"`
	Scenarios       []Scenario        `yaml:"scenarios"`
}

type Scenario struct {
	Name           string   `yaml:"name"`
	Weight         string   `yaml:"weight"`
	MinWaitingTime int      `yaml:"minwaitingtime"`
	Shoot          []string `yaml:"shoot"`
}

type Preprocessor interface {
	// TODO
}

type Postprocessor interface {
	// TODO
}

type Request struct {
	Method         string            `yaml:"method"`
	Headers        map[string]string `yaml:"headers"`
	Tag            string            `yaml:"tag"`
	Body           string            `yaml:"body"`
	Name           string            `yaml:"name"`
	Uri            string            `yaml:"uri"`
	Preprocessors  []Preprocessor    `yaml:"preprocessors"`
	Postprocessors []Postprocessor   `yaml:"postprocessors"`
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
