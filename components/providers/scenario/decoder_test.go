package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_gcd(t *testing.T) {
	tests := []struct {
		name string
		a    int64
		b    int64
		want int64
	}{
		{
			name: "",
			a:    40,
			b:    60,
			want: 20,
		},
		{
			name: "",
			a:    2,
			b:    3,
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, gcd(tt.a, tt.b), "gcd(%v, %v)", tt.a, tt.b)
		})
	}
}

func Test_lcm(t *testing.T) {
	tests := []struct {
		name string
		a    int64
		b    int64
		want int64
	}{
		{
			name: "",
			a:    40,
			b:    60,
			want: 120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, lcm(tt.a, tt.b), "lcm(%v, %v)", tt.a, tt.b)
		})
	}
}

func Test_lcmm(t *testing.T) {
	tests := []struct {
		name string
		a    []int64
		want int64
	}{
		{
			name: "",
			a:    []int64{3, 4, 6},
			want: 12,
		},
		{
			name: "",
			a:    []int64{3, 4, 5, 6, 7}, // 140,105,84,70,60
			want: 420,
		},
		{
			name: "",
			a:    []int64{2, 4, 5, 10},
			want: 20,
		},
		{
			name: "",
			a:    []int64{20, 20, 20, 20},
			want: 20,
		},
		{
			name: "",
			a:    []int64{40, 50, 70},
			want: 1400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, lcmm(tt.a...), "lcmm(%v)", tt.a)
		})
	}
}

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

func Test_gcdm(t *testing.T) {
	tests := []struct {
		name    string
		weights []int64
		want    int64
	}{
		{
			name:    "",
			weights: []int64{20, 30, 60},
			want:    10,
		},
		{
			name:    "",
			weights: []int64{6, 6, 6},
			want:    6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, gcdm(tt.weights...), "gcdm(%v)", tt.weights)
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
