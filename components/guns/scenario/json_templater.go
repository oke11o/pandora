package scenario

type JsonTemplater struct {
}

func (j *JsonTemplater) Apply(parts *RequestParts, vs map[string]any, scenarioName, stepName string) error {
	//TODO implement me
	panic("implement me")
}

func NewJsonTempalter() Templater {
	return &JsonTemplater{}
}
