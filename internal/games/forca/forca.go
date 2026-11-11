package forca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"jogos-cli/internal/engine"
	"jogos-cli/internal/games"
)

const (
	MaxErrors          = 6
	PointsPerLetter    = 10
	WinBaseBonus       = 50
	MaxWordLength      = 20
	MinWordLength      = 2
	DefaultStoragePath = "data/forca_words.json"
)

// DefaultWords contém as 10 palavras pré-prontas iniciais do jogo da forca.
var DefaultWords = []string{
	"COMPUTADOR",
	"PROGRAMADOR",
	"ALGORITMO",
	"TERMINAL",
	"TECLADO",
	"INTERNET",
	"SERVIDOR",
	"VARIAVEL",
	"COMPILADOR",
	"SISTEMA",
}

type savedWordsFile struct {
	CustomWords []string `json:"custom_words"`
}

// LoadCustomWords carrega as palavras personalizadas salvas em disco.
// Se o arquivo ainda não existir, retorna uma fatia vazia sem erro.
func LoadCustomWords(storagePath string) ([]string, error) {
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

	var payload savedWordsFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	cleaned := make([]string, 0, len(payload.CustomWords))
	for _, w := range payload.CustomWords {
		sw := SanitizeWord(w)
		if len(sw) >= MinWordLength && !slices.Contains(DefaultWords, sw) && !slices.Contains(cleaned, sw) {
			cleaned = append(cleaned, sw)
		}
	}
	return cleaned, nil
}

