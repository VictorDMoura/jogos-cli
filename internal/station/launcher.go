package station

import (
	"context"
	"fmt"
	"strings"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

// Category agrupa jogos por tema ou assunto no menu da estação.
type Category struct {
	ID          string
	Title       string
	Description string
	Games       []games.Game
}

// GameStats guarda a pontuação do jogador na sessão atual (em memória).
type GameStats struct {
	LastScore   int
	BestScore   int
	TimesPlayed int
}

// Launcher gerencia o menu principal de categorias e os submenus de jogos da estação.
type Launcher struct {
	categories      []Category
	inSubmenu       bool
	selectedCatIdx  int
	selectedGameIdx int
	stats           map[string]*GameStats
	totalScore      int
}

// NewLauncher cria um menu padrão com uma categoria única caso uma lista simples seja passada.
func NewLauncher(catalog []games.Game) *Launcher {
	return NewCategorizedLauncher([]Category{
		{
			ID:          "geral",
			Title:       "🕹️ Catálogo Geral de Jogos",
			Description: "Todos os jogos disponíveis na estação.",
			Games:       catalog,
		},
	})
}

// NewCategorizedLauncher cria o menu da estação organizado por categorias e submenus.
func NewCategorizedLauncher(categories []Category) *Launcher {
	stats := make(map[string]*GameStats)
	for _, cat := range categories {
		for _, g := range cat.Games {
			stats[g.Meta().ID] = &GameStats{}
		}
	}
	return &Launcher{
		categories: categories,
		stats:      stats,
	}
}

// Run executa o loop principal do menu de categorias e submenus da estação.
func (l *Launcher) Run(ctx context.Context) error {
	screen := engine.NewScreen()
	input := engine.InputFromContext(ctx)
	input.Drain()

	screen.ClearScreen()
	l.render(screen)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case key := <-input.Events():
			if !l.inSubmenu {
				if exit := l.handleMainMenuKey(ctx, screen, input, key); exit {
					return nil
				}
			} else {
				if err := l.handleSubmenuKey(ctx, screen, input, key); err != nil {
					return err
				}
			}
		}
	}
}

func (l *Launcher) handleMainMenuKey(ctx context.Context, screen *engine.Screen, input *engine.InputReader, key engine.Key) bool {
	// Total de opções = categorias + opção "Sair da Estação"
	totalOptions := len(l.categories) + 1

	switch key {
	case engine.KeyUp, engine.KeyW:
		l.selectedCatIdx = (l.selectedCatIdx - 1 + totalOptions) % totalOptions
		l.render(screen)

	case engine.KeyDown, engine.KeyS:
		l.selectedCatIdx = (l.selectedCatIdx + 1) % totalOptions
		l.render(screen)

	case engine.KeyEnter, engine.KeySpace, engine.KeyRight, engine.KeyD:
		if l.selectedCatIdx == len(l.categories) {
			return true
		}
		l.inSubmenu = true
		l.selectedGameIdx = 0
		l.render(screen)

	case engine.KeyEsc, engine.KeyQuit, engine.KeyQ:
		return true
	}
	return false
}

func (l *Launcher) handleSubmenuKey(ctx context.Context, screen *engine.Screen, input *engine.InputReader, key engine.Key) error {
	currentCat := l.categories[l.selectedCatIdx]
	// Total de opções no submenu = jogos da categoria + opção "Voltar às Categorias"
	totalOptions := len(currentCat.Games) + 1

	switch key {
	case engine.KeyUp, engine.KeyW:
		l.selectedGameIdx = (l.selectedGameIdx - 1 + totalOptions) % totalOptions
		l.render(screen)

	case engine.KeyDown, engine.KeyS:
		l.selectedGameIdx = (l.selectedGameIdx + 1) % totalOptions
		l.render(screen)

	case engine.KeyEnter, engine.KeySpace:
		// Se selecionou "Voltar às Categorias"
		if l.selectedGameIdx == len(currentCat.Games) {
			l.inSubmenu = false
			l.render(screen)
			return nil
		}

		selectedGame := currentCat.Games[l.selectedGameIdx]
		score, err := selectedGame.Run(ctx)
		if err != nil && ctx.Err() != nil {
			return err
		}

		// Atualiza a pontuação do jogador na sessão
		meta := selectedGame.Meta()
		st := l.stats[meta.ID]
		if st == nil {
			st = &GameStats{}
			l.stats[meta.ID] = st
		}
		st.LastScore = score
		st.TimesPlayed++
		if score > st.BestScore {
			st.BestScore = score
		}
		l.totalScore += score

		input.Drain()
		screen.ClearScreen()
		l.render(screen)

	case engine.KeyEsc, engine.KeyBackspace, engine.KeyLeft, engine.KeyA, engine.KeyQ, engine.KeyQuit:
		if key == engine.KeyQuit {
			return context.Canceled
		}
		l.inSubmenu = false
		l.render(screen)
	}
	return nil
}

