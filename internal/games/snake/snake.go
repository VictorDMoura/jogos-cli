package snake

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

const (
	DefaultWidth  = 24
	DefaultHeight = 16
	PointsPerFood = 10
)

// Direction representa a direção de movimento da cobra.
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// IsOpposite verifica se duas direções são opostas (impede colisão imediata ao inverter).
func (d Direction) IsOpposite(other Direction) bool {
	return (d == DirUp && other == DirDown) ||
		(d == DirDown && other == DirUp) ||
		(d == DirLeft && other == DirRight) ||
		(d == DirRight && other == DirLeft)
}

// Point representa uma coordenada (X, Y) no tabuleiro.
type Point struct {
	X int
	Y int
}

// State armazena todo o estado lógico do jogo da cobrinha (independente de I/O).
type State struct {
	Width     int
	Height    int
	Body      []Point
	Dir       Direction
	turnQueue []Direction
	Food      Point
	Score     int
	BestScore int
	FoodsEaten int
	Paused    bool
	GameOver  bool
	Won       bool
}

// NewState cria um novo estado inicial do jogo com as dimensões informadas.
func NewState(width, height int) *State {
	s := &State{
		Width:  max(width, 8),
		Height: max(height, 8),
	}
	s.Reset()
	return s
}

// Reset reinicia a partida mantendo o recorde da sessão (BestScore).
func (s *State) Reset() {
	midX := s.Width / 2
	midY := s.Height / 2

	s.Body = []Point{
		{X: midX, Y: midY},
		{X: midX - 1, Y: midY},
		{X: midX - 2, Y: midY},
	}
	s.Dir = DirRight
	s.turnQueue = s.turnQueue[:0]
	s.Score = 0
	s.FoodsEaten = 0
	s.Paused = false
	s.GameOver = false
	s.Won = false
	s.SpawnFood()
}

// ChangeDirection enfileira uma mudança de direção válida (até 2 comandos por tick).
func (s *State) ChangeDirection(next Direction) {
	if s.GameOver || s.Paused {
		return
	}
	lastPlanned := s.Dir
	if len(s.turnQueue) > 0 {
		lastPlanned = s.turnQueue[len(s.turnQueue)-1]
	}
	if next == lastPlanned || next.IsOpposite(lastPlanned) {
		return
	}
	if len(s.turnQueue) < 2 {
		s.turnQueue = append(s.turnQueue, next)
	}
}

// TogglePause alterna o estado de pausa se o jogo estiver em andamento.
func (s *State) TogglePause() {
	if !s.GameOver {
		s.Paused = !s.Paused
	}
}

// Step avança o jogo em 1 tick.
func (s *State) Step() {
	if s.GameOver || s.Paused {
		return
	}

	if len(s.turnQueue) > 0 {
		s.Dir = s.turnQueue[0]
		s.turnQueue = s.turnQueue[1:]
	}

	head := s.Body[0]
	next := head
	switch s.Dir {
	case DirUp:
		next.Y--
	case DirDown:
		next.Y++
	case DirLeft:
		next.X--
	case DirRight:
		next.X++
	}

	// Verifica colisão com as bordas
	if next.X < 0 || next.X >= s.Width || next.Y < 0 || next.Y >= s.Height {
		s.GameOver = true
		return
	}

	willEat := next == s.Food

	// Verifica colisão com o próprio corpo.
	// Se não comer fruta, a ponta da cauda sairá da posição atual neste mesmo tick.
	checkBody := s.Body[:len(s.Body)-1]
	if willEat {
		checkBody = s.Body
	}
	for _, segment := range checkBody {
		if segment == next {
			s.GameOver = true
			return
		}
	}

	// Move a cabeça para frente
	s.Body = append([]Point{next}, s.Body...)

	if willEat {
		s.FoodsEaten++
		s.Score += PointsPerFood
		if s.Score > s.BestScore {
			s.BestScore = s.Score
		}
		if len(s.Body) >= s.Width*s.Height {
			s.Won = true
			s.GameOver = true
			return
		}
		s.SpawnFood()
	} else {
		// Remove o último segmento da cauda
		s.Body = s.Body[:len(s.Body)-1]
	}
}

