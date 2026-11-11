# 🕹️ jogos-cli — Estação de Jogos no Terminal

Uma estação de jogos interativa executada inteiramente no terminal (*CLI Arcade Station*), construída em **Go 1.27+** utilizando **exclusivamente a biblioteca padrão (100% Go Standard Library — Zero Dependências Externas)**.

---

## ✨ Destaques

- **Zero Dependências (`go.mod` limpo)**: Sem `bubbletea`, `tcell`, `termbox` ou `golang.org/x/...`. Todo o controle de terminal, leitura de teclado em tempo real e renderização ANSI foram escritos do zero com pacotes nativos do Go (`syscall`, `os`, `strings`, `time`, `math/rand/v2`, `encoding/json`, `unicode`).
- **Renderização Sem Flicker**: Motor gráfico ANSI com *Double Buffering* em memória e suporte a *Alternate Screen Buffer* (`\x1b[?1049h`), preservando o histórico do seu terminal ao sair.
- **Multiplataforma**: Suporte nativo a *Raw Mode* no Windows (`kernel32.dll`), Linux e macOS (`termios`/`ioctl`).
- **Hub de Pontuação da Sessão**: Menu principal interativo que registra a última pontuação, o recorde de cada jogo e a pontuação total acumulada na sessão.

---

## 🎮 Jogos Disponíveis

### 1. 🐍 Snake (Jogo da Cobrinha)
Guie a cobrinha pelo tabuleiro, colete as maçãs para crescer e somar pontos sem colidir com as bordas ou com o próprio corpo. A velocidade aumenta progressivamente conforme sua pontuação sobe!
- **Controles**:
  - `Setas` ou `W A S D`: Mover a cobrinha
  - `P` ou `Espaço`: Pausar / Continuar
  - `R`: Reiniciar partida rapidamente
  - `ESC` ou `Q`: Voltar ao Menu Principal

### 2. 🔤 Jogo da Forca (Hangman)
Descubra a palavra secreta sorteada aleatoriamente letra por letra antes que as 6 tentativas acabem!
- **Banco de Palavras Expansível e Persistente**:
  - Já vem com **10 palavras pré-prontas**.
  - Pressione `TAB` a qualquer momento para **cadastrar novas palavras**. As palavras adicionadas ficam salvas automaticamente em `data/forca_words.json` e entram no sorteio aleatório mesmo após fechar o programa.
- **Controles**:
  - `A - Z`: Chutar uma letra (normaliza acentos e `ç` automaticamente)
  - `TAB`: Abrir modo de adição de nova palavra (`Enter` salva, `Backspace` corrige, `ESC` cancela)
  - `Enter` / `Espaço` / `R`: Sortear nova palavra ao fim da rodada
  - `ESC`: Voltar ao Menu Principal

---

## 🚀 Como Executar

### Pré-requisitos
- **Go 1.27** ou superior instalado.

### Rodar diretamente a Estação
```bash
go run ./cmd/arcade
```

### Compilar o executável
```bash
# Windows
go build -o arcade.exe ./cmd/arcade

# Linux / macOS
go build -o arcade ./cmd/arcade
```

### Executar os Testes Automatizados
```bash
go test -v ./...
```

---

## 📂 Estrutura do Projeto

```text
jogos-cli/
├── AGENTS.md                 # Constituição arquitetural e guia perene do projeto
├── README.md                 # Documentação de uso do projeto
├── go.mod                    # Módulo Go 1.27 (zero dependências externas)
├── data/                     # Dados persistidos localmente em JSON (ex.: palavras da Forca)
├── cmd/
│   └── arcade/
│       └── main.go           # Ponto de entrada: inicializa Raw Mode, sinais de SO e Launcher
└── internal/
    ├── engine/               # Motor de terminal compartilhado (Raw Mode, Screen ANSI e Input)
    ├── station/              # Menu principal interativo e placar da sessão
    └── games/                # Catálogo modular de jogos
        ├── game.go           # Interface universal `games.Game`
        ├── snake/            # Implementação e testes do Jogo da Cobrinha
        └── forca/            # Implementação e testes do Jogo da Forca
```

---

## 🧩 Como Adicionar Novos Jogos

O projeto possui uma arquitetura plugável baseada no contrato `games.Game`. Consulte o arquivo [AGENTS.md](./AGENTS.md) para ver todas as regras de arquitetura, padrões de estado puro (`State`), renderização sem flicker e compartilhamento de entrada.