func (l *Launcher) categoryBestAndTotal(cat Category) (best int, played int) {
	for _, g := range cat.Games {
		if st := l.stats[g.Meta().ID]; st != nil {
			if st.BestScore > best {
				best = st.BestScore
			}
			played += st.TimesPlayed
		}
	}
	return best, played
}

func (l *Launcher) render(screen *engine.Screen) {
	screen.BeginFrame()

	boxWidth := 72
	border := strings.Repeat("═", boxWidth)

	screen.WriteLine("")
	screen.WriteLine("  " + engine.FgBrightCyan + "╔" + border + "╗" + engine.Reset)
	screen.WriteLine(fmt.Sprintf("  %s║%s%s%s                 🕹️  JOGOS-CLI — ESTAÇÃO ARCADE & ENSINO               %s%s║%s",
		engine.FgBrightCyan, engine.Reset, engine.Bold, engine.FgBrightYellow, engine.Reset, engine.FgBrightCyan, engine.Reset))
	screen.WriteLine("  " + engine.FgBrightCyan + "╚" + border + "╝" + engine.Reset)
	screen.WriteLine("")

	if !l.inSubmenu {
		l.renderMainMenu(screen, boxWidth)
	} else {
		l.renderSubmenu(screen, boxWidth)
	}

	screen.Flush()
}

func (l *Launcher) renderMainMenu(screen *engine.Screen, boxWidth int) {
	screen.WriteLine(fmt.Sprintf("  %s📊 RESUMO DA SESSÃO:%s   Pontuação Total Acumulada: %s%04d pts%s",
		engine.Bold, engine.Reset,
		engine.FgBrightCyan, l.totalScore, engine.Reset,
	))
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("  %s📂 ESCOLHA UMA CATEGORIA:%s", engine.Bold, engine.Reset))
	screen.WriteLine("")

	for i, cat := range l.categories {
		cursor := "   "
		titleStyle := engine.FgWhite
		if i == l.selectedCatIdx {
			cursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightGreen)
			titleStyle = engine.Bold + engine.FgBrightGreen
		}

		bestCat, playedCat := l.categoryBestAndTotal(cat)
		badge := fmt.Sprintf("%s(%d jogos | Melhor: %04d pts | Partidas: %d)%s",
			engine.Dim, len(cat.Games), bestCat, playedCat, engine.Reset)

		screen.WriteLine(fmt.Sprintf("%s%s%s  %s", cursor, titleStyle, cat.Title, badge))
	}

	// Opção Sair
	exitCursor := "   "
	exitStyle := engine.FgWhite
	if l.selectedCatIdx == len(l.categories) {
		exitCursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightRed)
		exitStyle = engine.Bold + engine.FgBrightRed
	}
	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("%s%s🚪 Sair da Estação%s", exitCursor, exitStyle, engine.Reset))
	screen.WriteLine("")
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)

	// Descrição da categoria selecionada
	if l.selectedCatIdx < len(l.categories) {
		cat := l.categories[l.selectedCatIdx]
		screen.WriteLine(fmt.Sprintf("  %s%sSobre a Categoria:%s %s", engine.Bold, engine.FgBrightCyan, engine.Reset, cat.Description))
		gameNames := make([]string, 0, len(cat.Games))
		for _, g := range cat.Games {
			gameNames = append(gameNames, g.Meta().Title)
		}
		screen.WriteLine(fmt.Sprintf("  %s%sJogos Inclusos:%s    %s", engine.Bold, engine.FgBrightYellow, engine.Reset, strings.Join(gameNames, "  •  ")))
	} else {
		screen.WriteLine(fmt.Sprintf("  %sEncerra a estação de jogos e retorna ao terminal.%s", engine.Dim, engine.Reset))
		screen.WriteLine("")
	}

	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine(fmt.Sprintf("  %sNavegação: [↑/↓ ou W/S] Mover  •  [ENTER/ESPAÇO] Abrir Categoria  •  [ESC/Q] Sair%s", engine.Dim, engine.Reset))
}

