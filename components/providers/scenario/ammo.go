package scenario

import (
	"time"

	"github.com/yandex/pandora/components/guns/scenario"
)

var _ scenario.Ammo = (*Ammo)(nil)

type Ammo struct {
	InputParams    []string
	returnedParams []string

	Requests       []Request
	Id             uint64
	name           string
	minWaitingTime time.Duration
}

func (a *Ammo) GetMinWaitingTime() time.Duration {
	return a.minWaitingTime
}

func (a *Ammo) Steps() []scenario.Step {
	result := make([]scenario.Step, 0)
	for i := range a.Requests {
		result = append(result, &a.Requests[i])
	}
	return result
}

func (a *Ammo) ID() uint64 {
	return a.Id
}

func (a *Ammo) VariableStorage() scenario.VariableStorage {
	return map[string]any{}
}

func (a *Ammo) Name() string {
	return a.name
}

type Request struct {
	method         string
	headers        map[string]string
	tag            string
	body           *string
	name           string
	uri            string
	preprocessors  []Preprocessor
	postprocessors []scenario.Postprocessor
	templater      string
	sleep          time.Duration
}

func (r *Request) GetPostProcessors() []scenario.Postprocessor {
	return r.postprocessors
}

func (r *Request) GetTemplater() string {
	return r.templater
}

var _ scenario.Step = (*Request)(nil)

func (r *Request) GetName() string {
	return r.name
}
func (r *Request) GetMethod() string {
	return r.method
}

func (r *Request) GetBody() []byte {
	if r.body == nil {
		return nil
	}
	return []byte(*r.body)
}

func (r *Request) GetHeaders() map[string]string {
	return r.headers
}

func (r *Request) GetTag() string {
	return r.tag
}

func (r *Request) GetURL() string {
	return r.uri
}

func (r *Request) GetSleep() time.Duration {
	return r.sleep
}
