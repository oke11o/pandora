package httpscenario

import (
	"testing"

	httpscenario "github.com/yandex/pandora/components/guns/http_scenario"

	"github.com/stretchr/testify/assert"
)

func TestPreprocessor_Process(t *testing.T) {
	tests := []struct {
		name      string
		prep      Preprocessor
		templVars map[string]httpscenario.GetSetter
		wantMap   httpscenario.Getter
		wantErr   bool
	}{
		{
			name: "Nil templateVars",
			prep: Preprocessor{
				Variables: map[string]string{
					"var1": "source.items[0].id",
					"var2": "source.items[1]",
				},
			},
			wantErr: true,
		},
		{
			name: "Simple Processing",
			prep: Preprocessor{
				Variables: map[string]string{
					"var1": "source.items[0].id",
					"var2": "source.items[1]",
					"var3": "request.auth.token",
				},
			},
			templVars: map[string]httpscenario.GetSetter{
				"request": &gunGetSetter{
					m: map[string]any{
						"auth": map[string]any{"token": "Bearer token"},
					},
				},
				"source": &gunGetSetter{
					m: map[string]any{
						"items": []map[string]any{
							{"id": "1"},
							{"id": "2"},
						},
					},
				},
			},
			wantMap: gunGetSetter{
				m: map[string]any{
					"var1": "1",
					"var2": map[string]any{"id": "2"},
					"var3": "Bearer token",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get, err := tt.prep.Process(tt.templVars)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMap, get)
			}
		})
	}
}
