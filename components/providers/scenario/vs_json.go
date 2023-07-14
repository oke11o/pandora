package scenario

import (
	"github.com/spf13/afero"
)

type VariableSourceJson struct {
	Name    string         `yaml:"name"`
	File    string         `yaml:"file"`
	Mapping map[string]any `yaml:"mapping"`
	fs      afero.Fs
}

func (v VariableSourceJson) GetName() string {
	return v.Name
}

func (v VariableSourceJson) GetMapping() map[string]any {
	return v.Mapping
}

func NewVSJson(cfg VariableSourceJson, fs afero.Fs) (VariableSource, error) {
	cfg.fs = fs
	return &cfg, nil
}
