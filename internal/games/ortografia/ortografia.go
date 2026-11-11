package ortografia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"time"
	"unicode"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

const (
	BoardWidth         = 56
	BoardHeight        = 12
	InitialLives       = 5
	PointsPerHit       = 15
	StreakBonusPoints  = 5
	DefaultStoragePath = "data/ortografia_challenges.json"
)

// Challenge representa um desafio ortográfico com nível, palavra lacunada, letra correta e explicação pedagógica.
type Challenge struct {
	Level      int    `json:"level"`       // 1 = Infantil/Base, 2 = Intermediário, 3 = Avançado/Concurso
	FullWord   string `json:"full_word"`   // Ex: "TAMBOR"
	MaskedWord string `json:"masked_word"` // Ex: "TA_BOR"
	CorrectKey rune   `json:"correct_key"` // Ex: 'M'
	WrongKey   rune   `json:"wrong_key"`   // Ex: 'N'
	Category   string `json:"category"`    // Ex: "M ou N"
	Hint       string `json:"hint"`        // Dica semântica da palavra
	Rule       string `json:"rule"`        // Explicação gramatical/ortográfica (Bizú)
}

// DefaultChallenges contém os 18 desafios progressivos pré-prontos (do Infantil ao Concurso).
var DefaultChallenges = []Challenge{
	// NÍVEL 1: Infantil / Base (M/N, S/Z básico, G/J básico, X/C)
	{
		Level:      1,
		FullWord:   "TAMBOR",
		MaskedWord: "TA_BOR",
		CorrectKey: 'M',
		WrongKey:   'N',
		Category:   "M ou N",
		Hint:       "Instrumento musical de percussão",
		Rule:       "Regra de Ouro: Antes das consoantes P e B, usamos sempre a letra M!",
	},
	{
		Level:      1,
		FullWord:   "LAMPADA",
		MaskedWord: "LÂ_PADA",
		CorrectKey: 'M',
		WrongKey:   'N',
		Category:   "M ou N",
		Hint:       "Objeto que ilumina o quarto",
		Rule:       "Regra de Ouro: Antes de P e B (lâmpada, tempo, pomba), usa-se sempre M!",
	},
	{
		Level:      1,
		FullWord:   "CINTO",
		MaskedWord: "CI_TO",
		CorrectKey: 'N',
		WrongKey:   'M',
		Category:   "M ou N",
		Hint:       "Acessório usado para segurar a calça",
		Rule:       "Antes de consoantes que NÃO são P nem B (como T, D, C), usamos sempre N!",
	},
	{
		Level:      1,
		FullWord:   "CASA",
		MaskedWord: "CA_A",
		CorrectKey: 'S',
		WrongKey:   'Z',
		Category:   "S ou Z",
		Hint:       "Lugar onde moramos com a família",
		Rule:       "O 'S' sozinho entre duas vogais (intervocálico) tem som de Z: casa, mesa, rosa.",
	},
	{
		Level:      1,
		FullWord:   "BELEZA",
		MaskedWord: "BELE_A",
		CorrectKey: 'Z',
		WrongKey:   'S',
		Category:   "S ou Z",
		Hint:       "Qualidade do que é belo, bonito",
		Rule:       "Substantivos formados a partir de adjetivos (belo -> beleza, rico -> riqueza) usam -EZA com Z!",
	},
	{
		Level:      1,
		FullWord:   "RELOGIO",
		MaskedWord: "RELÓ_IO",
		CorrectKey: 'G',
		WrongKey:   'J',
		Category:   "G ou J",
		Hint:       "Aparelho que marca as horas",
		Rule:       "Palavras terminadas em -ágio, -égio, -ígio, -ógio e -úgio escrevem-se sempre com G!",
	},
	{
		Level:      1,
		FullWord:   "JEITO",
		MaskedWord: "_EITO",
		CorrectKey: 'J',
		WrongKey:   'G',
		Category:   "G ou J",
		Hint:       "Modo, maneira ou habilidade de fazer algo",
		Rule:       "'Jeito' e todas as palavras da sua família (ajeitar, desajeitado, rejeitar) escrevem-se com J!",
	},
	{
		Level:      1,
		FullWord:   "BRUXA",
		MaskedWord: "BRU_A",
		CorrectKey: 'X',
		WrongKey:   'C',
		Category:   "X ou CH",
		Hint:       "Personagem dos contos de fadas que voa na vassoura",
		Rule:       "'Bruxa', 'bexiga' e 'lagartixa' são grafadas com X!",
	},

	// NÍVEL 2: Fundamental / Intermediário
	{
		Level:      2,
		FullWord:   "PESQUISA",
		MaskedWord: "PESQUI_A",
		CorrectKey: 'S',
		WrongKey:   'Z',
		Category:   "S ou Z",
		Hint:       "Investigação ou busca de informações para a escola",
		Rule:       "'Pesquisa' e 'pesquisar' são grafadas com S na segunda sílaba!",
	},
	{
		Level:      2,
		FullWord:   "VIAGEM",
		MaskedWord: "VIA_EM",
		CorrectKey: 'G',
		WrongKey:   'J',
		Category:   "G ou J",
		Hint:       "O passeio para outra cidade (substantivo: 'A boa ___')",
		Rule:       "Substantivo ('a viagem') é sempre com G! Com J ('viajem') é apenas o verbo ('espero que eles viajem').",
	},
	{
		Level:      2,
		FullWord:   "CORAGEM",
		MaskedWord: "CORA_EM",
		CorrectKey: 'G',
		WrongKey:   'J',
		Category:   "G ou J",
		Hint:       "Bravura, ausência de medo diante do perigo",
		Rule:       "Substantivos terminados em -AGEM, -IGEM e -UGEM (coragem, vertigem, ferrugem) escrevem-se com G!",
	},
	{
		Level:      2,
		FullWord:   "ENXADA",
		MaskedWord: "EN_ADA",
		CorrectKey: 'X',
		WrongKey:   'C',
		Category:   "X ou CH",
		Hint:       "Ferramenta agrícola usada para capinar a terra",
		Rule:       "Após a sílaba inicial 'EN-', a regra geral é usar X (enxada, enxame, enxaqueca, enxugar). Exceção: encher/enchova.",
	},
	{
		Level:      2,
		FullWord:   "FAXINA",
		MaskedWord: "FA_INA",
		CorrectKey: 'X',
		WrongKey:   'C',
		Category:   "X ou CH",
		Hint:       "Limpeza completa e caprichada da casa",
		Rule:       "'Faxina' e 'faxineiro' escrevem-se sempre com X!",
	},
	{
		Level:      2,
		FullWord:   "ANALISAR",
		MaskedWord: "ANALI_AR",
		CorrectKey: 'S',
		WrongKey:   'Z',
		Category:   "S ou Z",
		Hint:       "Examinar com atenção (verbo derivado de 'análise')",
		Rule:       "Como 'análise' já possui S no radical, o verbo derivado mantém o S: ANALISAR!",
	},

	// NÍVEL 3: Avançado / Concursos & ENEM
	{
		Level:      3,
		FullWord:   "PRIVILEGIO",
		MaskedWord: "PRIVIL_GIO",
		CorrectKey: 'E',
		WrongKey:   'I',
		Category:   "E ou I (Concurso)",
		Hint:       "Vantagem ou direito especial concedido a alguém",
		Rule:       "Pegadinha clássica de prova: escreve-se PRIVILÉGIO (com E), além de terminação -ÉGIO com G!",
	},
	{
		Level:      3,
		FullWord:   "EMPECILHO",
		MaskedWord: "EMPEC_LHO",
		CorrectKey: 'I',
		WrongKey:   'E',
		Category:   "E ou I (Concurso)",
		Hint:       "Obstáculo, impedimento que atrapalha um plano",
		Rule:       "Muito cobrado em concursos: escreve-se EMPECILHO (começa com E e tem I no meio)!",
	},
	{
		Level:      3,
		FullWord:   "PARALISAR",
		MaskedWord: "PARALI_AR",
		CorrectKey: 'S',
		WrongKey:   'Z',
		Category:   "S ou Z (Concurso)",
		Hint:       "Interromper o movimento, derivados de 'paralisia'",
		Rule:       "Como deriva de 'paralisia' (que tem S), o verbo escreve-se PARALISAR com S!",
	},
	{
		Level:      3,
		FullWord:   "CATEQUIZAR",
		MaskedWord: "CATEQUI_AR",
		CorrectKey: 'Z',
		WrongKey:   'S',
		Category:   "S ou Z (Concurso)",
		Hint:       "Instruir na fé ou doutrina (formado com sufixo -izar)",
		Rule:       "Atenção: 'catequese' tem S, mas a raiz 'catequ-' recebe o sufixo verbal -IZAR com Z: CATEQUIZAR!",
	},
	{
		Level:      3,
		FullWord:   "EXCECAO",
		MaskedWord: "E_CEÇÃO",
		CorrectKey: 'X',
		WrongKey:   'S',
		Category:   "X ou S (Concurso)",
		Hint:       "Aquilo que foge à regra geral",
		Rule:       "'Exceção', 'excepcional' e 'exceder' possuem o dígrafo XC após o E inicial!",
	},
	{
		Level:      3,
		FullWord:   "OBSEQUIO",
		MaskedWord: "OB_ÉQUIO",
		CorrectKey: 'S',
		WrongKey:   'Z',
		Category:   "S ou Z (Concurso)",
		Hint:       "Favor, gentileza ('Por ___')",
		Rule:       "Apesar de ter som de Z, após o prefixo OB- grafa-se com S: OBSÉQUIO, obsequiar!",
	},
}

