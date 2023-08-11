package scenario

import (
	"sync"

	"github.com/spf13/afero"

	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
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

		RegisterPostprocessor("var/jsonpath", postprocessor.NewVarJsonpathPostprocessor)
		RegisterPostprocessor("var/xpath", postprocessor.NewVarXpathPostprocessor)
		RegisterPostprocessor("var/header", postprocessor.NewVarHeaderPostprocessor)
		RegisterPostprocessor("assert/response", postprocessor.NewAssertResponsePostprocessor)
	})
}

func RegisterPostprocessor(name string, mwConstructor interface{}, defaultConfigOptional ...interface{}) {
	var ptr *postprocessor.Postprocessor
	register.RegisterPtr(ptr, name, mwConstructor, defaultConfigOptional...)
}
