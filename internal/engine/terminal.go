package engine

import (
	"fmt"
	"os"
)

// Terminal gerencia o modo bruto (Raw Mode) e a tela alternativa do console.
type Terminal struct {
	state *platformState
}

// EnableRawMode coloca o terminal em modo bruto, ativa suporte ANSI,
// entra na tela alternativa e oculta o cursor.
func EnableRawMode() (*Terminal, error) {
	st, err := enablePlatformRawMode()
	if err != nil {
		return nil, err
	}
	// Entra no Alternate Screen Buffer (\x1b[?1049h), limpa a tela (\x1b[2J\x1b[H) e oculta o cursor (\x1b[?25l)
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[2J\x1b[H\x1b[?25l")
	return &Terminal{state: st}, nil
}

// Restore restaura o cursor, sai da tela alternativa e devolve o terminal ao estado original.
func (t *Terminal) Restore() {
	// Mostra o cursor (\x1b[?25h), reseta cores (\x1b[0m) e sai do Alternate Screen Buffer (\x1b[?1049l)
	fmt.Fprint(os.Stdout, "\x1b[0m\x1b[?25h\x1b[?1049l")
	if t != nil && t.state != nil {
		restorePlatformRawMode(t.state)
		t.state = nil
	}
}
