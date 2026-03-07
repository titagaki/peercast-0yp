package repository

import "testing"

func TestStripYPPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ap:Music", "Music"},       // normal prefix
		{"yp:Rock", "Rock"},         // different prefix
		{"Music", "Music"},          // no prefix
		{"", ""},                    // empty string
		{"yp:Rock:Heavy", "Rock:Heavy"}, // only first colon stripped
		{":", ""},                   // colon only
	}
	for _, tc := range tests {
		got := stripYPPrefix(tc.input)
		if got != tc.want {
			t.Errorf("stripYPPrefix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
