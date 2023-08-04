package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreprocessor_Process(t *testing.T) {
	tests := []struct {
		name      string
		prep      Preprocessor
		reqMap    map[string]any
		wantMap   map[string]any
		wantErr   bool
		errSubstr string
	}{
		{
			name: "Simple Processing",
			prep: Preprocessor{
				Variables: map[string]string{
					"var1": "source.items[0].id",
					"var2": "source.items[1]",
				},
			},
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			wantMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
				"preprocessor": map[string]any{
					"var1": "1",
					"var2": map[string]any{"id": "2"},
				},
			},
			wantErr:   false,
			errSubstr: "",
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prep.Process(tt.reqMap)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMap, tt.reqMap)
			}
		})
	}
}

func TestPreprocessor_getValue(t *testing.T) {
	tests := []struct {
		name    string
		reqMap  map[string]any
		v       string
		want    any
		wantErr bool
	}{
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			v:       "source.items[0].id",
			want:    "1",
			wantErr: false,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			v:       "source.items[1]",
			want:    map[string]any{"id": "2"},
			wantErr: false,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			v:       "source.items[1].title",
			want:    nil,
			wantErr: true,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			v: "source.items",
			want: []map[string]any{
				{"id": "1"},
				{"id": "2"},
			},
			wantErr: false,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []map[string]string{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			v:       "source.items[0].id",
			want:    "1",
			wantErr: false,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []any{11, 22, 33},
				},
			},
			v:       "source.items[0]",
			want:    11,
			wantErr: false,
		},
		{
			name: "",
			reqMap: map[string]any{
				"source": map[string]any{
					"items": []any{11, 22, 33},
				},
			},
			v:       "source.items[0].id",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Preprocessor{}
			got, err := p.getValue(tt.reqMap, tt.v)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equalf(t, tt.want, got, "getValue(%v, %v)", tt.reqMap, tt.v)
		})
	}
}
