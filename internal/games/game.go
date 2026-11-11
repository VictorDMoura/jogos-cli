package games

import "context"

// Metadata descreve as informações exibidas no menu da estação.
type Metadata struct {
	ID          string
	Title       string
	Description string
	Controls    string
}

// Game define o contrato que todo jogo da estação deve cumprir.
type Game interface {
	Meta() Metadata
	// Run executa o jogo assumindo o controle do terminal e retorna a pontuação final.
	Run(ctx context.Context) (score int, err error)
}
