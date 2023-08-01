package scenario

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/yandex/pandora/core/config"
	"github.com/yandex/pandora/core/plugin/pluginconfig"
)

func Test_decode_parseVariableSourceJson(t *testing.T) {
	const exampleVariableSourceJson = `
src:
  type: "file/json"
  name: "json_src"
  file: "_files/users.json"
`

	Import(nil)
	testOnce.Do(func() {
		pluginconfig.AddHooks()
	})

	data := make(map[string]any)
	err := yaml.Unmarshal([]byte(exampleVariableSourceJson), &data)
	require.NoError(t, err)

	out := struct {
		Src VariableSource `yaml:"src"`
	}{}

	err = config.DecodeAndValidate(data, &out)
	require.NoError(t, err)

	vs, ok := out.Src.(*VariableSourceJson)
	require.True(t, ok)
	require.Equal(t, "json_src", vs.GetName())
}
