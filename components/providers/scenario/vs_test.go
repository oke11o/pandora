package scenario

import (
	"testing"

	"github.com/yandex/pandora/core/plugin/pluginconfig"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"

	"github.com/yandex/pandora/core/config"
)

const exampleVariableSourceYAML = `
src:
  type: "file/csv"
  name: "users_src"
  file: "_files/users.csv"
  mapping:
    users: [ "user_id", "name" ]
`

func Test_decoder_parseVariableSource(t *testing.T) {
	Import(nil)
	testOnce.Do(func() {
		pluginconfig.AddHooks()
	})

	data := make(map[string]any)
	err := yaml.Unmarshal([]byte(exampleVariableSourceYAML), &data)
	require.NoError(t, err)

	out := struct {
		Src VariableSource `yaml:"src"`
	}{}

	err = config.DecodeAndValidate(data, &out)
	require.NoError(t, err)

	require.Equal(t, "users_src", out.Src.GetName())
	require.Equal(t, map[string]any{"users": []any{"user_id", "name"}}, out.Src.GetMapping())

}