func (l *Launcher) renderSubmenu(screen *engine.Screen, boxWidth int) {
	cat := l.categories[l.selectedCatIdx]

	bestGame := 0
	lastGame := 0
	if l.selectedGameIdx < len(cat.Games) {
		id := cat.Games[l.selectedGameIdx].Meta().ID
		if st := l.stats[id]; st != nil {
			bestGame = st.BestScore
			lastGame = st.LastScore
		}
	}

	screen.WriteLine(fmt.Sprintf("  %s📊 SUA PONTUAÇÃO:%s   Última: %s%04d pts%s  │  Melhor no Jogo: %s%04d pts%s  │  Total Sessão: %s%04d pts%s",
		engine.Bold, engine.Reset,
		engine.FgBrightYellow, lastGame, engine.Reset,
		engine.FgBrightGreen, bestGame, engine.Reset,
		engine.FgBrightCyan, l.totalScore, engine.Reset,
	))
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("  %s%s%s %s(Submenu — Selecione um Jogo):%s",
		engine.Bold, engine.FgBrightCyan, cat.Title, engine.Dim, engine.Reset))
	screen.WriteLine("")

	for i, g := range cat.Games {
		meta := g.Meta()
		st := l.stats[meta.ID]
		cursor := "   "
		titleStyle := engine.FgWhite
		if i == l.selectedGameIdx {
			cursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightGreen)
			titleStyle = engine.Bold + engine.FgBrightGreen
		}

		scoreBadge := fmt.Sprintf("%s(Melhor: %04d pts | Última: %04d pts)%s",
			engine.Dim, st.BestScore, st.LastScore, engine.Reset)

		screen.WriteLine(fmt.Sprintf("%s%s%s  %s", cursor, titleStyle, meta.Title, scoreBadge))
	}

	// Opção Voltar ao Menu de Categorias
	backCursor := "   "
	backStyle := engine.FgWhite
	if l.selectedGameIdx == len(cat.Games) {
		backCursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightYellow)
		backStyle = engine.Bold + engine.FgBrightYellow
	}
	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("%s%s⬅️  Voltar às Categorias%s", backCursor, backStyle, engine.Reset))
	screen.WriteLine("")
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)

	if l.selectedGameIdx < len(cat.Games) {
		meta := cat.Games[l.selectedGameIdx].Meta()
		screen.WriteLine(fmt.Sprintf("  %s%sSobre o Jogo:%s %s", engine.Bold, engine.FgBrightCyan, engine.Reset, meta.Description))
		screen.WriteLine(fmt.Sprintf("  %s%sControles:%s    %s", engine.Bold, engine.FgBrightYellow, engine.Reset, meta.Controls))
	} else {
		screen.WriteLine(fmt.Sprintf("  %sRetorna para a tela inicial de seleção de categorias (Arcade / Assuntos).%s", engine.Dim, engine.Reset))
		screen.WriteLine("")
	}

	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine(fmt.Sprintf("  %sNavegação: [↑/↓ ou W/S] Selecionar  •  [ENTER/ESPAÇO] Jogar  •  [ESC/Backspace] Voltar%s", engine.Dim, engine.Reset))
}