type savedChallengesFile struct {
	CustomChallenges []Challenge `json:"custom_challenges"`
}

// LoadCustomChallenges carrega desafios personalizados salvos em disco.
func LoadCustomChallenges(storagePath string) ([]Challenge, error) {
	if storagePath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(storagePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var payload savedChallengesFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload.CustomChallenges, nil
}

// State armazena o estado puro do jogo Chuva da Ortografia.
type State struct {
	Challenges    []Challenge
	Current       Challenge
	DropX         int
	DropY         int
	Score         int
	BestScore     int
	Lives         int
	Streak        int
	Hits          int
	CurrentLevel  int
	Paused        bool
	GameOver      bool
	LastFeedback  string
	LastRule      string
	LastWasHit    bool
}

// NewState cria um novo estado inicial para a Chuva da Ortografia.
func NewState(challenges []Challenge) *State {
	pool := slices.Clone(challenges)
	if len(pool) == 0 {
		pool = slices.Clone(DefaultChallenges)
	}
	s := &State{
		Challenges: pool,
	}
	s.Reset()
	return s
}

// Reset reinicia a partida no Nível 1 (Infantil/Base), mantendo o BestScore.
func (s *State) Reset() {
	s.Score = 0
	s.Lives = InitialLives
	s.Streak = 0
	s.Hits = 0
	s.CurrentLevel = 1
	s.Paused = false
	s.GameOver = false
	s.LastFeedback = "Aperte no teclado a letra que falta antes que a palavra chegue ao chão!"
	s.LastRule = "Você começa no Nível 1 (Infantil/Base) e avança até o Nível 3 (Concursos/ENEM) conforme acerta!"
	s.LastWasHit = true
	s.SpawnNextWord()
}

// UpdateProgression ajusta o nível atual conforme a quantidade de acertos do jogador.
func (s *State) UpdateProgression() {
	switch {
	case s.Hits >= 10:
		s.CurrentLevel = 3
	case s.Hits >= 5:
		s.CurrentLevel = 2
	default:
		s.CurrentLevel = 1
	}
}

// SpawnNextWord sorteia uma palavra compatível com o nível atual do jogador.
func (s *State) SpawnNextWord() {
	s.UpdateProgression()

	prevWord := s.Current.FullWord
	candidates := make([]Challenge, 0, len(s.Challenges))
	for _, c := range s.Challenges {
		if c.Level <= s.CurrentLevel {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		candidates = s.Challenges
	}

	// Prioriza questões do nível atual em 70% das vezes para reforçar o avanço
	levelMatches := make([]Challenge, 0, len(candidates))
	for _, c := range candidates {
		if c.Level == s.CurrentLevel {
			levelMatches = append(levelMatches, c)
		}
	}
	if len(levelMatches) > 0 && rand.IntN(10) < 7 {
		candidates = levelMatches
	}

	for range 10 {
		pick := candidates[rand.IntN(len(candidates))]
		if pick.FullWord != prevWord || len(candidates) == 1 {
			s.Current = pick
			break
		}
		s.Current = pick
	}

	s.DropY = 0
	wordLen := len([]rune(s.Current.MaskedWord))
	maxX := max(BoardWidth-wordLen-6, 4)
	s.DropX = 3 + rand.IntN(maxX)
}

// Step faz a palavra cair 1 linha em direção ao chão.
func (s *State) Step() {
	if s.GameOver || s.Paused {
		return
	}

	s.DropY++
	if s.DropY >= BoardHeight {
		s.Lives--
		s.Streak = 0
		s.LastWasHit = false
		s.LastFeedback = fmt.Sprintf("⏰ A palavra caiu no chão! O correto era '%s' (letra %c).", s.Current.FullWord, s.Current.CorrectKey)
		s.LastRule = s.Current.Rule
		if s.Lives <= 0 {
			s.GameOver = true
			return
		}
		s.SpawnNextWord()
	}
}

// Answer verifica a letra pressionada pelo jogador.
func (s *State) Answer(letter rune) {
	if s.GameOver || s.Paused {
		return
	}
	letter = unicode.ToUpper(letter)
	if letter < 'A' || letter > 'Z' {
		return
	}

	if letter == unicode.ToUpper(s.Current.CorrectKey) {
		s.Hits++
		s.Streak++
		gained := PointsPerHit + (s.Streak-1)*StreakBonusPoints
		s.Score += gained
		if s.Score > s.BestScore {
			s.BestScore = s.Score
		}
		s.LastWasHit = true
		s.LastFeedback = fmt.Sprintf("✅ CERTO! %s escreve-se com '%c' (+%d pts | Sequência: %dx)",
			s.Current.FullWord, s.Current.CorrectKey, gained, s.Streak)
		s.LastRule = s.Current.Rule
		s.SpawnNextWord()
	} else {
		s.Lives--
		s.Streak = 0
		s.LastWasHit = false
		s.LastFeedback = fmt.Sprintf("❌ OPS! Você apertou '%c', mas o correto em %s é '%c'.",
			letter, s.Current.FullWord, s.Current.CorrectKey)
		s.LastRule = s.Current.Rule
		if s.Lives <= 0 {
			s.GameOver = true
			return
		}
		s.SpawnNextWord()
	}
}

// TogglePause pausa ou retoma a queda da palavra.
func (s *State) TogglePause() {
	if !s.GameOver {
		s.Paused = !s.Paused
	}
}

// TickDuration define a velocidade de queda da palavra (suave para dar tempo de ler a dica).
func (s *State) TickDuration() time.Duration {
	baseMs := 650
	speedup := (s.Hits / 4) * 35
	ms := max(baseMs-speedup, 380)
	return time.Duration(ms) * time.Millisecond
}

// LevelTitle retorna o nome pedagógico do nível atual.
func (s *State) LevelTitle() string {
	switch s.CurrentLevel {
	case 1:
		return "🟢 Nível 1: Base Infantil (M/N, S/Z, G/J)"
	case 2:
		return "🟡 Nível 2: Intermediário (Enx-, -agem, -isar)"
	default:
		return "🔴 Nível 3: Avançado / Concurso & ENEM"
	}
}

// Game implementa a interface games.Game para a Chuva da Ortografia.
type Game struct {
	storagePath string
	challenges  []Challenge
	bestScore   int
}

// New cria uma nova instância da Chuva da Ortografia.
func New() *Game {
	return NewWithStorage(DefaultStoragePath)
}

// NewWithStorage cria uma instância com caminho de arquivo configurável.
func NewWithStorage(storagePath string) *Game {
	challenges := slices.Clone(DefaultChallenges)
	if custom, err := LoadCustomChallenges(storagePath); err == nil && len(custom) > 0 {
		challenges = append(challenges, custom...)
	}
	return &Game{
		storagePath: storagePath,
		challenges:  challenges,
	}
}

// Meta retorna os metadados para exibição no menu principal da estação.
func (g *Game) Meta() games.Metadata {
	return games.Metadata{
		ID:          "ortografia",
		Title:       "🌧️ Chuva da Ortografia (Português: Infantil ao Concurso)",
		Description: "Complete a letra correta das palavras antes que toquem o chão! Começa com regras infantis (M/N, S/Z, G/J) e avança até pegadinhas de Concursos e ENEM com explicação de cada regra.",
		Controls:    "Teclas [A-Z] ou [1/2]: Escolher Letra | Espaço: Pausar | R: Reiniciar | ESC: Menu",
	}
}

// Run executa o loop em tempo real da Chuva da Ortografia.
func (g *Game) Run(ctx context.Context) (int, error) {
	screen := engine.NewScreen()
	input := engine.InputFromContext(ctx)
	input.Drain()

	state := NewState(g.challenges)
	state.BestScore = g.bestScore
	sessionMaxScore := 0

	screen.ClearScreen()
	renderFrame(screen, state)

	ticker := time.NewTicker(state.TickDuration())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return max(state.Score, sessionMaxScore), ctx.Err()

		case key := <-input.Events():
			switch {
			case key == engine.KeyEsc || key == engine.KeyQuit || (state.GameOver && key == engine.KeyQ):
				if state.BestScore > g.bestScore {
					g.bestScore = state.BestScore
				}
				return max(state.Score, sessionMaxScore), nil

			case key == engine.KeySpace:
				if state.GameOver {
					if state.Score > sessionMaxScore {
						sessionMaxScore = state.Score
					}
					state.Reset()
					ticker.Reset(state.TickDuration())
				} else {
					state.TogglePause()
				}
				renderFrame(screen, state)

			case state.GameOver && (key == engine.KeyRestart || key == engine.KeyEnter):
				if state.Score > sessionMaxScore {
					sessionMaxScore = state.Score
				}
				state.Reset()
				ticker.Reset(state.TickDuration())
				renderFrame(screen, state)

			case !state.GameOver && !state.Paused:
				// Suporta atalhos numéricos [1] e [2] para as duas opções exibidas na tela
				opt1, opt2 := orderedOptions(state.Current)
				if key == engine.Key('1') {
					state.Answer(opt1)
				} else if key == engine.Key('2') {
					state.Answer(opt2)
				} else if key.IsLetter() {
					state.Answer(key.UpperRune())
				} else {
					continue
				}

				if state.Score > sessionMaxScore {
					sessionMaxScore = state.Score
				}
				if state.BestScore > g.bestScore {
					g.bestScore = state.BestScore
				}
				ticker.Reset(state.TickDuration())
				renderFrame(screen, state)
			}

		case <-ticker.C:
			state.Step()
			renderFrame(screen, state)
		}
	}
}

