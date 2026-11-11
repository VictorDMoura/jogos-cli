package engine

import (
	"io"
	"os"
	"strings"
)

// Códigos ANSI de cores e estilos.
const (
	Reset     = "\x1b[0m"
	Bold      = "\x1b[1m"
	Dim       = "\x1b[2m"
	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	FgBrightGreen  = "\x1b[92m"
	FgBrightYellow = "\x1b[93m"
	FgBrightCyan   = "\x1b[96m"
	FgBrightRed    = "\x1b[91m"
	FgBrightWhite  = "\x1b[97m"

	BgBlue  = "\x1b[44m"
	BgGreen = "\x1b[42m"
	BgRed   = "\x1b[41m"
)

// Screen realiza double-buffering em memória para renderização sem flicker no terminal.
type Screen struct {
	buf strings.Builder
	out io.Writer
}

// NewScreen cria um gerenciador de tela apontando para os.Stdout.
func NewScreen() *Screen {
	return &Screen{
		out: os.Stdout,
	}
}

// ClearScreen limpa toda a tela física do terminal (usado na troca de cenas).
func (s *Screen) ClearScreen() {
	_, _ = io.WriteString(s.out, "\x1b[2J\x1b[H")
}

// BeginFrame inicia a construção de um novo frame posicionando o cursor no topo.
func (s *Screen) BeginFrame() {
	s.buf.Reset()
	s.buf.WriteString("\x1b[H")
}

// WriteLine adiciona uma linha ao buffer e limpa o restante da linha no terminal (\x1b[K).
func (s *Screen) WriteLine(line string) {
	s.buf.WriteString(line)
	s.buf.WriteString("\x1b[K\r\n")
}

// Flush envia todo o frame acumulado em uma única chamada de escrita.
func (s *Screen) Flush() {
	// Limpa linhas residuais abaixo do frame atual
	s.buf.WriteString("\x1b[J")
	_, _ = io.WriteString(s.out, s.buf.String())
}
