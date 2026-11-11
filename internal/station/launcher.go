package station

import (
	"context"
	"fmt"
	"strings"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

// GameStats guarda a pontuação do jogador na sessão atual (em memória).
type GameStats struct {
	LastScore  int
	BestScore  int
	TimesPlayed int
}

// Launcher gerencia o menu principal da Estação de Jogos.
type Launcher struct {
	catalog     []games.Game
	selectedIdx int
	stats       map[string]*GameStats
	totalScore  int
}

// NewLauncher cria uma nova instância do menu principal com os jogos disponíveis.
func NewLauncher(catalog []games.Game) *Launcher {
	stats := make(map[string]*GameStats, len(catalog))
	for _, g := range catalog {
		stats[g.Meta().ID] = &GameStats{}
	}
	return &Launcher{
		catalog: catalog,
		stats:   stats,
	}
}

// Run executa o loop principal do menu da estação.
func (l *Launcher) Run(ctx context.Context) error {
	screen := engine.NewScreen()
	input := engine.InputFromContext(ctx)
	input.Drain()

	// Total de opções = jogos do catálogo + opção "Sair da Estação"
	totalOptions := len(l.catalog) + 1

	screen.ClearScreen()
	l.render(screen)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case key := <-input.Events():
			switch key {
			case engine.KeyUp, engine.KeyW:
				l.selectedIdx = (l.selectedIdx - 1 + totalOptions) % totalOptions
				l.render(screen)

			case engine.KeyDown, engine.KeyS:
				l.selectedIdx = (l.selectedIdx + 1) % totalOptions
				l.render(screen)

			case engine.KeyEnter, engine.KeySpace:
				// Se selecionou "Sair da Estação"
				if l.selectedIdx == len(l.catalog) {
					return nil
				}

				selectedGame := l.catalog[l.selectedIdx]
				score, err := selectedGame.Run(ctx)
				if err != nil && ctx.Err() != nil {
					return err
				}

				// Atualiza a pontuação do jogador na sessão
				meta := selectedGame.Meta()
				st := l.stats[meta.ID]
				st.LastScore = score
				st.TimesPlayed++
				if score > st.BestScore {
					st.BestScore = score
				}
				l.totalScore += score

				input.Drain()
				screen.ClearScreen()
				l.render(screen)

			case engine.KeyEsc, engine.KeyQuit, engine.KeyQ:
				return nil
			}
		}
	}
}

func (l *Launcher) render(screen *engine.Screen) {
	screen.BeginFrame()

	boxWidth := 66
	border := strings.Repeat("═", boxWidth)

	screen.WriteLine("")
	screen.WriteLine("  " + engine.FgBrightCyan + "╔" + border + "╗" + engine.Reset)
	screen.WriteLine(fmt.Sprintf("  %s║%s%s%s              🕹️  JOGOS-CLI — ESTAÇÃO ARCADE (Go 1.27)            %s%s║%s",
		engine.FgBrightCyan, engine.Reset, engine.Bold, engine.FgBrightYellow, engine.Reset, engine.FgBrightCyan, engine.Reset))
	screen.WriteLine("  " + engine.FgBrightCyan + "╚" + border + "╝" + engine.Reset)
	screen.WriteLine("")

	// Painel de Pontuação do Jogador na Sessão
	bestOverall := 0
	lastOverall := 0
	if l.selectedIdx < len(l.catalog) {
		id := l.catalog[l.selectedIdx].Meta().ID
		if st := l.stats[id]; st != nil {
			bestOverall = st.BestScore
			lastOverall = st.LastScore
		}
	}

	screen.WriteLine(fmt.Sprintf("  %s📊 SUA PONTUAÇÃO NA SESSÃO:%s", engine.Bold, engine.Reset))
	screen.WriteLine(fmt.Sprintf("     Última Partida: %s%04d pts%s   │   Melhor no Jogo: %s%04d pts%s   │   Total Acumulado: %s%04d pts%s",
		engine.FgBrightYellow, lastOverall, engine.Reset,
		engine.FgBrightGreen, bestOverall, engine.Reset,
		engine.FgBrightCyan, l.totalScore, engine.Reset,
	))
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine("")

	// Lista de Jogos Disponíveis
	screen.WriteLine(fmt.Sprintf("  %sSELECIONE UM JOGO:%s", engine.Bold, engine.Reset))
	screen.WriteLine("")

	for i, g := range l.catalog {
		meta := g.Meta()
		st := l.stats[meta.ID]
		cursor := "   "
		titleStyle := engine.FgWhite
		if i == l.selectedIdx {
			cursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightGreen)
			titleStyle = engine.Bold + engine.FgBrightGreen
		}

		scoreBadge := fmt.Sprintf("%s(Melhor: %04d pts | Última: %04d pts)%s",
			engine.Dim, st.BestScore, st.LastScore, engine.Reset)

		screen.WriteLine(fmt.Sprintf("%s%s%s  %s", cursor, titleStyle, meta.Title, scoreBadge))
	}

	// Opção de Sair
	exitCursor := "   "
	exitStyle := engine.FgWhite
	if l.selectedIdx == len(l.catalog) {
		exitCursor = fmt.Sprintf("  %s%s▶ ", engine.Bold, engine.FgBrightRed)
		exitStyle = engine.Bold + engine.FgBrightRed
	}
	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("%s%s🚪 Sair da Estação%s", exitCursor, exitStyle, engine.Reset))
	screen.WriteLine("")
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)

	// Detalhes do item selecionado
	if l.selectedIdx < len(l.catalog) {
		meta := l.catalog[l.selectedIdx].Meta()
		screen.WriteLine(fmt.Sprintf("  %s%sSobre o Jogo:%s %s", engine.Bold, engine.FgBrightCyan, engine.Reset, meta.Description))
		screen.WriteLine(fmt.Sprintf("  %s%sControles:%s    %s", engine.Bold, engine.FgBrightYellow, engine.Reset, meta.Controls))
	} else {
		screen.WriteLine(fmt.Sprintf("  %sEncerra a estação de jogos e retorna ao terminal.%s", engine.Dim, engine.Reset))
		screen.WriteLine("")
	}

	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", boxWidth+2) + engine.Reset)
	screen.WriteLine(fmt.Sprintf("  %sNavegação: [↑/↓ ou W/S] Selecionar  •  [ENTER/ESPAÇO] Jogar  •  [ESC/Q] Sair%s", engine.Dim, engine.Reset))
	screen.Flush()
}
