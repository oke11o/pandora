package scenario

import (
	"github.com/spf13/afero"
)

type VariableSourceCsv struct {
	Name    string         `yaml:"name"`
	File    string         `yaml:"file"`
	Mapping map[string]any `yaml:"mapping"`
	fs      afero.Fs
}

func (v VariableSourceCsv) GetName() string {
	return v.Name
}

func (v VariableSourceCsv) GetMapping() map[string]any {
	return v.Mapping
}

func NewVSCSV(cfg VariableSourceCsv, fs afero.Fs) (VariableSource, error) {
	cfg.fs = fs
	return &cfg, nil
}
