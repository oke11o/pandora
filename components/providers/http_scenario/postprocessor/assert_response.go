package postprocessor

import (
	"net/http"
	"strings"
)

type errAssert struct {
	pattern string
	t       string
}

func (e *errAssert) Error() string {
	return "assert failed: " + e.t + " does not contain " + e.pattern
}

type AssertResponse struct {
	Headers map[string]string
	Body    []string
}

func (a AssertResponse) Process(reqMap map[string]any, resp *http.Response, body []byte) error {
	for _, v := range a.Body {
		if !strings.Contains(string(body), v) {
			return &errAssert{pattern: v, t: "body"}
		}
	}
	for k, v := range a.Headers {
		if !(strings.Contains(resp.Header.Get(k), v)) {
			return &errAssert{pattern: v, t: "header " + k}
		}
	}

	return nil
}

func NewAssertResponsePostprocessor(cfg AssertResponse) Postprocessor {
	return &cfg
}
