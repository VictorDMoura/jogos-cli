package engine

import (
	"slices"
	"testing"
)

func TestParseKeys(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want []Key
	}{
		{
			name: "ANSI arrow keys",
			raw:  []byte("\x1b[A\x1b[B\x1b[C\x1b[D"),
			want: []Key{KeyUp, KeyDown, KeyRight, KeyLeft},
		},
		{
			name: "WASD keys",
			raw:  []byte("wSaD"),
			want: []Key{KeyUp, KeyDown, KeyLeft, KeyRight},
		},
		{
			name: "Control keys",
			raw:  []byte("\r \x1bqrp"),
			want: []Key{KeyEnter, KeySpace, KeyEsc, KeyQuit, KeyRestart, KeyPause},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseKeys(tc.raw)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("ParseKeys(%q) = %v, esperava %v", tc.raw, got, tc.want)
			}
		})
	}
}