// SaveCustomWords salva a lista de palavras personalizadas em disco no formato JSON.
func SaveCustomWords(storagePath string, customWords []string) error {
	if storagePath == "" {
		return nil
	}
	dir := filepath.Dir(storagePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	payload := savedWordsFile{
		CustomWords: customWords,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storagePath, data, 0o644)
}

// Mode define se o jogador está chutando letras ou cadastrando uma nova palavra.
type Mode int

const (
	ModePlaying Mode = iota
	ModeAddingWord
)

// State representa o estado puro e as regras de negócio do Jogo da Forca.
type State struct {
	Words            []string
	CustomWords      []string
	CustomWordsCount int
	TargetWord       string
	Guessed          map[rune]bool
	Errors           int
	Score            int
	BestScore        int
	WordsSolved      int
	RoundOver        bool
	Won              bool
	Mode             Mode
	InputBuffer      []rune
	StatusMessage    string
}

// NewState inicializa o estado do jogo da forca com as palavras informadas e sorteia a primeira.
func NewState(initialWords []string) *State {
	words := make([]string, 0, len(initialWords))
	for _, w := range initialWords {
		if clean := SanitizeWord(w); len(clean) >= MinWordLength && !slices.Contains(words, clean) {
			words = append(words, clean)
		}
	}
	if len(words) == 0 {
		words = slices.Clone(DefaultWords)
	}

	var custom []string
	for _, w := range words {
		if !slices.Contains(DefaultWords, w) {
			custom = append(custom, w)
		}
	}

	s := &State{
		Words:            words,
		CustomWords:      custom,
		CustomWordsCount: len(custom),
		Guessed:          make(map[rune]bool),
		Mode:             ModePlaying,
	}
	s.StartNewRound()
	return s
}

// SanitizeWord normaliza uma palavra para letras maiúsculas sem acentos (A-Z).
func SanitizeWord(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		r = unicode.ToUpper(r)
		switch r {
		case 'Á', 'À', 'Â', 'Ã', 'Ä':
			r = 'A'
		case 'É', 'È', 'Ê', 'Ë':
			r = 'E'
		case 'Í', 'Ì', 'Î', 'Ï':
			r = 'I'
		case 'Ó', 'Ò', 'Ô', 'Õ', 'Ö':
			r = 'O'
		case 'Ú', 'Ù', 'Û', 'Ü':
			r = 'U'
		case 'Ç':
			r = 'C'
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// StartNewRound sorteia aleatoriamente uma palavra do banco e reinicia as tentativas da rodada.
func (s *State) StartNewRound() {
	// Se perdeu a rodada anterior, reinicia a pontuação da sequência atual mantendo o BestScore
	if s.RoundOver && !s.Won {
		s.Score = 0
		s.WordsSolved = 0
	}

	prevWord := s.TargetWord
	if len(s.Words) == 1 {
		s.TargetWord = s.Words[0]
	} else if len(s.Words) > 1 {
		for range 10 {
			candidate := s.Words[rand.IntN(len(s.Words))]
			if candidate != prevWord {
				s.TargetWord = candidate
				break
			}
			s.TargetWord = candidate
		}
	}

	clear(s.Guessed)
	s.Errors = 0
	s.RoundOver = false
	s.Won = false
	s.Mode = ModePlaying
	s.StatusMessage = "Digite uma letra de A a Z para adivinhar a palavra secreta!"
}

// GuessLetter processa o chute de uma letra pelo jogador.
func (s *State) GuessLetter(letter rune) {
	if s.RoundOver || s.Mode != ModePlaying {
		return
	}

	letter = unicode.ToUpper(letter)
	if letter < 'A' || letter > 'Z' {
		return
	}

	if s.Guessed[letter] {
		s.StatusMessage = fmt.Sprintf("A letra '%c' já foi tentada! Escolha outra letra.", letter)
		return
	}

	s.Guessed[letter] = true

	if strings.ContainsRune(s.TargetWord, letter) {
		s.Score += PointsPerLetter
		if s.Score > s.BestScore {
			s.BestScore = s.Score
		}
		s.StatusMessage = fmt.Sprintf("Boa! A letra '%c' existe na palavra (+%d pts).", letter, PointsPerLetter)

		if s.isWordComplete() {
			s.Won = true
			s.RoundOver = true
			s.WordsSolved++
			bonus := WinBaseBonus + (MaxErrors-s.Errors)*10
			s.Score += bonus
			if s.Score > s.BestScore {
				s.BestScore = s.Score
			}
			s.StatusMessage = fmt.Sprintf("🎉 Você descobriu a palavra '%s'! Bônus de vitória: +%d pts!", s.TargetWord, bonus)
		}
	} else {
		s.Errors++
		s.StatusMessage = fmt.Sprintf("Que pena! A letra '%c' não existe na palavra (%d/%d erros).", letter, s.Errors, MaxErrors)
		if s.Errors >= MaxErrors {
			s.RoundOver = true
			s.Won = false
			s.StatusMessage = fmt.Sprintf("💀 Fim de jogo! A palavra secreta era: %s", s.TargetWord)
		}
	}
}

func (s *State) isWordComplete() bool {
	for _, r := range s.TargetWord {
		if !s.Guessed[r] {
			return false
		}
	}
	return len(s.TargetWord) > 0
}

// MaskedWord retorna a palavra com letras descobertas e '_' nas ocultas (ex: "T E _ M I N A L").
func (s *State) MaskedWord() string {
	parts := make([]string, 0, len(s.TargetWord))
	for _, r := range s.TargetWord {
		if s.Guessed[r] || (s.RoundOver && !s.Won) {
			parts = append(parts, string(r))
		} else {
			parts = append(parts, "_")
		}
	}
	return strings.Join(parts, " ")
}

// EnterAddWordMode abre o formulário para o usuário adicionar uma nova palavra ao banco.
func (s *State) EnterAddWordMode() {
	s.Mode = ModeAddingWord
	s.InputBuffer = s.InputBuffer[:0]
	s.StatusMessage = "Digite a nova palavra (A-Z) e pressione [ENTER] para salvar ou [ESC] para cancelar."
}

// CancelAddWordMode sai do modo de adição de palavra e volta para a partida.
func (s *State) CancelAddWordMode() {
	s.Mode = ModePlaying
	s.InputBuffer = s.InputBuffer[:0]
	s.StatusMessage = "Adição de palavra cancelada."
}

// TypeWordLetter adiciona uma letra ao buffer da nova palavra.
func (s *State) TypeWordLetter(letter rune) {
	if s.Mode != ModeAddingWord {
		return
	}
	letter = unicode.ToUpper(letter)
	if letter >= 'A' && letter <= 'Z' && len(s.InputBuffer) < MaxWordLength {
		s.InputBuffer = append(s.InputBuffer, letter)
	}
}

// BackspaceWordLetter apaga a última letra digitada no modo de adição de palavra.
func (s *State) BackspaceWordLetter() {
	if s.Mode == ModeAddingWord && len(s.InputBuffer) > 0 {
		s.InputBuffer = s.InputBuffer[:len(s.InputBuffer)-1]
	}
}

// ConfirmAddWord valida e adiciona a palavra digitada à lista de palavras do jogo.
func (s *State) ConfirmAddWord() bool {
	if s.Mode != ModeAddingWord {
		return false
	}
	word := SanitizeWord(string(s.InputBuffer))
	if len(word) < MinWordLength {
		s.StatusMessage = fmt.Sprintf("⚠️  A palavra deve ter pelo menos %d letras!", MinWordLength)
		return false
	}
	if slices.Contains(s.Words, word) {
		s.StatusMessage = fmt.Sprintf("⚠️  A palavra '%s' já existe no banco de palavras!", word)
		return false
	}

	s.Words = append(s.Words, word)
	s.CustomWords = append(s.CustomWords, word)
	s.CustomWordsCount = len(s.CustomWords)
	s.InputBuffer = s.InputBuffer[:0]
	s.Mode = ModePlaying
	s.StatusMessage = fmt.Sprintf("✅ Palavra '%s' salva! Agora o banco tem %d palavras para sorteio.", word, len(s.Words))
	return true
}

// Game implementa o contrato games.Game para o Jogo da Forca.
type Game struct {
	storagePath string
	words       []string
	bestScore   int
}

// New cria uma nova instância do Jogo da Forca usando o arquivo padrão de persistência de palavras.
func New() *Game {
	return NewWithStorage(DefaultStoragePath)
}

// NewWithStorage cria uma instância do Jogo da Forca apontando para um caminho de arquivo específico.
func NewWithStorage(storagePath string) *Game {
	words := slices.Clone(DefaultWords)
	if custom, err := LoadCustomWords(storagePath); err == nil {
		for _, w := range custom {
			if !slices.Contains(words, w) {
				words = append(words, w)
			}
		}
	}
	return &Game{
		storagePath: storagePath,
		words:       words,
	}
}

// Meta retorna os metadados do Jogo da Forca para o menu da estação.
func (g *Game) Meta() games.Metadata {
	return games.Metadata{
		ID:          "forca",
		Title:       "🔤 Jogo da Forca (Hangman)",
		Description: "Descubra a palavra secreta sorteada aleatoriamente letra por letra! Inclui 10 palavras prontas e salva em disco as palavras que você adicionar.",
		Controls:    "A-Z: Chutar Letra | TAB: Adicionar Palavra | R/Enter: Nova Palavra | ESC: Menu",
	}
}

// Run executa o loop interativo do Jogo da Forca.
func (g *Game) Run(ctx context.Context) (int, error) {
	screen := engine.NewScreen()
	input := engine.InputFromContext(ctx)
	input.Drain()

	// Recarrega palavras salvas em disco caso existam
	words := slices.Clone(DefaultWords)
	if custom, err := LoadCustomWords(g.storagePath); err == nil {
		for _, w := range custom {
			if !slices.Contains(words, w) {
				words = append(words, w)
			}
		}
		g.words = words
	}

	state := NewState(g.words)
	state.BestScore = g.bestScore
	sessionMaxScore := 0

	screen.ClearScreen()
	renderFrame(screen, state)

	for {
		select {
		case <-ctx.Done():
			return max(state.Score, sessionMaxScore), ctx.Err()

		case key := <-input.Events():
			if state.Mode == ModeAddingWord {
				switch {
				case key == engine.KeyEsc:
					state.CancelAddWordMode()
				case key == engine.KeyBackspace:
					state.BackspaceWordLetter()
				case key == engine.KeyEnter:
					if state.ConfirmAddWord() {
						g.words = slices.Clone(state.Words)
						if err := SaveCustomWords(g.storagePath, state.CustomWords); err != nil {
							state.StatusMessage = fmt.Sprintf("⚠️  Palavra adicionada na sessão, mas falha ao salvar arquivo: %v", err)
						}
					}
				case key.IsLetter():
					state.TypeWordLetter(key.UpperRune())
				}
				renderFrame(screen, state)
				continue
			}

			// Modo de Jogo Normal (ModePlaying)
			switch {
			case key == engine.KeyEsc || key == engine.KeyQuit || (state.RoundOver && key == engine.KeyQ):
				if state.BestScore > g.bestScore {
					g.bestScore = state.BestScore
				}
				return max(state.Score, sessionMaxScore), nil

			case key == engine.KeyTab:
				state.EnterAddWordMode()
				renderFrame(screen, state)

			case state.RoundOver && (key == engine.KeyEnter || key == engine.KeySpace || key == engine.KeyRestart):
				if state.Score > sessionMaxScore {
					sessionMaxScore = state.Score
				}
				state.StartNewRound()
				renderFrame(screen, state)

			case !state.RoundOver && key.IsLetter():
				state.GuessLetter(key.UpperRune())
				if state.Score > sessionMaxScore {
					sessionMaxScore = state.Score
				}
				if state.BestScore > g.bestScore {
					g.bestScore = state.BestScore
				}
				renderFrame(screen, state)
			}
		}
	}
}

func renderFrame(screen *engine.Screen, s *State) {
	screen.BeginFrame()

	screen.WriteLine("")
	screen.WriteLine(fmt.Sprintf("  %s%s🔤 JOGO DA FORCA — ESTAÇÃO DE JOGOS CLI%s", engine.Bold, engine.FgBrightCyan, engine.Reset))
	screen.WriteLine(fmt.Sprintf("  %sPONTOS:%s %s%04d%s  │  %sMELHOR:%s %s%04d%s  │  %sACERTOS:%s %d  │  %sBANCO:%s %d palavras (%d salvas pelo usuário)",
		engine.Bold, engine.Reset,
		engine.FgBrightYellow, s.Score, engine.Reset,
		engine.Bold, engine.Reset,
		engine.FgBrightGreen, s.BestScore, engine.Reset,
		engine.Bold, engine.Reset, s.WordsSolved,
		engine.Bold, engine.Reset, len(s.Words), s.CustomWordsCount,
	))
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", 78) + engine.Reset)

	// Desenho da Forca
	gallows := buildGallowsLines(s.Errors)
	for _, line := range gallows {
		screen.WriteLine("  " + line)
	}

	screen.WriteLine("")

	// Exibição da Palavra Secreta
	wordColor := engine.FgBrightWhite
	if s.RoundOver {
		if s.Won {
			wordColor = engine.FgBrightGreen
		} else {
			wordColor = engine.FgBrightRed
		}
	}
	screen.WriteLine(fmt.Sprintf("  %sPALAVRA (%d letras):%s   %s%s%s%s",
		engine.Bold, len(s.TargetWord), engine.Reset,
		engine.Bold, wordColor, s.MaskedWord(), engine.Reset,
	))
	screen.WriteLine(fmt.Sprintf("  %sTENTATIVAS RESTANTES:%s %s%d / %d%s",
		engine.Bold, engine.Reset,
		errorCountColor(s.Errors), MaxErrors-s.Errors, MaxErrors, engine.Reset,
	))
	screen.WriteLine("")

	// Teclado Virtual de Letras (A-Z)
	screen.WriteLine(fmt.Sprintf("  %sLETRAS:%s", engine.Bold, engine.Reset))
	screen.WriteLine("  " + formatAlphabetRow("ABCDEFGHIJKLM", s.Guessed, s.TargetWord))
	screen.WriteLine("  " + formatAlphabetRow("NOPQRSTUVWXYZ", s.Guessed, s.TargetWord))
	screen.WriteLine("")

	// Área de Adição de Nova Palavra ou Mensagem de Status
	screen.WriteLine("  " + engine.Dim + strings.Repeat("─", 78) + engine.Reset)
	if s.Mode == ModeAddingWord {
		typed := string(s.InputBuffer)
		screen.WriteLine(fmt.Sprintf("  %s%s➕ ADICIONAR NOVA PALAVRA AO BANCO (SALVA EM DISCO):%s", engine.Bold, engine.FgBrightYellow, engine.Reset))
		screen.WriteLine(fmt.Sprintf("     Nova Palavra: %s[%s█]%s (%d/%d letras)",
			engine.FgBrightGreen, typed, engine.Reset, len(s.InputBuffer), MaxWordLength))
		screen.WriteLine(fmt.Sprintf("  %s%s%s", engine.FgCyan, s.StatusMessage, engine.Reset))
		screen.WriteLine(fmt.Sprintf("  %sControles: [A-Z] Digitar  •  [Backspace] Apagar  •  [ENTER] Salvar  •  [ESC] Cancelar%s", engine.Dim, engine.Reset))
	} else {
		statusColor := engine.FgBrightWhite
		if s.RoundOver && s.Won {
			statusColor = engine.FgBrightGreen
		} else if s.RoundOver && !s.Won {
			statusColor = engine.FgBrightRed
		} else {
			statusColor = engine.FgBrightYellow
		}
		screen.WriteLine(fmt.Sprintf("  %s%s%s", statusColor, s.StatusMessage, engine.Reset))

		if s.RoundOver {
			screen.WriteLine(fmt.Sprintf("  %sPressione [ENTER / ESPAÇO / R] para sortear outra palavra, [TAB] para adicionar palavra ou [ESC / Q] para o Menu.%s",
				engine.FgBrightCyan, engine.Reset))
		} else {
			screen.WriteLine(fmt.Sprintf("  %sPressione [TAB] a qualquer momento para adicionar e salvar novas palavras!%s",
				engine.Dim, engine.Reset))
		}
		screen.WriteLine(fmt.Sprintf("  %sControles: [A-Z] Chutar Letra  •  [TAB] +Nova Palavra  •  [ESC] Voltar ao Menu%s", engine.Dim, engine.Reset))
	}

	screen.Flush()
}

func buildGallowsLines(errors int) []string {
	head := " "
	body := " "
	leftArm := " "
	rightArm := " "
	leftLeg := " "
	rightLeg := " "

	if errors >= 1 {
		head = engine.FgBrightYellow + "O" + engine.Reset
	}
	if errors >= 2 {
		body = engine.FgBrightYellow + "│" + engine.Reset
	}
	if errors >= 3 {
		leftArm = engine.FgBrightYellow + "/" + engine.Reset
	}
	if errors >= 4 {
		rightArm = engine.FgBrightYellow + "\\" + engine.Reset
	}
	if errors >= 5 {
		leftLeg = engine.FgBrightRed + "/" + engine.Reset
	}
	if errors >= 6 {
		rightLeg = engine.FgBrightRed + "\\" + engine.Reset
	}

	c := engine.FgCyan
	r := engine.Reset
	return []string{
		fmt.Sprintf("%s╔═══════╗%s", c, r),
		fmt.Sprintf("%s║       │%s", c, r),
		fmt.Sprintf("%s║%s       %s", c, r, head),
		fmt.Sprintf("%s║%s      %s%s%s", c, r, leftArm, body, rightArm),
		fmt.Sprintf("%s║%s      %s %s", c, r, leftLeg, rightLeg),
		fmt.Sprintf("%s║%s", c, r),
		fmt.Sprintf("%s╩═══════════%s", c, r),
	}
}

func formatAlphabetRow(letters string, guessed map[rune]bool, target string) string {
	var b strings.Builder
	for _, ch := range letters {
		if !guessed[ch] {
			b.WriteString(fmt.Sprintf(" [%c] ", ch))
		} else if strings.ContainsRune(target, ch) {
			b.WriteString(fmt.Sprintf(" %s%s[%c]%s ", engine.Bold, engine.FgBrightGreen, ch, engine.Reset))
		} else {
			b.WriteString(fmt.Sprintf(" %s%s[%c]%s ", engine.Dim, engine.FgRed, ch, engine.Reset))
		}
	}
	return b.String()
}

func errorCountColor(errors int) string {
	switch {
	case errors <= 2:
		return engine.FgBrightGreen
	case errors <= 4:
		return engine.FgBrightYellow
	default:
		return engine.FgBrightRed
	}
}
