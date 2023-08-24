package postprocessor

import (
	"net/http"

	httpscenario "github.com/yandex/pandora/components/guns/http_scenario"
)

type Config struct {
	Mapping map[string]string
}

type Postprocessor interface {
	Process(request httpscenario.Setter, resp *http.Response, body []byte) error
}
