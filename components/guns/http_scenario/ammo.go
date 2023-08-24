package httpscenario

import (
	"net/http"
	"time"

	"github.com/yandex/pandora/lib/mp"
)

type GetSetter interface {
	Getter
	Setter
}

type Getter interface {
	Get(path string) (any, error)
}

type Setter interface {
	Set(key string, val any) error
}

type gunGetSetter struct {
	m    map[string]any
	iter mp.Iterator
}

func (g *gunGetSetter) Get(path string) (any, error) {
	return mp.GetMapValue(g.m, path, g.iter)
}

func (g *gunGetSetter) Set(key string, val any) error {
	g.m[key] = val
	return nil
}

type Preprocessor interface {
	// Process is called before request is sent
	// templateVars - variables from template. Can be modified
	// sourceVars - variables from sources. Must NOT be modified
	//Process(templateVars map[string]any, sourceVars map[string]any) error
	Process(get map[string]GetSetter) (Getter, error)
}

type Postprocessor interface {
	Process(request Setter, resp *http.Response, body []byte) error
}

type VariableStorage interface {
	Variables() map[string]any
}

type Step interface {
	GetName() string
	GetURL() string
	GetMethod() string
	GetBody() []byte
	GetHeaders() map[string]string
	GetTag() string
	GetTemplater() string
	GetPostProcessors() []Postprocessor
	Preprocessor() Preprocessor
	GetSleep() time.Duration
}

type requestParts struct {
	URL     string
	Method  string
	Body    []byte
	Headers map[string]string
}

type Ammo interface {
	Steps() []Step
	ID() uint64
	Sources() VariableStorage
	Name() string
	GetMinWaitingTime() time.Duration
}
