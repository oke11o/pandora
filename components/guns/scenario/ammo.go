package scenario

type VariableStorage map[string]string

type Step interface {
	GetURL() string
	GetMethod() string
	GetBody() []byte
	GetHeaders() map[string]string
	GetTag() string
	OutputParams() []string
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
	OutputParams() []string
	Name() string
}
