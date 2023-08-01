package scenario

import "github.com/yandex/pandora/core/register"

type VariableSource interface {
	GetName() string
	GetVariables() any
	Init() error
}

func RegisterVariableSource(name string, mwConstructor interface{}, defaultConfigOptional ...interface{}) {
	var ptr *VariableSource
	register.RegisterPtr(ptr, name, mwConstructor, defaultConfigOptional...)
}

type Storage struct {
	Map map[string]any
}

func (s *Storage) AddStorage(name string, storage any) {
	s.Map[name] = storage
}

func (s *Storage) GlobalVariables() map[string]any {
	return s.Map
}
