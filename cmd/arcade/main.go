package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
	"jogos-cli/internal/games/forca"
	"jogos-cli/internal/games/snake"
	"jogos-cli/internal/station"
)

func main() {
	term, err := engine.EnableRawMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar modo bruto do terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx = engine.WithInput(ctx, engine.DefaultInput())

	catalog := []games.Game{
		snake.New(),
		forca.New(),
	}

	launcher := station.NewLauncher(catalog)
	if err := launcher.Run(ctx); err != nil && ctx.Err() == nil {
		term.Restore()
		fmt.Fprintf(os.Stderr, "Erro durante a execução da estação: %v\n", err)
		os.Exit(1)
	}
}
