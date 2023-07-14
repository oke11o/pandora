package scenario

import (
	"sync"

	"github.com/spf13/afero"

	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/register"
)

var once = &sync.Once{}

func Import(fs afero.Fs) {
	once.Do(func() {
		register.Provider("http/scenario", func(cfg Config) (core.Provider, error) {
			return NewProvider(fs, cfg)
		})

		RegisterVariableSource("file/csv", func(cfg VariableSourceCsv) (VariableSource, error) {
			return NewVSCSV(cfg, fs)
		})

		RegisterVariableSource("file/json", func(cfg VariableSourceJson) (VariableSource, error) {
			return NewVSJson(cfg, fs)
		})
	})

	//register.Provider("http/scenario", func(cfg Config) (core.Provider, error) {
	//	return NewProvider(fs, cfg)
	//})
	//
	//RegisterVariableSource("file/csv", func(cfg VariableSourceCsv) (VariableSource, error) {
	//	return NewVSCSV(cfg, fs)
	//})
	//
	//RegisterVariableSource("file/json", func(cfg VariableSourceJson) (VariableSource, error) {
	//	return NewVSJson(cfg, fs)
	//})
}
