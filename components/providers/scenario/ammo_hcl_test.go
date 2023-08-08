package scenario

import (
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex/pandora/core/plugin/pluginconfig"
)

func Test_convertingYamlToHCL(t *testing.T) {
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
}

func ExampleEncodeAmmoHCLVariablesSources() {
	app := AmmoHCL{
		Variables: map[string]string{"host": "localhost"},
		VariableSources: []SourceHCL{
			{
				Type:   "file/csv",
				Name:   "user_srs",
				File:   "users.json",
				Fields: &([]string{"id", "name", "email"}),
			},
			{
				Type:   "file/json",
				Name:   "data_srs",
				File:   "datas.json",
				Fields: &([]string{"id", "name", "email"}),
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
	// variable_source "user_srs" "file/csv" {
	//   file   = "users.json"
	//   fields = ["id", "name", "email"]
	// }
	// variable_source "data_srs" "file/json" {
	//   file   = "datas.json"
	//   fields = ["id", "name", "email"]
	// }
}

func Test_decodeHCL(t *testing.T) {

	fs := afero.NewOsFs()
	file, err := fs.Open("decode_sample_config_test.hcl")
	require.NoError(t, err)
	defer file.Close()

	ammoHCL, err := ParseHCLFile(file)
	require.NoError(t, err)

	assert.Equal(t, "scenario1", ammoHCL.Scenarios[0].Name)
	assert.Len(t, ammoHCL.Scenarios[0].Shoots, 5)
	assert.Equal(t, "scenario2", ammoHCL.Scenarios[1].Name)
	assert.Len(t, ammoHCL.Scenarios[1].Shoots, 5)
}

func TestConvertHCLToAmmo(t *testing.T) {
	fs := afero.NewMemMapFs()
	tests := []struct {
		name    string
		ammo    AmmoHCL
		want    AmmoConfig
		wantErr bool
	}{
		{
			name: "BasicConversion",
			ammo: AmmoHCL{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []SourceHCL{
					{Name: "source1", Type: "file/json", File: "data.json"},
				},
				Requests: []RequestHCL{
					{Name: "req1", Method: "GET", Uri: "/api"},
				},
				Scenarios: []ScenarioHCL{
					{Name: "scenario1", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot1"}},
				},
			},
			want: AmmoConfig{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []VariableSource{
					&VariableSourceJson{Name: "source1", File: "data.json", fs: fs},
				},
				Requests: []RequestConfig{
					{Name: "req1", Method: "GET", Uri: "/api"},
				},
				Scenarios: []ScenarioConfig{
					{Name: "scenario1", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "UnsupportedVariableSourceType",
			ammo: AmmoHCL{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []SourceHCL{
					{Name: "source1", Type: "unknown", File: "data.csv"},
				},
			},
			want:    AmmoConfig{},
			wantErr: true,
		},
		{
			name: "UnsupportedPostprocessorType",
			ammo: AmmoHCL{
				Requests: []RequestHCL{
					{
						Name: "req1", Method: "GET", Uri: "/api",
						Postprocessors: []PostprocessorHCL{
							{Type: "unknown", Mapping: map[string]string{"key": "value"}},
						},
					},
				},
			},
			want:    AmmoConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertHCLToAmmo(tt.ammo, fs)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equalf(t, tt.want, got, "ConvertHCLToAmmo(%v, %v)", tt.ammo, fs)
		})
	}
}
