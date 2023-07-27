package scenario

type Templater interface {
	Apply(parts *RequestParts, vs map[string]any, scenarioName, stepName string) error
}
