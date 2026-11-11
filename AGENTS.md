# AGENTS.md — Constituição e Guia Perene do Projeto `jogos-cli`

> Este documento define os **princípios imutáveis, contratos de arquitetura e padrões de engenharia** do `jogos-cli`. Ele foi desenhado para ser **atemporal (à prova de validade)**: independentemente de quantos jogos ou recursos sejam adicionados no futuro, estas regras permanecem válidas.

---

## 1. Propósito do Projeto
O **`jogos-cli`** é uma **Estação de Jogos de Terminal (CLI Arcade Station)** escrita em **Go (1.27+)**. Seu objetivo é oferecer um hub interativo executável diretamente no console, capaz de hospedar múltiplos jogos desacoplados sob um único motor de terminal.

---

## 2. Regra de Ouro Imutável: 100% Go Standard Library
- **Zero Dependências Externas**: O arquivo `go.mod` **jamais** deve conter diretivas `require` (proibido `bubbletea`, `tcell`, `termbox`, `lipgloss`, `golang.org/x/...` ou qualquer pacote externo).
- Toda solução — controle de terminal em modo bruto (*Raw Mode*), leitura de teclado, renderização ANSI, *game loop*, estado, placares e testes — deve utilizar exclusivamente pacotes nativos da biblioteca padrão do Go (`os`, `syscall`, `fmt`, `strings`, `time`, `math/rand/v2`, `context`, `slices`, `io`, `sync`, etc.).

---

## 3. Arquitetura Baseada em Contratos (Invariante)

A estrutura do projeto é dividida em 4 camadas com responsabilidades estritas. Novos recursos devem respeitar estas fronteiras:

```text
jogos-cli/
├── AGENTS.md                 # Constituição perene do projeto
├── go.mod                    # Módulo Go (sem bloco require)
├── cmd/
│   └── arcade/
│       └── main.go           # Bootstrap: inicializa Raw Mode, sinais de SO, Input e Launcher
└── internal/
    ├── engine/               # Motor de terminal compartilhado (I/O, ANSI e Teclado)
    │   ├── terminal*.go      # Raw Mode multiplataforma (Windows Console API / POSIX termios)
    │   ├── screen.go         # Double-buffering ANSI sem flicker
    │   └── input.go          # Leitor único assíncrono de teclado compartilhado via Context
    ├── station/              # Hub / Menu Principal da Estação
    │   └── launcher.go       # Seleção de jogos, exibição de pontuação e ciclo de vida de sessão
    └── games/                # Catálogo modular de jogos
        ├── game.go           # Contrato universal `Game` e `Metadata`
        └── <id_do_jogo>/     # Cada jogo vive em seu próprio subpacote isolado
```

### Regras de Acoplamento entre Camadas
1. **`internal/engine`** não conhece nenhum jogo nem o menu da estação.
2. **`internal/games/<id>`** conhece apenas `internal/engine` e `internal/games`. Um jogo **nunca** importa outro jogo nem importa `internal/station`.
3. **`internal/station`** conhece apenas a interface `games.Game` (nunca a implementação interna de um jogo específico).
4. **`cmd/arcade/main.go`** é o único ponto que instancia o catálogo de jogos e os injeta no `station.Launcher`.

---

## 4. Contratos Fundamentais da Engine e dos Jogos

### 4.1. Contrato Universal de Jogo (`internal/games/game.go`)
Todo jogo, presente ou futuro, deve implementar:

```go
type Metadata struct {
	ID          string
	Title       string
	Description string
	Controls    string
}

type Game interface {
	Meta() Metadata
	Run(ctx context.Context) (score int, err error)
}
```

- **`Meta()`**: Retorna textos prontos para exibição automática no menu principal.
- **`Run(ctx)`**: Assume o loop de tela/entrada enquanto a partida dura e **sempre** retorna a pontuação obtida na sessão (`score`) quando o jogador volta ao menu (`ESC`/`Q`) ou quando o contexto é cancelado.

