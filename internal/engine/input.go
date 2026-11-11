package engine

import (
	"context"
	"io"
	"os"
	"sync"
)

// Key representa uma tecla mapeada pela engine.
type Key int

const (
	KeyNone Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeySpace
	KeyQuit
	KeyRestart
	KeyPause
)

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
	buf := make([]byte, 16)
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
	for i := 0; i < len(b); i++ {
		if b[i] == 0x1b { // ESC ou sequência ANSI
			if i+2 < len(b) && (b[i+1] == '[' || b[i+1] == 'O') {
				switch b[i+2] {
				case 'A':
					keys = append(keys, KeyUp)
				case 'B':
					keys = append(keys, KeyDown)
				case 'C':
					keys = append(keys, KeyRight)
				case 'D':
					keys = append(keys, KeyLeft)
				}
				i += 2
				continue
			}
			keys = append(keys, KeyEsc)
			continue
		}

		switch b[i] {
		case 'w', 'W':
			keys = append(keys, KeyUp)
		case 's', 'S':
			keys = append(keys, KeyDown)
		case 'a', 'A':
			keys = append(keys, KeyLeft)
		case 'd', 'D':
			keys = append(keys, KeyRight)
		case '\r', '\n':
			keys = append(keys, KeyEnter)
		case ' ':
			keys = append(keys, KeySpace)
		case 'q', 'Q', 0x03: // q, Q ou Ctrl+C
			keys = append(keys, KeyQuit)
		case 'r', 'R':
			keys = append(keys, KeyRestart)
		case 'p', 'P':
			keys = append(keys, KeyPause)
		}
	}
	return keys
}
