package scenario

import (
	"strings"
	"sync"
	"testing"
	"time"

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
    fields: [ "user_id", "name", "pass", "created_at" ]
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
    preprocessor:
      variables:
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
    preprocessor:
      variables:
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
    shoots: [
      auth(1),
      sleep(100),
      list(1),
      sleep(100),
      order(3)
    ]
  - name: scenario2
    weight: 10
    minwaitingtime: 1000
    shoots: [
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
	cfg, err := ParseAmmoConfig(reader)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"hostname": "localhost"}, cfg.Variables)
	assert.Equal(t, 2, len(cfg.VariableSources))
	assert.Equal(t, "users_src", cfg.VariableSources[0].GetName())

	assert.Equal(t, "filter_src", cfg.VariableSources[1].GetName())
	assert.Equal(t, 3, len(cfg.Requests))
	assert.Equal(t, "auth_req", cfg.Requests[0].Name)
	require.Equal(t, 2, len(cfg.Requests[0].Postprocessors))
	require.Equal(t, map[string]string{"httpAuthorization": "Http-Authorization", "contentType": "Content-Type|lower"}, cfg.Requests[0].Postprocessors[0].(*postprocessor.VarHeaderPostprocessor).Mapping)
	require.Equal(t, map[string]string{"token": "$.data.authToken"}, cfg.Requests[0].Postprocessors[1].(*postprocessor.VarJsonpathPostprocessor).Mapping)

	assert.Equal(t, "list_req", cfg.Requests[1].Name)
	assert.Equal(t, "order_req", cfg.Requests[2].Name)
	assert.Equal(t, 2, len(cfg.Scenarios))
	assert.Equal(t, "scenario1", cfg.Scenarios[0].Name)
	assert.Equal(t, "scenario2", cfg.Scenarios[1].Name)

}

func Test_spreadNames(t *testing.T) {
	tests := []struct {
		name      string
		input     []ScenarioConfig
		want      map[string]int
		wantTotal int
	}{
		{
			name:      "",
			input:     []ScenarioConfig{{Name: "a", Weight: 20}, {Name: "b", Weight: 30}, {Name: "c", Weight: 60}},
			want:      map[string]int{"a": 2, "b": 3, "c": 6},
			wantTotal: 11,
		},
		{
			name:      "",
			input:     []ScenarioConfig{{Name: "a", Weight: 100}, {Name: "b", Weight: 100}, {Name: "c", Weight: 100}},
			want:      map[string]int{"a": 1, "b": 1, "c": 1},
			wantTotal: 3,
		},
		{
			name:      "",
			input:     []ScenarioConfig{{Name: "a", Weight: 100}},
			want:      map[string]int{"a": 1},
			wantTotal: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total := spreadNames(tt.input)
			assert.Equalf(t, tt.want, got, "spreadNames(%v)", tt.input)
			assert.Equalf(t, tt.wantTotal, total, "spreadNames(%v)", tt.input)
		})
	}
}

func TestParseShootName(t *testing.T) {
	testCases := []struct {
		input    string
		wantName string
		wantCnt  int
		wantErr  bool
	}{
		{"shoot", "shoot", 1, false},
		{"shoot(5)", "shoot", 5, false},
		{"shoot(3,4,5)", "shoot", 3, false},
		{"shoot(5,6)", "shoot", 5, false},
		{"space test(7)", "space test", 7, false},
		{"symbol#(3)", "symbol#", 3, false},
		{"shoot(  9  )", "shoot", 9, false},
		{"shoot (6)", "shoot", 6, false},
		{"shoot()", "shoot", 1, false},
		{"shoot(abc)", "", 0, true},
		{"shoot(6", "", 0, true},
		{"shoot(6),", "", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			name, cnt, err := parseShootName(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantName, name, "Name does not match for input: %s", tc.input)
			assert.Equal(t, tc.wantCnt, cnt, "Count does not match for input: %s", tc.input)
		})
	}
}

func Test_convertScenarioToAmmo(t *testing.T) {
	req1 := RequestConfig{
		Method: "GET",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Name: "req1",
		Uri:  "https://example.com/api/endpoint",
	}
	req2 := RequestConfig{
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer abcdef",
		},
		Name: "req2",
		Uri:  "https://example.com/api/another-endpoint",
	}

	reqRegistry := map[string]RequestConfig{
		"req1": req1,
		"req2": req2,
	}

	tests := []struct {
		name    string
		sc      ScenarioConfig
		want    *Ammo
		wantErr bool
	}{
		{
			name: "",
			sc: ScenarioConfig{
				Name:           "testScenario",
				Weight:         1,
				MinWaitingTime: 1000,
				Shoots: []string{
					"req1",
					"req2",
					"req2(2)",
					"sleep(500)",
				},
			},
			want: &Ammo{
				name:           "testScenario",
				minWaitingTime: time.Millisecond * 1000,
				Requests: []Request{
					convertConfigToRequestWithSleep(req1, 0),
					convertConfigToRequestWithSleep(req2, 0),
					convertConfigToRequestWithSleep(req2, 0),
					convertConfigToRequestWithSleep(req2, time.Millisecond*500),
				},
			},
			wantErr: false,
		},
		{
			name: "Scenario with unknown request",
			sc: ScenarioConfig{
				Name:           "unknownScenario",
				Weight:         1,
				MinWaitingTime: 1000,
				Shoots: []string{
					"unknownReq",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertScenarioToAmmo(tt.sc, reqRegistry)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "convertScenarioToAmmo(%v, %v)", tt.sc, reqRegistry)
		})
	}
}

func convertConfigToRequestWithSleep(req RequestConfig, sleep time.Duration) Request {
	res := convertConfigToRequest(req)
	res.sleep = sleep
	return res
}
