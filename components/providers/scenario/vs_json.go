package scenario

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/afero"
)

type VariableSourceJson struct {
	Name  string `yaml:"name"`
	File  string `yaml:"file"`
	fs    afero.Fs
	store any
}

func (v *VariableSourceJson) GetName() string {
	return v.Name
}

func (v *VariableSourceJson) GetVariables() any {
	return v.store
}

func (v *VariableSourceJson) Init() (err error) {
	const op = "VariableSourceJson.Init"
	var file afero.File
	file, err = v.fs.Open(v.File)
	if err != nil {
		return fmt.Errorf("%s fs.Open %w", op, err)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%s multiple errors faced: %w, with close err: %s", op, err, closeErr)
			} else {
				err = fmt.Errorf("%s, %w", op, closeErr)
			}
		}
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("%s io.ReadAll %w", op, err)
	}
	err = json.Unmarshal(data, &v.store)
	if err != nil {
		return fmt.Errorf("%s readCsv %w", op, err)
	}

	return nil
}

func NewVSJson(cfg VariableSourceJson, fs afero.Fs) (VariableSource, error) {
	cfg.fs = fs
	return &cfg, nil
}

var _ VariableSource = (*VariableSourceJson)(nil)
