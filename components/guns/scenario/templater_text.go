package scenario

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

func NewTextTempalter() Templater {
	return &TextTemplater{}
}

type TextTemplater struct {
	templatesCache sync.Map
}

func (t *TextTemplater) Apply(parts *RequestParts, vs map[string]any, scenarioName, stepName string) error {
	const op = "scenario/TextTemplater.Apply"
	tmpl, err := t.getTemplate(parts.URL, scenarioName, stepName, "url")
	if err != nil {
		return fmt.Errorf("%s, template.New, %w", op, err)
	}

	strBuilder := &strings.Builder{}
	err = tmpl.Execute(strBuilder, vs)
	if err != nil {
		return fmt.Errorf("%s, template.Execute url, %w", op, err)
	}
	parts.URL = strBuilder.String()
	strBuilder.Reset()

	for k, v := range parts.Headers {
		tmpl, err = t.getTemplate(v, scenarioName, stepName, k)
		if err != nil {
			return fmt.Errorf("%s, template.Execute Header %s, %w", op, k, err)
		}
		err = tmpl.Execute(strBuilder, vs)
		if err != nil {
			return fmt.Errorf("%s, template.Execute Header %s, %w", op, k, err)
		}
		parts.Headers[k] = strBuilder.String()
		strBuilder.Reset()
	}
	if parts.Body != nil {
		tmpl, err = t.getTemplate(string(parts.Body), scenarioName, stepName, "body")
		if err != nil {
			return fmt.Errorf("%s, template.Execute body, %w", op, err)
		}
		err = tmpl.Execute(strBuilder, vs)
		if err != nil {
			return fmt.Errorf("%s, template.Execute body, %w", op, err)
		}
		parts.Body = []byte(strBuilder.String())
		strBuilder.Reset()
	}
	return nil
}

func (t *TextTemplater) getTemplate(tmplBody, scenarioName, stepName, key string) (*template.Template, error) {
	urlKey := fmt.Sprintf("%s_%s_%s", scenarioName, stepName, key)
	tmpl, ok := t.templatesCache.Load(urlKey)
	if !ok {
		var err error
		tmpl, err = template.New(urlKey).Parse(tmplBody)
		if err != nil {
			return nil, fmt.Errorf("scenario/TextTemplater.Apply, template.New, %w", err)
		}
		t.templatesCache.Store(urlKey, tmpl)
	}
	return tmpl.(*template.Template), nil
}

func (t *TextTemplater) SaveResponseToVS(resp *http.Response, varPrefix string, params []string, vs map[string]any) error {
	headers := resp.Header
	for _, param := range params {
		if param == "status" {
			vs[varPrefix+".status"] = resp.Status
		} else if param == "headers" {
			for k, v := range headers {
				vs[varPrefix+".headers."+k] = v[0]
			}
		} else {
			vs[varPrefix+"."+param] = "TODO"
			// TODO: postprocessors job
		}
	}
	return nil
}

func (t *TextTemplater) needsParseResponse(params []string) bool {
	return len(params) > 0
}