### 4.2. Compartilhamento de Entrada (`engine.InputReader`)
- **Regra Crítica de Concorrência**: Nunca crie múltiplas goroutines lendo diretamente de `os.Stdin`, pois elas disputarão bytes de entrada.
- Sempre obtenha o leitor compartilhado através de `input := engine.InputFromContext(ctx)` e chame `input.Drain()` ao iniciar uma nova tela ou partida para descartar teclas residuais no buffer.

### 4.3. Renderização Sem Flicker (`engine.Screen`)
- **Proibido** executar comandos externos (`cls` / `clear`) ou imprimir linhas soltas com `fmt.Println` durante o loop.
- Fluxo obrigatório de renderização a cada frame:
  1. `screen.BeginFrame()` (posiciona o cursor em `\x1b[H` na memória).
  2. `screen.WriteLine(...)` para cada linha do quadro (aplica `\x1b[K\r\n` automaticamente).
  3. `screen.Flush()` (escreve o frame inteiro em uma única operação de I/O e limpa resíduos abaixo com `\x1b[J`).
- Ao alternar entre o Menu da Estação e um Jogo, chame `screen.ClearScreen()` uma única vez antes do primeiro frame.

### 4.4. Segurança do Terminal (`engine.Terminal`)
- O modo bruto (`EnableRawMode`) ativa o *Alternate Screen Buffer* (`\x1b[?1049h`), oculta o cursor (`\x1b[?25l`) e configura UTF-8/ANSI.
- O `main` deve sempre garantir `defer term.Restore()` e interceptar sinais do sistema operacional (`os.Interrupt`, `syscall.SIGTERM`) via `signal.NotifyContext` para que o terminal do usuário nunca fique travado ou sem cursor ao sair.

---

## 5. Padrão de Design Interno de Cada Jogo

Para manter o código sustentável e testável em qualquer época, todo jogo em `internal/games/<id>` deve separar:

1. **Estado e Regras de Negócio (`State`)**:
   - Struct pura, sem dependência de terminal, ANSI ou `os.Stdin`.
   - Métodos determinísticos ou baseados em `math/rand/v2` (ex.: `NewState`, `Reset`, `Step`, mudança de direção/ação).
   - Testado diretamente por testes unitários em `<id>_test.go`.
2. **Loop Interativo e Renderização (`Game.Run` e `renderFrame`)**:
   - Conecta `time.Ticker` (para jogos em tempo real) ou eventos de `input.Events()` (para jogos por turno) ao `State`.
   - Traduz o `State` em linhas formatadas no `engine.Screen`.

### Convenções de UX Obrigatórias em Todos os Jogos
- **Movimentação / Ação**: Suportar tanto **Setas direcionais** quanto **WASD**, além de `Enter`/`Espaço` para confirmação/ação.
- **Pausa**: Tecla `P` (ou `Espaço` quando não usada como ação principal) deve pausar/retomar jogos em tempo real.
- **Reinício Rápido**: Tecla `R` deve permitir reiniciar a partida imediatamente (especialmente na tela de *Game Over*).
- **Saída Segura para o Menu**: Teclas `ESC` e `Q` devem encerrar a partida atual e retornar ao menu principal da estação devolvendo a pontuação alcançada.
- **HUD Clara**: Sempre exibir pontuação atual, controles resumidos no rodapé e feedback visual claro para estados de *Pausa*, *Vitória* ou *Fim de Jogo*.

---

## 6. Protocolo Perene para Adicionar ou Modificar Funcionalidades

Sempre que um agente ou desenvolvedor atuar neste repositório, deve seguir este checklist:

1. **Criar/Atualizar Pacote**:
   - Se for um novo jogo, criar `internal/games/<id>` implementando `games.Game` + testes `<id>_test.go` e registrá-lo na fatia `catalog` em `cmd/arcade/main.go`.
2. **Validar Pureza do Módulo**:
   - Confirmar que `go.mod` permanece sem nenhuma dependência externa (`require`).
3. **Executar Verificação Automatizada**:
   - `go build ./...` (deve compilar sem erros em qualquer plataforma suportada).
   - `go test ./...` (todos os testes de engine e jogos devem passar).
