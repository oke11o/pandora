package scenario

type Templater interface {
	Apply(parts *requestParts, vs map[string]any, scenarioName, stepName string) error
}
