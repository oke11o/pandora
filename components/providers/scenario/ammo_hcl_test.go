package scenario

import (
	"fmt"
	"io"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"

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
      name: prepare
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
      name: prepare
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

func Test_parseHCLAmmoConfig(t *testing.T) {
	Import(nil)
	testOnce.Do(func() {
		pluginconfig.AddHooks()
	})

	fs := afero.NewOsFs()
	file, err := fs.Open("decode_sample_config_test.yml")
	require.NoError(t, err)
	defer file.Close()

	ammoConfig, err := ParseAmmoConfig(file)
	require.NoError(t, err)

	ammoHCL, err := ConvertAmmoToHCL(ammoConfig)
	require.NoError(t, err)

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&ammoHCL, f.Body())
	bytes := f.Bytes()

	goldenFile, err := fs.Open("decode_sample_config_test.golden.hcl")
	require.NoError(t, err)
	defer goldenFile.Close()
	goldenBytes, err := io.ReadAll(goldenFile)
	require.NoError(t, err)

	assert.Equal(t, string(goldenBytes), string(bytes))

	// Output:
	// variables = {
	//   hostname = "localhost"
	// }
	//
	// variablesource "users_src" "file/csv" {
	//   file             = "_files/users.csv"
	//   fields           = ["user_id", "name", "pass", "created_at"]
	//   skip_header      = false
	//   header_as_fields = false
	// }
	// variablesource "filter_src" "file/json" {
	//   file             = "_files/filter.json"
	//   fields           = null
	//   skip_header      = false
	//   header_as_fields = false
	// }
	//
	// request "auth_req" {
	//   method = "POST"
	//   headers = {
	//     ContentType = "application/json"
	//     Hostname    = "{{hostname}}"
	//     Useragent   = "Tank"
	//   }
	//   tag  = "auth"
	//   body = "{\"user_name\": {{source.users_src.users[next].name}}, \"user_pass\": {{source.users_src.users[next].pass}} }"
	//   uri  = "/auth"
	//
	//   preprocessor "" {
	//     variables = null
	//   }
	//
	//   postprocessor "var/header" {
	//     mapping = {
	//       contentType       = "Content-Type|lower"
	//       httpAuthorization = "Http-Authorization"
	//     }
	//   }
	//   postprocessor "var/jsonpath" {
	//     mapping = {
	//       token = "$.data.authToken"
	//     }
	//   }
	//
	//   templater = "text"
	// }
	// request "list_req" {
	//   method = "GET"
	//   headers = {
	//     Authorization = "Bearer {{request.auth_req.token}}"
	//     ContentType   = "application/json"
	//     Hostname      = "{{hostname}}"
	//     Useragent     = "Tank"
	//   }
	//   tag = "list"
	//   uri = "/list/?{{filter|query}}"
	//
	//   preprocessor "prepare" {
	//     variables = {
	//       filter = "source.filter_src.list[rand]"
	//     }
	//   }
	//
	//   postprocessor "var/jsonpath" {
	//     mapping = {
	//       items = "$.data.items"
	//     }
	//   }
	//
	//   templater = ""
	// }
	// request "order_req" {
	//   method = "POST"
	//   headers = {
	//     Authorization = "Bearer {{request.auth_req.token}}"
	//     ContentType   = "application/json"
	//     Hostname      = "{{hostname}}"
	//     Useragent     = "Tank"
	//   }
	//   tag  = "order"
	//   body = "{}"
	//   uri  = "/order"
	//
	//   preprocessor "prepare" {
	//     variables = {
	//       item = "list_req.items.items[rand]"
	//     }
	//   }
	//
	//   postprocessor "var/jsonpath" {
	//     mapping = {
	//       delivery_id = "$.data.delivery_id"
	//     }
	//   }
	//
	//   templater = ""
	// }
	//
	// scenario "scenario1" {
	//   weight           = 50
	//   min_waiting_time = 1000
	//   shoot            = ["auth(1)", "sleep(100)", "list(1)", "sleep(100)", "order(3)"]
	// }
	// scenario "scenario2" {
	//   weight           = 10
	//   min_waiting_time = 1000
	//   shoot            = ["auth(1)", "sleep(100)", "list(1)", "sleep(100)", "order(3)"]
	// }
}

