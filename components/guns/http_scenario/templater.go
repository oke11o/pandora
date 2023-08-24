package httpscenario

type Templater interface {
	Apply(parts *requestParts, vs any, scenarioName, stepName string) error
}
