package scenario

type HTMLTemplater struct {
}

func (j *HTMLTemplater) Apply(parts *RequestParts, vs map[string]any, scenarioName, stepName string) error {
	//TODO implement me
	panic("implement me")
}

func NewHTMLTemplater() Templater {
	return &HTMLTemplater{}
}
