package ortografia

import "testing"

func TestInitialState(t *testing.T) {
	s := NewState(DefaultChallenges)
	if s.Lives != InitialLives {
		t.Fatalf("esperava %d vidas iniciais, obteve %d", InitialLives, s.Lives)
	}
	if s.CurrentLevel != 1 {
		t.Fatalf("esperava iniciar no Nível 1, obteve %d", s.CurrentLevel)
	}
	if s.Current.FullWord == "" {
		t.Fatal("esperava um desafio sorteado inicialmente")
	}
}

func TestAnswerAndLevelProgression(t *testing.T) {
	s := NewState(DefaultChallenges)

	// Acerta 5 seguidas para subir para o Nível 2
	for range 5 {
		s.Answer(s.Current.CorrectKey)
	}
	if s.Hits != 5 {
		t.Fatalf("esperava 5 acertos, obteve %d", s.Hits)
	}
	if s.CurrentLevel != 2 {
		t.Fatalf("esperava subir para o Nível 2 após 5 acertos, obteve %d", s.CurrentLevel)
	}

	// Acerta mais 5 (total 10) para subir para o Nível 3 (Concursos/ENEM)
	for range 5 {
		s.Answer(s.Current.CorrectKey)
	}
	if s.CurrentLevel != 3 {
		t.Fatalf("esperava subir para o Nível 3 após 10 acertos, obteve %d", s.CurrentLevel)
	}
	if s.Score <= 0 {
		t.Fatalf("esperava pontuação positiva, obteve %d", s.Score)
	}
}

func TestWrongAnswerAndFloorMiss(t *testing.T) {
	s := NewState(DefaultChallenges)
	initialLives := s.Lives

	// Erra a letra
	s.Answer(s.Current.WrongKey)
	if s.Lives != initialLives-1 {
		t.Fatalf("esperava perder 1 vida ao errar letra, obteve %d vidas", s.Lives)
	}
	if s.LastWasHit {
		t.Fatal("LastWasHit deveria ser falso após erro")
	}

	// Deixa a palavra cair até o chão
	for range BoardHeight {
		s.Step()
	}
	if s.Lives != initialLives-2 {
		t.Fatalf("esperava perder mais 1 vida ao deixar palavra cair no chão, obteve %d vidas", s.Lives)
	}
}
