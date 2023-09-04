package httpscenario

import (
	"io"
	"net/http"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.22.1 --inpackage --name=Preprocessor --filename=mock_preprocessor.go
//go:generate go run github.com/vektra/mockery/v2@v2.22.1 --inpackage --name=Postprocessor --filename=mock_postprocessor.go
//go:generate go run github.com/vektra/mockery/v2@v2.22.1 --inpackage --name=Step --filename=mock_step.go
//go:generate go run github.com/vektra/mockery/v2@v2.22.1 --inpackage --name=Ammo --filename=mock_ammo.go

type Preprocessor interface {
	// Process is called before request is sent
	// templateVars - variables from template. Can be modified
	// sourceVars - variables from sources. Must NOT be modified
	Process(templateVars map[string]any) (map[string]any, error)
}

type Postprocessor interface {
	Process(resp *http.Response, body io.Reader) (map[string]any, error)
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