// orderedOptions ordena as duas opções em ordem alfabética para que a posição correta não seja sempre a 1ª.
func orderedOptions(c Challenge) (rune, rune) {
	a := unicode.ToUpper(c.CorrectKey)
	b := unicode.ToUpper(c.WrongKey)
	if a > b {
		return b, a
	}
	return a, b
}

func renderFrame(screen *engine.Screen, s *State) {
	screen.BeginFrame()

	hearts := strings.Repeat("❤️ ", s.Lives) + strings.Repeat("🖤 ", max(InitialLives-s.Lives, 0))

	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("  %s%s🌧️  CHUVA DA ORTOGRAFIA — PORTUGUÊS PROGRESSIVO%s", engine.Bold, engine.FgBrightCyan, engine.Reset))
	screen.WriteLine(fmt.Sprintf("  %sPONTOS:%s %s%04d%s  │  %sMELHOR:%s %s%04d%s  │  %sACERTOS:%s %d  │  %sVIDAS:%s %s",
		engine.Bold, engine.Reset,
		engine.FgBrightYellow, s.Score, engine.Reset,
		engine.Bold, engine.Reset,
		engine.FgBrightGreen, s.BestScore, engine.Reset,
		engine.Bold, engine.Reset, s.Hits,
		engine.Bold, engine.Reset, hearts,
	))
	screen.WriteLine(fmt.Sprintf("  %sFASE ATUAL:%s %s   │   %sTEMA:%s %s%s%s",
		engine.Bold, engine.Reset, s.LevelTitle(),
		engine.Bold, engine.Reset, engine.FgBrightYellow, s.Current.Category, engine.Reset,
	))

	// Dica da palavra atual e botões de resposta rápida
	opt1, opt2 := orderedOptions(s.Current)
	screen.WriteLine(fmt.Sprintf("  %s💡 Dica da Palavra:%s %s\"%s\"%s   👉 Aperte %s[%c]%s (1) ou %s[%c]%s (2) no teclado!",
		engine.Bold, engine.Reset,
		engine.FgBrightWhite, s.Current.Hint, engine.Reset,
		engine.Bold+engine.FgBrightGreen, opt1, engine.Reset,
		engine.Bold+engine.FgBrightGreen, opt2, engine.Reset,
	))

	topBorder := "  ╔" + strings.Repeat("═", BoardWidth) + "╗"
	bottomBorder := "  ╚" + strings.Repeat("═", BoardWidth) + "╝"

	screen.WriteLine(engine.FgCyan + topBorder + engine.Reset)

	for y := range BoardHeight {
		if y == s.DropY && !s.GameOver {
			badge := fmt.Sprintf(" ☁️  %s ", s.Current.MaskedWord)
			badgeRuneLen := len([]rune(s.Current.MaskedWord)) + 5
			leftPad := min(max(s.DropX, 1), max(BoardWidth-badgeRuneLen, 1))
			rightPad := max(BoardWidth-leftPad-badgeRuneLen, 0)

			dropColor := engine.FgBrightGreen
			if y >= BoardHeight*2/3 {
				dropColor = engine.FgBrightRed
			} else if y >= BoardHeight/3 {
				dropColor = engine.FgBrightYellow
			}

			line := "  " + engine.FgCyan + "║" + engine.Reset +
				strings.Repeat(" ", leftPad) +
				engine.Bold + dropColor + badge + engine.Reset +
				strings.Repeat(" ", rightPad) +
				engine.FgCyan + "║" + engine.Reset
			screen.WriteLine(line)
		} else {
			screen.WriteLine("  " + engine.FgCyan + "║" + engine.Reset + strings.Repeat(" ", BoardWidth) + engine.FgCyan + "║" + engine.Reset)
		}
	}

	// Linha do chão
	screen.WriteLine("  " + engine.FgCyan + "╠" + strings.Repeat("♒", BoardWidth) + "╣" + engine.Reset)
	screen.WriteLine(engine.FgCyan + bottomBorder + engine.Reset)

	// Painel Pedagógico ("Bizú do Professor")
	feedbackColor := engine.FgBrightGreen
	if !s.LastWasHit {
		feedbackColor = engine.FgBrightRed
	}
	screen.WriteLine(fmt.Sprintf("  %s%s%s", engine.Bold, feedbackColor, s.LastFeedback+engine.Reset))
	screen.WriteLine(fmt.Sprintf("  %s📚 BIZÚ / REGRA:%s %s%s%s", engine.Bold, engine.Reset, engine.FgBrightYellow, s.LastRule, engine.Reset))

	if s.GameOver {
		screen.WriteLine(fmt.Sprintf("  %s%s🏁 FIM DE JOGO! Você acertou %d palavras e fez %d pontos!%s",
			engine.Bold, engine.FgBrightRed, s.Hits, s.Score, engine.Reset))
		screen.WriteLine(fmt.Sprintf("  %sPressione [R / ENTER / ESPAÇO] para jogar novamente ou [ESC / Q] para voltar ao Menu.%s",
			engine.FgBrightCyan, engine.Reset))
	} else if s.Paused {
		screen.WriteLine(fmt.Sprintf("  %s%s⏸  PAUSADO (Aproveite para ler a regra acima!) — Pressione [ESPAÇO] para continuar.%s",
			engine.Bold, engine.FgBrightYellow, engine.Reset))
	} else {
		screen.WriteLine(fmt.Sprintf("  %sControles: [Digite a Letra A-Z ou 1/2] Responder  •  [ESPAÇO] Pausar p/ ler a regra  •  [ESC] Menu%s",
			engine.Dim, engine.Reset))
	}

	screen.Flush()
}
