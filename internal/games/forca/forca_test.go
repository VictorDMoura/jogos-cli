package forca

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestDefaultWordsCount(t *testing.T) {
	if len(DefaultWords) != 10 {
		t.Fatalf("esperava exatamente 10 palavras pré-prontas, obteve %d", len(DefaultWords))
	}

	s := NewState(DefaultWords)
	if len(s.Words) != 10 {
		t.Fatalf("esperava banco inicial com 10 palavras, obteve %d", len(s.Words))
	}
	if !slices.Contains(DefaultWords, s.TargetWord) {
		t.Fatalf("palavra sorteada %q não pertence às palavras iniciais", s.TargetWord)
	}
}

func TestGuessLettersAndWin(t *testing.T) {
	s := NewState([]string{"GO"})
	if s.TargetWord != "GO" {
		t.Fatalf("esperava palavra 'GO', obteve %q", s.TargetWord)
	}

	s.GuessLetter('G')
	if s.RoundOver {
		t.Fatal("não deveria encerrar com apenas 1 letra de 2")
	}
	if s.MaskedWord() != "G _" {
		t.Fatalf("esperava máscara 'G _', obteve %q", s.MaskedWord())
	}

	s.GuessLetter('O')
	if !s.RoundOver || !s.Won {
		t.Fatal("esperava vitória ao acertar todas as letras")
	}
	if s.Score <= 0 {
		t.Fatalf("esperava pontuação positiva ao vencer, obteve %d", s.Score)
	}
}

func TestMaxErrorsGameOver(t *testing.T) {
	s := NewState([]string{"GO"})
	wrongLetters := []rune{'A', 'B', 'C', 'D', 'E', 'F'}
	for _, l := range wrongLetters {
		s.GuessLetter(l)
	}

	if !s.RoundOver {
		t.Fatal("esperava fim da rodada após atingir o limite de erros")
	}
	if s.Won {
		t.Fatal("não deveria vencer após 6 erros")
	}
	if s.Errors != MaxErrors {
		t.Fatalf("esperava %d erros, obteve %d", MaxErrors, s.Errors)
	}
}

func TestAddCustomWordAndPersistToDisk(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "forca_words.json")

	// Simula primeira execução do jogo adicionando uma palavra e salvando em disco
	g1 := NewWithStorage(tempFile)
	s1 := NewState(g1.words)
	s1.EnterAddWordMode()

	for _, r := range "GOLANGX" {
		s1.TypeWordLetter(r)
	}
	s1.BackspaceWordLetter()

	if !s1.ConfirmAddWord() {
		t.Fatal("esperava confirmação bem-sucedida ao adicionar 'GOLANG'")
	}

	if err := SaveCustomWords(tempFile, s1.CustomWords); err != nil {
		t.Fatalf("erro ao salvar palavras em disco: %v", err)
	}

	// Simula fechar e abrir o jogo novamente lendo do mesmo arquivo
	g2 := NewWithStorage(tempFile)
	s2 := NewState(g2.words)

	if len(s2.Words) != 11 {
		t.Fatalf("esperava 11 palavras no banco após recarregar do disco, obteve %d", len(s2.Words))
	}
	if !slices.Contains(s2.Words, "GOLANG") {
		t.Fatal("esperava encontrar 'GOLANG' na lista de palavras recarregada do disco")
	}
	if s2.CustomWordsCount != 1 {
		t.Fatalf("esperava CustomWordsCount = 1 após recarregar, obteve %d", s2.CustomWordsCount)
	}
}
