package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_spreadNames(t *testing.T) {
	tests := []struct {
		name      string
		input     []Scenario
		want      map[string]int
		wantTotal int
	}{
		{
			name:      "",
			input:     []Scenario{{Name: "a", Weight: 20}, {Name: "b", Weight: 30}, {Name: "c", Weight: 60}},
			want:      map[string]int{"a": 2, "b": 3, "c": 6},
			wantTotal: 11,
		},
		{
			name:      "",
			input:     []Scenario{{Name: "a", Weight: 100}, {Name: "b", Weight: 100}, {Name: "c", Weight: 100}},
			want:      map[string]int{"a": 1, "b": 1, "c": 1},
			wantTotal: 3,
		},
		{
			name:      "",
			input:     []Scenario{{Name: "a", Weight: 100}},
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

func Test_extractParams(t *testing.T) {
	body := "Lorem ipsum {{dolor}} sit amet, {{con.sectetur}} adipiscing elit. "
	tests := []struct {
		name string
		req  Request
		want []string
	}{
		{
			name: "",
			req: Request{
				Uri:     "{{dolor}}asdf ljvaosdv {{ alskdfjasfl.asdjfo.['asdfl;]}} laskdfjla\n\n\n{{s v a }}",
				Body:    &body,
				Tag:     "{{tag1}}",
				Headers: map[string]string{"{{dolor1}}": "con.sectetur", "{{con.sectetur1}}": "{{tag2}}"},
			},
			want: []string{
				"dolor",
				"alskdfjasfl.asdjfo.['asdfl;]",
				"s v a",
				"dolor",
				"con.sectetur",
				"dolor1",
				"con.sectetur1",
				"tag2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractParams(tt.req)
			assert.Equalf(t, tt.want, got, "extractParams(%v)", tt.req)
		})
	}
}
