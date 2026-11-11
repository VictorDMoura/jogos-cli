package engine

import (
	"context"
	"io"
	"os"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Key representa uma tecla mapeada pela engine.
// Letras ('a'-'z') e dígitos ('0'-'9') são representados pelo seu próprio rune em minúsculo.
// Teclas especiais de controle/navegação começam acima da faixa ASCII (256+).
type Key rune

const (
	KeyNone Key = 0

	KeyUp Key = iota + 256
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeySpace
	KeyBackspace
	KeyTab
	KeyQuit // Ctrl+C
)

// Atalhos semânticos para letras frequentemente usadas como comandos em jogos arcade.
const (
	KeyW       Key = 'w'
	KeyA       Key = 'a'
	KeyS       Key = 's'
	KeyD       Key = 'd'
	KeyP       Key = 'p'
	KeyR       Key = 'r'
	KeyQ       Key = 'q'
	KeyRestart Key = 'r'
	KeyPause   Key = 'p'
)

// IsLetter informa se a tecla corresponde a uma letra de A a Z.
func (k Key) IsLetter() bool {
	return k >= 'a' && k <= 'z'
}

// UpperRune retorna a letra em maiúsculo ('A'-'Z') caso IsLetter() seja verdadeiro.
func (k Key) UpperRune() rune {
	if k.IsLetter() {
		return rune(k - 'a' + 'A')
	}
	return 0
}

type inputContextKey struct{}

// InputReader lê eventos de teclado de forma assíncrona a partir de um io.Reader.
type InputReader struct {
	ch   chan Key
	once sync.Once
	r    io.Reader
}

var (
	defaultReader     *InputReader
	defaultReaderOnce sync.Once
)

// DefaultInput retorna o leitor de entrada padrão ligado ao os.Stdin.
func DefaultInput() *InputReader {
	defaultReaderOnce.Do(func() {
		defaultReader = NewInputReader(os.Stdin)
		defaultReader.Start()
	})
	return defaultReader
}

// NewInputReader cria um novo leitor de teclado a partir de um io.Reader.
func NewInputReader(r io.Reader) *InputReader {
	return &InputReader{
		ch: make(chan Key, 32),
		r:  r,
	}
}

// WithInput anexa um InputReader ao contexto.
func WithInput(ctx context.Context, ir *InputReader) context.Context {
	return context.WithValue(ctx, inputContextKey{}, ir)
}

// InputFromContext obtém o InputReader do contexto ou retorna o DefaultInput.
func InputFromContext(ctx context.Context) *InputReader {
	if ir, ok := ctx.Value(inputContextKey{}).(*InputReader); ok && ir != nil {
		return ir
	}
	return DefaultInput()
}

// Start inicia a goroutine de leitura contínua de teclas.
func (ir *InputReader) Start() {
	ir.once.Do(func() {
		go ir.readLoop()
	})
}

// Events retorna o canal de teclas lidas.
func (ir *InputReader) Events() <-chan Key {
	return ir.ch
}

// Drain descarta teclas acumuladas no buffer antes de iniciar uma nova tela/partida.
func (ir *InputReader) Drain() {
	for {
		select {
		case <-ir.ch:
		default:
			return
		}
	}
}

func (ir *InputReader) readLoop() {
	buf := make([]byte, 32)
	for {
		n, err := ir.r.Read(buf)
		if err != nil || n == 0 {
			return
		}
		for _, k := range ParseKeys(buf[:n]) {
			if k != KeyNone {
				select {
				case ir.ch <- k:
				default:
				}
			}
		}
	}
}

// ParseKeys converte bytes brutos lidos do terminal em eventos Key.
func ParseKeys(b []byte) []Key {
	var keys []Key
	for i := 0; i < len(b); {
		if b[i] == 0x1b { // ESC ou sequência ANSI
			if i+2 < len(b) && (b[i+1] == '[' || b[i+1] == 'O') {
				switch b[i+2] {
				case 'A':
					keys = append(keys, KeyUp)
					i += 3
					continue
				case 'B':
					keys = append(keys, KeyDown)
					i += 3
					continue
				case 'C':
					keys = append(keys, KeyRight)
					i += 3
					continue
				case 'D':
					keys = append(keys, KeyLeft)
					i += 3
					continue
				}
			}
			keys = append(keys, KeyEsc)
			i++
			continue
		}

		switch b[i] {
		case '\r', '\n':
			keys = append(keys, KeyEnter)
			i++
			continue
		case ' ':
			keys = append(keys, KeySpace)
			i++
			continue
		case '\t':
			keys = append(keys, KeyTab)
			i++
			continue
		case 0x08, 0x7f: // Backspace / Delete
			keys = append(keys, KeyBackspace)
			i++
			continue
		case 0x03: // Ctrl+C
			keys = append(keys, KeyQuit)
			i++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		i += size
		if r == utf8.RuneError {
			continue
		}

		if norm := normalizeBaseLetter(r); norm != 0 {
			keys = append(keys, Key(norm))
			continue
		}

		if r >= '0' && r <= '9' {
			keys = append(keys, Key(r))
		}
	}
	return keys
}

func normalizeBaseLetter(r rune) rune {
	r = unicode.ToLower(r)
	if r >= 'a' && r <= 'z' {
		return r
	}
	switch r {
	case 'á', 'à', 'â', 'ã', 'ä':
		return 'a'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'ó', 'ò', 'ô', 'õ', 'ö':
		return 'o'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'ç':
		return 'c'
	}
	return 0
}
