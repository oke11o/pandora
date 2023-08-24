package postprocessor

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	multierr "github.com/hashicorp/go-multierror"
	"golang.org/x/net/html"

	httpscenario "github.com/yandex/pandora/components/guns/http_scenario"
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

func (p *VarXpathPostprocessor) Process(request httpscenario.Setter, _ *http.Response, body []byte) error {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	for k, path := range p.Mapping {
		val, e := p.getValuesFromDOM(doc, path)
		if e != nil {
			err = multierr.Append(err, fmt.Errorf("failed to get value by jsonpath %s: %w", path, e))
			continue
		}
		e = request.Set(k, val)
		if e != nil {
			err = multierr.Append(err, fmt.Errorf("failed to set `%s` value %s: %w", k, val, e))
		}
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
