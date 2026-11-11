package snake

import "testing"

func TestSnakeInitialState(t *testing.T) {
	s := NewState(20, 15)
	if len(s.Body) != 3 {
		t.Fatalf("esperava tamanho inicial 3, obteve %d", len(s.Body))
	}
	if s.Score != 0 {
		t.Fatalf("esperava pontuação inicial 0, obteve %d", s.Score)
	}
	if s.GameOver {
		t.Fatal("jogo não deveria iniciar em GameOver")
	}
}

func TestSnakeMoveAndEatFood(t *testing.T) {
	s := NewState(20, 15)
	head := s.Body[0]
	// Posiciona a comida exatamente à direita da cabeça (direção inicial é DirRight)
	s.Food = Point{X: head.X + 1, Y: head.Y}

	s.Step()

	if s.GameOver {
		t.Fatal("não deveria dar GameOver ao comer comida válida")
	}
	if len(s.Body) != 4 {
		t.Fatalf("esperava tamanho 4 após comer, obteve %d", len(s.Body))
	}
	if s.Score != PointsPerFood {
		t.Fatalf("esperava pontuação %d, obteve %d", PointsPerFood, s.Score)
	}
	if s.BestScore != PointsPerFood {
		t.Fatalf("esperava BestScore %d, obteve %d", PointsPerFood, s.BestScore)
	}
}

func TestSnakePreventOppositeDirection(t *testing.T) {
	s := NewState(20, 15)
	// Direção inicial é DirRight; tentar mudar para DirLeft deve ser ignorado
	s.ChangeDirection(DirLeft)
	s.Step()

	if s.Dir != DirRight {
		t.Fatalf("esperava manter direção DirRight ao tentar inverter para DirLeft, obteve %v", s.Dir)
	}
}

func TestSnakeWallCollision(t *testing.T) {
	s := NewState(10, 10)
	s.Body = []Point{{X: 9, Y: 5}, {X: 8, Y: 5}, {X: 7, Y: 5}}
	s.Dir = DirRight

	s.Step()

	if !s.GameOver {
		t.Fatal("esperava GameOver ao colidir com a parede direita")
	}
}

func TestSnakeSelfCollision(t *testing.T) {
	s := NewState(10, 10)
	// Cobra em formato de U prestes a bater no próprio corpo ao subir
	s.Body = []Point{
		{X: 5, Y: 5}, // cabeça
		{X: 4, Y: 5},
		{X: 4, Y: 4},
		{X: 5, Y: 4}, // segmento acima da cabeça
		{X: 6, Y: 4},
	}
	s.Dir = DirDown
	// Vira para cima (como estava indo para baixo, primeiro vira à direita ou testamos DirUp direto)
	s.Dir = DirLeft
	s.ChangeDirection(DirUp)

	s.Step()

	if !s.GameOver {
		t.Fatal("esperava GameOver ao colidir com o próprio corpo")
	}
}
