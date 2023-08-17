package scenario

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
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
					{
						Name:   "req1",
						Method: "GET",
						URI:    "/api",
						Postprocessors: []PostprocessorHCL{
							{Type: "var/header", Mapping: &(map[string]string{"key": "var/header"})},
							{Type: "var/xpath", Mapping: &(map[string]string{"key": "var/xpath"})},
							{Type: "var/jsonpath", Mapping: &(map[string]string{"key": "var/jsonpath"})},
						},
					},
				},
				Scenarios: []ScenarioHCL{
					{Name: "scenario1", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot1"}},
				},
			},
			want: AmmoConfig{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []VariableSource{
					&VariableSourceJSON{Name: "source1", File: "data.json", fs: fs},
				},
				Requests: []RequestConfig{
					{
						Name:   "req1",
						Method: "GET",
						URI:    "/api",
						Postprocessors: []postprocessor.Postprocessor{
							&postprocessor.VarHeaderPostprocessor{Mapping: map[string]string{"key": "var/header"}},
							&postprocessor.VarXpathPostprocessor{Mapping: map[string]string{"key": "var/xpath"}},
							&postprocessor.VarJsonpathPostprocessor{Mapping: map[string]string{"key": "var/jsonpath"}},
						},
					},
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
						Name: "req1", Method: "GET", URI: "/api",
						Postprocessors: []PostprocessorHCL{
							{Type: "unknown", Mapping: &(map[string]string{"key": "value"})},
						},
					},
				},
			},
			want:    AmmoConfig{},
			wantErr: true,
		},
		{
			name: "MultipleVariableSources",
			ammo: AmmoHCL{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []SourceHCL{
					{Name: "source1", Type: "file/json", File: "data.json"},
					{Name: "source2", Type: "file/csv", File: "data.csv"},
				},
			},
			want: AmmoConfig{
				Variables: map[string]string{"var1": "value1"},
				VariableSources: []VariableSource{
					&VariableSourceJSON{Name: "source1", File: "data.json", fs: fs},
					&VariableSourceCsv{Name: "source2", File: "data.csv", fs: fs},
				},
			},
			wantErr: false,
		},
		{
			name: "MultipleRequests",
			ammo: AmmoHCL{
				Requests: []RequestHCL{
					{Name: "req1", Method: "GET", URI: "/api/1"},
					{Name: "req2", Method: "POST", URI: "/api/2"},
				},
			},
			want: AmmoConfig{
				Requests: []RequestConfig{
					{Name: "req1", Method: "GET", URI: "/api/1"},
					{Name: "req2", Method: "POST", URI: "/api/2"},
				},
			},
			wantErr: false,
		},
		{
			name: "ComplexScenario",
			ammo: AmmoHCL{
				Scenarios: []ScenarioHCL{
					{
						Name:           "scenario1",
						Weight:         2,
						MinWaitingTime: 2000,
						Shoots:         []string{"shoot1", "shoot2"},
					},
					{
						Name:           "scenario2",
						Weight:         1,
						MinWaitingTime: 1000,
						Shoots:         []string{"shoot3"},
					},
				},
			},
			want: AmmoConfig{
				Scenarios: []ScenarioConfig{
					{
						Name:           "scenario1",
						Weight:         2,
						MinWaitingTime: 2000,
						Shoots:         []string{"shoot1", "shoot2"},
					},
					{
						Name:           "scenario2",
						Weight:         1,
						MinWaitingTime: 1000,
						Shoots:         []string{"shoot3"},
					},
				},
			},
			wantErr: false,
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

type unsupportedVariableSource struct{}

func (u unsupportedVariableSource) GetName() string   { return "" }
func (u unsupportedVariableSource) GetVariables() any { return nil }
func (u unsupportedVariableSource) Init() error       { return nil }

type unsupportedPostprocessor struct{}

func (u unsupportedPostprocessor) Process(_ map[string]any, _ *http.Response, _ []byte) error {
	return nil
}

func TestConvertAmmoToHCL(t *testing.T) {
	False := false
	True := true
	tests := []struct {
		name    string
		ammo    AmmoConfig
		want    AmmoHCL
		wantErr bool
	}{
		{
			name: "BasicConversion",
			ammo: AmmoConfig{
				VariableSources: []VariableSource{
					&VariableSourceJSON{Name: "source1", File: "data.json"},
				},
				Requests: []RequestConfig{
					{Name: "req1", Method: "GET", URI: "/api"},
				},
				Scenarios: []ScenarioConfig{
					{Name: "scenario1", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot1"}},
				},
			},
			want: AmmoHCL{
				VariableSources: []SourceHCL{
					{Name: "source1", Type: "file/json", File: "data.json"},
				},
				Requests: []RequestHCL{
					{Name: "req1", Method: "GET", URI: "/api"},
				},
				Scenarios: []ScenarioHCL{
					{Name: "scenario1", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "UnsupportedVariableSourceType",
			ammo: AmmoConfig{
				VariableSources: []VariableSource{
					unsupportedVariableSource{},
				},
			},
			want:    AmmoHCL{},
			wantErr: true,
		},
		{
			name: "UnsupportedPostprocessorType",
			ammo: AmmoConfig{
				Requests: []RequestConfig{
					{
						Name: "req1", Method: "GET", URI: "/api",
						Postprocessors: []postprocessor.Postprocessor{
							unsupportedPostprocessor{},
						},
					},
				},
			},
			want:    AmmoHCL{},
			wantErr: true,
		},
		{
			name: "MultipleVariableSources",
			ammo: AmmoConfig{
				VariableSources: []VariableSource{
					&VariableSourceJSON{Name: "source1", File: "data.json"},
					&VariableSourceCsv{Name: "source2", File: "data.csv"},
				},
			},
			want: AmmoHCL{
				VariableSources: []SourceHCL{
					{Name: "source1", Type: "file/json", File: "data.json"},
					{Name: "source2", Type: "file/csv", File: "data.csv", SkipHeader: &False, HeaderAsFields: &False, Fields: nil},
				},
			},
			wantErr: false,
		},
		{
			name: "MultipleVariableSources2",
			ammo: AmmoConfig{
				VariableSources: []VariableSource{
					&VariableSourceCsv{Name: "source2", File: "data.csv", HeaderAsFields: true, SkipHeader: true, Fields: []string{"field1", "field2"}},
					&VariableSourceCsv{Name: "source2", File: "data.csv", HeaderAsFields: true, SkipHeader: true, Fields: []string{"field3", "field4"}},
					&VariableSourceJSON{Name: "source1", File: "data.json"},
				},
			},
			want: AmmoHCL{
				VariableSources: []SourceHCL{
					{Name: "source2", Type: "file/csv", File: "data.csv", SkipHeader: &True, HeaderAsFields: &True, Fields: &([]string{"field1", "field2"})},
					{Name: "source2", Type: "file/csv", File: "data.csv", SkipHeader: &True, HeaderAsFields: &True, Fields: &([]string{"field3", "field4"})},
					{Name: "source1", Type: "file/json", File: "data.json"},
				},
			},
			wantErr: false,
		},
		{
			name: "MultipleRequests",
			ammo: AmmoConfig{
				Requests: []RequestConfig{
					{Name: "req1", Method: "GET", URI: "/api/1"},
					{Name: "req2", Method: "POST", URI: "/api/2"},
				},
			},
			want: AmmoHCL{
				Requests: []RequestHCL{
					{Name: "req1", Method: "GET", URI: "/api/1"},
					{Name: "req2", Method: "POST", URI: "/api/2"},
				},
			},
			wantErr: false,
		},
		{
			name: "ComplexScenario",
			ammo: AmmoConfig{
				Scenarios: []ScenarioConfig{
					{Name: "scenario1", Weight: 2, MinWaitingTime: 2000, Shoots: []string{"shoot1", "shoot2"}},
					{Name: "scenario2", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot3"}},
				},
			},
			want: AmmoHCL{
				Scenarios: []ScenarioHCL{
					{Name: "scenario1", Weight: 2, MinWaitingTime: 2000, Shoots: []string{"shoot1", "shoot2"}},
					{Name: "scenario2", Weight: 1, MinWaitingTime: 1000, Shoots: []string{"shoot3"}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertAmmoToHCL(tt.ammo)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equalf(t, tt.want, got, "ConvertAmmoToHCL(%v)", tt.ammo)
		})
	}
}