// SpawnFood posiciona uma nova comida em uma célula livre aleatória.
func (s *State) SpawnFood() {
	occupied := make(map[Point]bool, len(s.Body))
	for _, p := range s.Body {
		occupied[p] = true
	}

	freeCells := make([]Point, 0, s.Width*s.Height-len(s.Body))
	for y := range s.Height {
		for x := range s.Width {
			pt := Point{X: x, Y: y}
			if !occupied[pt] {
				freeCells = append(freeCells, pt)
			}
		}
	}

	if len(freeCells) == 0 {
		return
	}
	s.Food = freeCells[rand.IntN(len(freeCells))]
}

// TickDuration calcula o intervalo entre ticks com base nas comidas consumidas.
func (s *State) TickDuration() time.Duration {
	baseMs := 125
	speedup := (s.FoodsEaten / 3) * 6
	ms := max(baseMs-speedup, 55)
	return time.Duration(ms) * time.Millisecond
}

// SpeedLevel retorna um nível numérico de velocidade para exibição na HUD.
func (s *State) SpeedLevel() int {
	return 1 + (s.FoodsEaten / 3)
}

// Game implementa a interface games.Game para o Jogo da Cobrinha.
type Game struct {
	bestScore int
}

// New cria uma nova instância do Jogo da Cobrinha.
func New() *Game {
	return &Game{}
}

// Meta retorna os metadados do jogo exibidos no menu principal da estação.
func (g *Game) Meta() games.Metadata {
	return games.Metadata{
		ID:          "snake",
		Title:       "🐍 Snake (Jogo da Cobrinha)",
		Description: "Guie a cobrinha pelo tabuleiro, colete as maçãs para crescer e somar pontos sem bater nas paredes ou no próprio corpo!",
		Controls:    "Setas / WASD: Mover | P / Espaço: Pausar | R: Reiniciar | ESC / Q: Menu",
	}
}

// Run inicia o loop interativo do jogo no terminal e retorna a pontuação ao sair.
func (g *Game) Run(ctx context.Context) (int, error) {
	screen := engine.NewScreen()
	input := engine.InputFromContext(ctx)
	input.Drain()

	state := NewState(DefaultWidth, DefaultHeight)
	state.BestScore = g.bestScore
	lastPlayedScore := 0

	screen.ClearScreen()
	renderFrame(screen, state)

	ticker := time.NewTicker(state.TickDuration())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return max(state.Score, lastPlayedScore), ctx.Err()

		case key := <-input.Events():
			switch key {
			case engine.KeyUp:
				state.ChangeDirection(DirUp)
			case engine.KeyDown:
				state.ChangeDirection(DirDown)
			case engine.KeyLeft:
				state.ChangeDirection(DirLeft)
			case engine.KeyRight:
				state.ChangeDirection(DirRight)
			case engine.KeyPause, engine.KeySpace:
				state.TogglePause()
				renderFrame(screen, state)
			case engine.KeyRestart:
				if state.Score > lastPlayedScore {
					lastPlayedScore = state.Score
				}
				state.Reset()
				ticker.Reset(state.TickDuration())
				renderFrame(screen, state)
			case engine.KeyEsc, engine.KeyQuit:
				if state.BestScore > g.bestScore {
					g.bestScore = state.BestScore
				}
				if state.Score > 0 {
					return state.Score, nil
				}
				return lastPlayedScore, nil
			}

		case <-ticker.C:
			prevFoods := state.FoodsEaten
			state.Step()
			if state.BestScore > g.bestScore {
				g.bestScore = state.BestScore
			}
			if state.Score > lastPlayedScore {
				lastPlayedScore = state.Score
			}
			if state.FoodsEaten != prevFoods {
				ticker.Reset(state.TickDuration())
			}
			renderFrame(screen, state)
		}
	}
}

