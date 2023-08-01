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

func (v *VariableSourceJson) GetName() string {
	return v.Name
}

func (v *VariableSourceJson) GetVariables() any {
	return v.Mapping
}

func (v *VariableSourceJson) Init() error {
	panic("implement me")
	return nil
}

func NewVSJson(cfg VariableSourceJson, fs afero.Fs) (VariableSource, error) {
	cfg.fs = fs
	return &cfg, nil
}

var _ VariableSource = (*VariableSourceJson)(nil)
