package scenario

import (
	"net/http"
	"time"
)

type Postprocessor interface {
	Process(reqMap map[string]any, resp *http.Response, body []byte) error
}

type VariableStorage map[string]any

type Step interface {
	GetName() string
	GetURL() string
	GetMethod() string
	GetBody() []byte
	GetHeaders() map[string]string
	GetTag() string
	GetTemplater() string
	GetPostProcessors() []Postprocessor
	GetSleep() time.Duration
}

type RequestParts struct {
	URL     string
	Method  string
	Body    []byte
	Headers map[string]string
}

// TODO: Not used yet
type Ammo interface {
	Steps() []Step
	ID() uint64
	VariableStorage() VariableStorage
	Name() string
}
