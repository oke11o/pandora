package scenario

import "github.com/yandex/pandora/core/register"

type VariableSource interface {
	GetName() string
	GetMapping() map[string]any
}

func RegisterVariableSource(name string, mwConstructor interface{}, defaultConfigOptional ...interface{}) {
	var ptr *VariableSource
	register.RegisterPtr(ptr, name, mwConstructor, defaultConfigOptional...)
}
