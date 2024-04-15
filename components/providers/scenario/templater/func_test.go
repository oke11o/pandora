package templater

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandInt(t *testing.T) {
	tests := []struct {
		name      string
		args      []any
		want      int
		wantDelta float64
	}{
		{
			name:      "No args",
			args:      nil,
			want:      15,
			wantDelta: 15,
		},
		{
			name:      "Two args",
			args:      []any{10, 20},
			want:      15,
			wantDelta: 5,
		},
		{
			name:      "First string arg can be converted, second is invalid",
			args:      []any{"26", "invalid"},
			want:      13,
			wantDelta: 13,
		},
		{
			name:      "Two string args can be converted",
			args:      []any{"200", "300"},
			want:      250,
			wantDelta: 50,
		},
		{
			name:      "Two args, second invalid",
			args:      []any{20, "invalid"},
			want:      10,
			wantDelta: 10,
		},
		{
			name:      "More than two args",
			args:      []any{100, 200, 30},
			want:      15,
			wantDelta: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := RandInt(tt.args...)
			g, err := strconv.Atoi(get)
			require.NoError(t, err)
			require.InDelta(t, tt.want, g, tt.wantDelta)
		})
	}
}

func TestRandString(t *testing.T) {
	tests := []struct {
		name       string
		args       []any
		wantLength int
	}{
		{
			name:       "No args, default length",
			args:       nil,
			wantLength: 1,
		},
		{
			name:       "Specific length",
			args:       []any{5},
			wantLength: 5,
		},
		{
			name:       "Specific length and characters",
			args:       []any{10, "abc"},
			wantLength: 10,
		},
		{
			name:       "Invalid length argument",
			args:       []any{"invalid"},
			wantLength: 1,
		},
		{
			name:       "Invalid length, valid characters",
			args:       []any{"invalid", "def"},
			wantLength: 1,
		},
		{
			name:       "More than two args",
			args:       []any{5, "gh", "extra"},
			wantLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandString(tt.args...)
			require.Len(t, got, tt.wantLength)
		})
	}
}

func TestRandStringLetters(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		letters string
	}{
		{
			name:    "Simple",
			length:  10,
			letters: "ab",
		},
		{
			name:    "Simple",
			length:  100,
			letters: "absdfave",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandString(tt.length, tt.letters)
			require.Len(t, got, tt.length)

			l := map[rune]int{}
			for _, r := range got {
				l[r]++
			}
			gotCount := 0
			for _, c := range l {
				gotCount += c
			}
			require.Equal(t, tt.length, gotCount)
		})
	}
}

func TestParseFunc(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		wantF    any
		wantArgs []string
	}{
		{
			name:     "Simple",
			arg:      "randInt(10, 20)",
			wantF:    RandInt,
			wantArgs: []string{"10", "20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotF, gotArgs := ParseFunc(tt.arg)
			f := gotF.(func(args ...any) string)
			a := []any{}
			for _, arg := range gotArgs {
				a = append(a, arg)
			}
			f(a...)
			require.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
