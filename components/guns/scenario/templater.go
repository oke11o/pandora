package scenario

import "net/http"

type Templater struct {
}

func (t Templater) Apply(parts *RequestParts, vs map[string]string) error {
	parts.URL = t.templateString(parts.URL, vs)
	for k, v := range parts.Headers {
		parts.Headers[k] = t.templateString(v, vs)
	}
	if parts.Body != nil {
		parts.Body = t.templateBytes(parts.Body, vs)
	}
	// TODO: handle error
	return nil
}

func (t Templater) templateString(val string, vs map[string]string) string {
	// TODO:
	return val
}

func (t Templater) templateBytes(val []byte, vs map[string]string) []byte {
	// TODO:
	return val
}

func (t Templater) SaveResponseToVS(resp *http.Response, varPrefix string, params []string, vs map[string]string) error {
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
			// TODO:
		}
	}
	return nil
}

func (t Templater) needsParseResponse(params []string) bool {
	return len(params) > 0
}
