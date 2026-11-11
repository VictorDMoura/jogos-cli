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
			name: "WASD and alphabet letters",
			raw:  []byte("wSaDforca"),
			want: []Key{KeyW, KeyS, KeyA, KeyD, Key('f'), Key('o'), Key('r'), Key('c'), Key('a')},
		},
		{
			name: "Control keys",
			raw:  []byte("\r \x1bqrp\t\x7f\x03"),
			want: []Key{KeyEnter, KeySpace, KeyEsc, KeyQ, KeyRestart, KeyPause, KeyTab, KeyBackspace, KeyQuit},
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
