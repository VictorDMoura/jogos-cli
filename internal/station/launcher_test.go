package station

import (
	"context"
	"io"
	"testing"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

type mockGame struct {
	id    string
	score int
}

func (m *mockGame) Meta() games.Metadata {
	return games.Metadata{
		ID:          m.id,
		Title:       "Mock " + m.id,
		Description: "Jogo de teste",
		Controls:    "Enter",
	}
}

func (m *mockGame) Run(ctx context.Context) (int, error) {
	return m.score, nil
}

func TestCategorizedLauncherSubmenuAndScore(t *testing.T) {
	gArcade := &mockGame{id: "snake", score: 120}
	gPortugues := &mockGame{id: "ortografia", score: 85}

	categories := []Category{
		{
			ID:    "arcade",
			Title: "Jogos Arcade",
			Games: []games.Game{gArcade},
		},
		{
			ID:    "portugues",
			Title: "Português",
			Games: []games.Game{gPortugues},
		},
	}

	ctx := context.Background()
	screen := engine.NewScreen()
	input := engine.NewInputReader(io.LimitReader(nil, 0))
	launcher := NewCategorizedLauncher(categories)

	// 1. Seta para baixo no menu principal -> seleciona categoria 1 ("Português")
	if exit := launcher.handleMainMenuKey(ctx, screen, input, engine.KeyDown); exit {
		t.Fatal("não deveria sair ao navegar para baixo")
	}
	if launcher.selectedCatIdx != 1 {
		t.Fatalf("esperava selectedCatIdx = 1, obteve %d", launcher.selectedCatIdx)
	}

	// 2. Enter -> entra no submenu da categoria "Português"
	if exit := launcher.handleMainMenuKey(ctx, screen, input, engine.KeyEnter); exit {
		t.Fatal("não deveria sair ao entrar em uma categoria")
	}
	if !launcher.inSubmenu {
		t.Fatal("esperava estar dentro do submenu (inSubmenu = true)")
	}

	// 3. Enter dentro do submenu -> executa o jogo "ortografia" (retorna 85 pontos)
	if err := launcher.handleSubmenuKey(ctx, screen, input, engine.KeyEnter); err != nil {
		t.Fatalf("erro inesperado no submenu: %v", err)
	}
	if launcher.totalScore != 85 {
		t.Fatalf("esperava totalScore = 85 após jogar ortografia, obteve %d", launcher.totalScore)
	}
	if launcher.stats["ortografia"].BestScore != 85 {
		t.Fatalf("esperava BestScore = 85 em ortografia, obteve %d", launcher.stats["ortografia"].BestScore)
	}

	// 4. ESC dentro do submenu -> volta para o menu principal de categorias
	if err := launcher.handleSubmenuKey(ctx, screen, input, engine.KeyEsc); err != nil {
		t.Fatalf("erro ao voltar do submenu: %v", err)
	}
	if launcher.inSubmenu {
		t.Fatal("esperava ter voltado ao menu principal de categorias (inSubmenu = false)")
	}

	// 5. ESC no menu principal -> encerra a estação
	if exit := launcher.handleMainMenuKey(ctx, screen, input, engine.KeyEsc); !exit {
		t.Fatal("esperava encerrar a estação ao pressionar ESC no menu principal")
	}
}