func ExampleEncodeAmmoHCLVariablesSources() {
	app := AmmoHCL{
		Variables: map[string]string{"host": "localhost"},
		VariableSources: []SourceHCL{
			{
				Type:   "file/csv",
				Name:   "user_srs",
				File:   "users.json",
				Fields: []string{"id", "name", "email"},
			},
			{
				Type:   "file/json",
				Name:   "data_srs",
				File:   "datas.json",
				Fields: []string{"id", "name", "email"},
			},
		},
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&app, f.Body())
	bytes := f.Bytes()
	fmt.Printf("%s", bytes)

	// Output:
	// variables = {
	//   host = "localhost"
	// }
	//
	// variablesource "user_srs" "file/csv" {
	//   file             = "users.json"
	//   fields           = ["id", "name", "email"]
	//   skip_header      = false
	//   header_as_fields = false
	// }
	// variablesource "data_srs" "file/json" {
	//   file             = "datas.json"
	//   fields           = ["id", "name", "email"]
	//   skip_header      = false
	//   header_as_fields = false
	// }
}

const hclConfig = `
variables = {
  hostname = "localhost"
}

variablesource "users_src" "file/csv" {
  file             = "_files/users.csv"
  fields           = ["user_id", "name", "pass", "created_at"]
  skip_header      = false
  header_as_fields = false
}
variablesource "filter_src" "file/json" {
  file             = "_files/filter.json"
  fields           = null
  skip_header      = false
  header_as_fields = false
}

request "auth_req" {
  method = "POST"
  headers = {
    ContentType = "application/json"
    Hostname    = "{{hostname}}"
    Useragent   = "Tank"
  }
  tag  = "auth"
  body = "{\"user_name\": {{source.users_src.users[next].name}}, \"user_pass\": {{source.users_src.users[next].pass}} }"
  uri  = "/auth"

  postprocessor "var/header" {
    mapping = {
      contentType       = "Content-Type|lower"
      httpAuthorization = "Http-Authorization"
    }
  }
  postprocessor "var/jsonpath" {
    mapping = {
      token = "$.data.authToken"
    }
  }

  templater = "text"
}
request "list_req" {
  method = "GET"
  headers = {
    Authorization = "Bearer {{request.auth_req.token}}"
    ContentType   = "application/json"
    Hostname      = "{{hostname}}"
    Useragent     = "Tank"
  }
  tag = "list"
  uri = "/list/?{{filter|query}}"

  postprocessor "var/jsonpath" {
    mapping = {
      items = "$.data.items"
    }
  }

  templater = ""
}
request "order_req" {
  method = "POST"
  headers = {
    Authorization = "Bearer {{request.auth_req.token}}"
    ContentType   = "application/json"
    Hostname      = "{{hostname}}"
    Useragent     = "Tank"
  }
  tag  = "order"
  body = "{}"
  uri  = "/order"

  postprocessor "var/jsonpath" {
    mapping = {
      delivery_id = "$.data.delivery_id"
    }
  }

  templater = ""
}

scenario scenario1 {
  weight           = 50
  min_waiting_time = 1000
  shoot            = [
      "auth(1)",
      "sleep(100)",
      "list(1)",
      "sleep(100)",
      "order(3)"
  ]
}
scenario "scenario2" {
  weight           = 10
  min_waiting_time = 1000
  shoot            = [
	"auth(1)",
	"sleep(100)",
	"list(1)",
	"sleep(100)",
	"order(3)"
  ]
}
`

func Test_decodeHCL(t *testing.T) {

	var config AmmoHCL
	err := hclsimple.Decode("config.hcl", []byte(hclConfig), nil, &config)
	require.NoError(t, err)

	assert.Equal(t, "scenario1", config.Scenarios[0].Name)
	assert.Len(t, config.Scenarios[0].Shoots, 5)
	assert.Equal(t, "scenario2", config.Scenarios[1].Name)
	assert.Len(t, config.Scenarios[1].Shoots, 5)
}
