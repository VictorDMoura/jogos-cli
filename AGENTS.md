# AGENTS.md — Diretrizes do Projeto `jogos-cli`

## 1. Visão Geral do Projeto
O **`jogos-cli`** é uma **Estação de Jogos de Terminal (CLI Arcade Station)** desenvolvida em **Go 1.27**. O objetivo do projeto é fornecer um hub unificado no terminal onde o usuário pode navegar por um menu interativo, escolher e jogar diversos jogos clássicos e autorais renderizados inteiramente no console.

---

## 2. Regra de Ouro: 100% Go Standard Library (Zero Dependências)
> **PROIBIDO** o uso de bibliotecas de terceiros (`go.mod` não deve conter nenhuma diretiva `require` externa, como `bubbletea`, `tcell`, `termbox`, `lipgloss` ou `golang.org/x/...`).

Toda funcionalidade — desde manipulação de terminal, leitura de teclado em modo bruto (*raw mode*), renderização ANSI, loop de jogo, persistência de recordes (*high scores*) e testes — **deve** ser construída exclusivamente com pacotes da biblioteca padrão do Go 1.27 (`os`, `syscall`, `fmt`, `strings`, `bytes`, `time`, `math/rand/v2`, `encoding/json`, `io`, `sync`, `context`, etc.).

---

## 3. Arquitetura e Estrutura de Diretórios

```text
jogos-cli/
├── AGENTS.md                 # Diretrizes e regras para agentes e desenvolvedores
├── go.mod                    # Módulo Go 1.27 (sem dependências externas)
├── cmd/
│   └── arcade/
│       └── main.go           # Ponto de entrada da Estação de Jogos
└── internal/
    ├── engine/               # Motor compartilhado de terminal (100% stdlib)
    │   ├── terminal.go       # Controle de Raw Mode (Windows Console API / POSIX termios via syscall)
    │   ├── screen.go         # Double-buffer de tela, códigos ANSI, limpeza sem flicker
    │   ├── input.go          # Leitura assíncrona de teclas (setas, WASD, ESC, Enter)
    │   └── loop.go           # Game loop com tick rate configurável (time.Ticker)
    ├── station/              # Interface do Hub / Menu Principal da Estação
    │   ├── launcher.go       # Menu de seleção de jogos, placares e configurações
    │   └── scores.go         # Persistência local de High Scores (encoding/json)
    └── games/                # Catálogo de jogos implementando o contrato comum
        ├── game.go           # Interface comum `Game` e registro do catálogo
        ├── snake/            # Jogo da Cobrinha (exemplo de jogo em tempo real)
        ├── tetris/           # Blocos em queda (exemplo de jogo baseado em tick/grid)
        ├── pong/             # Clássico arcade
        └── ...               # Novos jogos plugáveis
```

---

## 4. Contrato de Jogos (`Game` Interface)
Para que a estação seja modular e qualquer agente possa criar um jogo novo sem acoplamento com o menu principal, todo jogo deve implementar uma interface padronizada em `internal/games/game.go`:

```go
package games

import "context"

// Metadata descreve as informações exibidas no menu da estação.
type Metadata struct {
	ID          string
	Title       string
	Description string
	Controls    string
}

// Game define o contrato que todo jogo da estação deve cumprir.
type Game interface {
	Meta() Metadata
	// Run executa o jogo assumindo o controle do terminal e retorna a pontuação final.
	Run(ctx context.Context) (score int, err error)
}
```

---

## 5. Diretrizes Técnicas de Terminal (Stdlib Only)

### 5.1. Entrada de Teclado (Raw Mode Multiplataforma)
Como apenas a *standard library* é permitida:
- **Windows (`terminal_windows.go`)**: Utilizar `syscall.NewLazyDLL("kernel32.dll")` com `GetConsoleMode` e `SetConsoleMode` para desabilitar `ENABLE_LINE_INPUT` e `ENABLE_ECHO_INPUT`, além de habilitar `ENABLE_VIRTUAL_TERMINAL_PROCESSING` para suporte completo a ANSI Escape Codes.
- **Linux/macOS (`terminal_unix.go`)**: Utilizar `syscall.Syscall(syscall.SYS_IOCTL, ...)` para manipular a struct `syscall.Termios` (desativando `ICANON` e `ECHO`).
- **Restauração Segura**: Sempre garantir um `defer term.Restore()` no `main` e captura de sinais (`os/signal`) para nunca deixar o terminal do usuário desconfigurado ou com o cursor oculto após encerramento.

### 5.2. Renderização Sem Flicker (Double Buffering)
- **Proibido** usar comandos de shell externos (`cls` ou `clear`) a cada frame.
- Construir todo o frame em memória usando `strings.Builder` ou `bytes.Buffer` ou um grid 2D de células (`rune` + cor ANSI).
- Posicionar o cursor no topo (`\x1b[H`) e fazer uma única escrita (`os.Stdout.Write`) por frame.
- Usar tela alternativa do terminal (`\x1b[?1049h` ao entrar e `\x1b[?1049l` ao sair) e ocultar o cursor (`\x1b[?25l` / `\x1b[?25h`).

### 5.3. Idiomas do Go 1.27
- Utilizar `math/rand/v2` para geração de números aleatórios.
- Utilizar `min`, `max`, `clear` nativos e iteradores (`iter`) ou `range` sobre inteiros (`for i := range n`) sempre que deixarem o código mais limpo.
- Manter lógica do estado do jogo desacoplada da renderização ANSI para permitir testes unitários rápidos com `go test ./...`.

---

## 6. Fluxo de Trabalho para Agentes
1. **Verificar Dependências**: Antes de concluir qualquer tarefa, garanta que `go.mod` contém **apenas** a declaração `module` e `go 1.27`, sem blocos `require`.
2. **Compilação e Testes**:
   - `go build ./...` deve compilar sem erros ou warnings.
   - `go test ./...` deve validar a lógica dos jogos e da engine.
3. **Como Adicionar um Novo Jogo**:
   - Criar o pacote em `internal/games/<nome_do_jogo>`.
   - Implementar a interface `games.Game`.
   - Adicionar testes unitários para as regras de negócio do jogo (`<nome_do_jogo>_test.go`).
   - Registrar o jogo na lista oficial do launcher da estação.
4. **Experiência do Usuário (UX)**:
   - Todo jogo deve exibir controles claros na lateral ou rodapé.
   - A tecla `ESC` ou `Q` deve sempre permitir pausar ou voltar ao menu principal da estação sem encerrar o programa abruptamente.
   - A tecla `R` em telas de *Game Over* deve permitir reiniciar a partida rapidamente.