func renderFrame(screen *engine.Screen, s *State) {
	screen.BeginFrame()

	borderWidth := s.Width * 2
	topBorder := "  ╔" + strings.Repeat("═", borderWidth) + "╗"
	bottomBorder := "  ╚" + strings.Repeat("═", borderWidth) + "╝"

	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("  %s%s🐍 SNAKE — ESTAÇÃO DE JOGOS CLI%s", engine.Bold, engine.FgBrightGreen, engine.Reset))
	screen.WriteLine(fmt.Sprintf("  %sPONTOS:%s %s%04d%s   │   %sMELHOR:%s %s%04d%s   │   %sTAMANHO:%s %d   │   %sVELOCIDADE:%s Nv.%d",
		engine.Bold, engine.Reset,
		engine.FgBrightYellow, s.Score, engine.Reset,
		engine.Bold, engine.Reset,
		engine.FgBrightCyan, s.BestScore, engine.Reset,
		engine.Bold, engine.Reset, len(s.Body),
		engine.Bold, engine.Reset, s.SpeedLevel(),
	))
	screen.WriteLine(engine.FgCyan + topBorder + engine.Reset)

	// Mapa rápido de posições do corpo para renderização O(1) por célula
	bodyIndex := make(map[Point]int, len(s.Body))
	for i, pt := range s.Body {
		// Guarda o menor índice caso haja sobreposição
		if _, exists := bodyIndex[pt]; !exists {
			bodyIndex[pt] = i
		}
	}

	var rowBuilder strings.Builder
	for y := range s.Height {
		rowBuilder.Reset()
		rowBuilder.WriteString("  ")
		rowBuilder.WriteString(engine.FgCyan + "║" + engine.Reset)

		for x := range s.Width {
			pt := Point{X: x, Y: y}
			if idx, ok := bodyIndex[pt]; ok {
				if idx == 0 {
					// Cabeça da cobra
					rowBuilder.WriteString(engine.FgBrightGreen + "██" + engine.Reset)
				} else {
					// Corpo da cobra
					rowBuilder.WriteString(engine.FgGreen + "▓▓" + engine.Reset)
				}
			} else if pt == s.Food {
				// Maçã / Comida
				rowBuilder.WriteString(engine.FgBrightRed + "●●" + engine.Reset)
			} else {
				// Espaço vazio com leve guia pontilhada
				rowBuilder.WriteString(engine.Dim + " ·" + engine.Reset)
			}
		}

		rowBuilder.WriteString(engine.FgCyan + "║" + engine.Reset)
		screen.WriteLine(rowBuilder.String())
	}

	screen.WriteLine(engine.FgCyan + bottomBorder + engine.Reset)

	// Mensagens de status (Game Over / Vitória / Pausa / Em jogo)
	if s.GameOver {
		if s.Won {
			screen.WriteLine(fmt.Sprintf("  %s%s🏆 PARABÉNS! VOCÊ VENCEU COM %d PONTOS!%s", engine.Bold, engine.FgBrightGreen, s.Score, engine.Reset))
		} else {
			screen.WriteLine(fmt.Sprintf("  %s%s💥 FIM DE JOGO! Pontuação Final: %d pontos%s", engine.Bold, engine.FgBrightRed, s.Score, engine.Reset))
		}
		screen.WriteLine(fmt.Sprintf("  %sPressione [R] para jogar novamente ou [ESC / Q] para voltar ao Menu.%s", engine.FgBrightYellow, engine.Reset))
	} else if s.Paused {
		screen.WriteLine(fmt.Sprintf("  %s%s⏸  JOGO PAUSADO — Pressione [P] ou [ESPAÇO] para continuar%s", engine.Bold, engine.FgBrightYellow, engine.Reset))
		screen.WriteLine("")
	} else {
		screen.WriteLine(fmt.Sprintf("  %sDica: Colete %s●●%s%s para ganhar +%d pontos!%s", engine.Dim, engine.FgBrightRed, engine.Reset, engine.Dim, PointsPerFood, engine.Reset))
		screen.WriteLine("")
	}

	screen.WriteLine(fmt.Sprintf("  %sControles: [Setas/WASD] Mover  •  [P/Espaço] Pausar  •  [R] Reiniciar  •  [ESC/Q] Menu%s", engine.Dim, engine.Reset))
	screen.Flush()
}
