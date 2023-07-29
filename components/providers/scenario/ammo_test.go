package scenario

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
	"github.com/yandex/pandora/core/plugin/pluginconfig"
)

const exampleAmmoFile = `
variables:
  hostname: localhost

variablesources:
  - type: "file/csv"
    name: "users_src"
    file: "_files/users.csv"
    mapping:
      users: [ "user_id", "name", "pass", "created_at" ]
  - type: "file/json"
    name: "filter_src"
    file: "_files/filter.json"

requests:
  - name: "auth_req"
    uri: '/auth'
    method: POST
    headers:
      Useragent: Tank
      ContentType: "application/json"
      Hostname: "{{hostname}}"
    tag: auth
    body: '{"user_name": {{source.users_src.users[next].name}}, "user_pass": {{source.users_src.users[next].pass}} }'
    templater: text
    postprocessors:
      - type: var/header
        mapping:
          httpAuthorization: "Http-Authorization"
          contentType: "Content-Type|lower"
      - type: 'var/jsonpath'
        mapping:
          token: "$.data.authToken"

  - name: list_req
    preprocessors:
      - type: prepare
        mapping:
          filter: source.filter_src.list[rand]
    uri: '/list/?{{filter|query}}'
    method: GET
    headers:
      Useragent: "Tank"
      ContentType: "application/json"
      Hostname: "{{hostname}}"
      Authorization: "Bearer {{request.auth_req.token}}"
    tag: list
    postprocessors:
      - type: var/jsonpath
        mapping:
          items: $.data.items

  - name: order_req
    preprocessors:
      - type: prepare
        mapping:
          item: list_req.items.items[rand]
    uri: '/order'
    tag: order	
    method: POST
    headers:
      Useragent: "Tank"
      ContentType: "application/json"
      Hostname: "{{hostname}}"
      Authorization: "Bearer {{request.auth_req.token}}"
    body: "{}"
    postprocessors:
      - type: var/jsonpath
        mapping:
          delivery_id: $.data.delivery_id

scenarios:
  - name: scenario1
    weight: 50
    minwaitingtime: 1000
    shoot: [
      auth(1),
      sleep(100),
      list(1),
      sleep(100),
      order(3)
    ]
  - name: scenario2
    weight: 10
    minwaitingtime: 1000
    shoot: [
      auth(1),
      sleep(100),
      list(1),
      sleep(100),
      order(3)
    ]
`

var testOnce = &sync.Once{}

func Test_parseAmmoConfig(t *testing.T) {
	Import(nil)
	testOnce.Do(func() {
		pluginconfig.AddHooks()
	})

	reader := strings.NewReader(exampleAmmoFile)
	cfg, err := parseAmmoConfig(reader)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"hostname": "localhost"}, cfg.Variables)
	assert.Equal(t, 2, len(cfg.VariableSources))
	assert.Equal(t, "users_src", cfg.VariableSources[0].GetName())
	assert.Equal(t, map[string]any{"users": []any{"user_id", "name", "pass", "created_at"}}, cfg.VariableSources[0].GetMapping())

	assert.Equal(t, "filter_src", cfg.VariableSources[1].GetName())
	assert.Equal(t, 3, len(cfg.Requests))
	assert.Equal(t, "auth_req", cfg.Requests[0].Name)
	require.Equal(t, 2, len(cfg.Requests[0].Postprocessors))
	require.Equal(t, 2, len(cfg.Requests[0].GetPostProcessors()))
	require.Equal(t, map[string]string{"httpAuthorization": "Http-Authorization", "contentType": "Content-Type|lower"}, cfg.Requests[0].Postprocessors[0].(*postprocessor.VarHeaderPostprocessor).Mapping)
	require.Equal(t, map[string]string{"token": "$.data.authToken"}, cfg.Requests[0].Postprocessors[1].(*postprocessor.VarJsonpathPostprocessor).Mapping)

	assert.Equal(t, "list_req", cfg.Requests[1].Name)
	assert.Equal(t, "order_req", cfg.Requests[2].Name)
	assert.Equal(t, 2, len(cfg.Scenarios))
	assert.Equal(t, "scenario1", cfg.Scenarios[0].Name)
	assert.Equal(t, "scenario2", cfg.Scenarios[1].Name)

}
