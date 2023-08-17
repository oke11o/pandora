package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreprocessor_Process(t *testing.T) {
	tests := []struct {
		name       string
		prep       Preprocessor
		templVars  map[string]any
		sourceVars map[string]any
		wantMap    map[string]any
		wantErr    bool
	}{
		{
			name: "Nil templateVars",
			prep: Preprocessor{
				Variables: map[string]string{
					"var1": "source.items[0].id",
					"var2": "source.items[1]",
				},
			},
			sourceVars: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
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
			templVars: map[string]any{
				"request": map[string]any{
					"auth": map[string]any{"token": "Bearer token"},
				},
			},
			sourceVars: map[string]any{
				"source": map[string]any{
					"items": []map[string]any{
						{"id": "1"},
						{"id": "2"},
					},
				},
			},
			wantMap: map[string]any{
				"request": map[string]any{
					"auth": map[string]any{"token": "Bearer token"},
				},
				"preprocessor": map[string]any{
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
			err := tt.prep.Process(tt.templVars, tt.sourceVars)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMap, tt.templVars)
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

func Test_getValue_iterators(t *testing.T) {
	p := &Preprocessor{iterator: newNextIterator(0)}

	reqMap := map[string]any{
		"source": map[string]any{
			"items": []any{11, 22, 33},
			"list":  []string{"11", "22", "33"},
		},
	}
	var got any

	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "11", got)
	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "22", got)
	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "33", got)
	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "11", got)
	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "22", got)
	got, _ = p.getValue(reqMap, "source.list[next]")
	assert.Equal(t, "33", got)

	got, _ = p.getValue(reqMap, "source.list[last]")
	assert.Equal(t, "33", got)
	got, _ = p.getValue(reqMap, "source.list[last]")
	assert.Equal(t, "33", got)
	got, _ = p.getValue(reqMap, "source.list[-2]")
	assert.Equal(t, "22", got)

	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 11, got)
	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 11, got)
	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 22, got)
	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 22, got)
	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 33, got)
	got, _ = p.getValue(reqMap, "source.items[rand]")
	assert.Equal(t, 22, got)

	got, _ = p.getValue(reqMap, "source.items[next]")
	assert.Equal(t, 11, got)
	got, _ = p.getValue(reqMap, "source.items[next]")
	assert.Equal(t, 22, got)
	got, _ = p.getValue(reqMap, "source.items[next]")
	assert.Equal(t, 33, got)
	got, _ = p.getValue(reqMap, "source.items[next]")
	assert.Equal(t, 11, got)
	got, _ = p.getValue(reqMap, "source.items[next]")
	assert.Equal(t, 22, got)

}
