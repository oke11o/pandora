package postprocessor

import (
	"bytes"
	"net/http"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
)

type VarXpathPostprocessor struct {
	Mapping map[string]string
}

func NewVarXpathPostprocessor(cfg Config) Postprocessor {
	return &VarXpathPostprocessor{
		Mapping: cfg.Mapping,
	}
}

func (p *VarXpathPostprocessor) ReturnedParams() []string {
	result := make([]string, len(p.Mapping))
	for k := range p.Mapping {
		result = append(result, k)
	}
	return result
}

func (p *VarXpathPostprocessor) Process(reqMap map[string]any, _ *http.Response, body []byte) error {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	for k, path := range p.Mapping {
		values, err := p.getValuesFromDOM(doc, path)
		if err != nil {
			return err
		}
		reqMap[k] = values
	}
	return nil
}

func (p *VarXpathPostprocessor) getValuesFromDOM(doc *html.Node, xpathQuery string) (any, error) {
	expr, err := xpath.Compile(xpathQuery)
	if err != nil {
		return nil, err
	}

	iter := expr.Evaluate(htmlquery.CreateXPathNavigator(doc)).(*xpath.NodeIterator)

	var values []string
	for iter.MoveNext() {
		node := iter.Current()
		values = append(values, node.Value())
	}

	return values, nil
}
