# animeGo-cli: guia para o Claude

Leia este arquivo no início de toda sessão. Ele guarda o modo de trabalho combinado e o andamento do projeto.

O enunciado completo está em [DESAFIO.md](DESAFIO.md). As decisões já tomadas estão em [ARCHITECTURE.md](ARCHITECTURE.md).

## Modo de trabalho

- **Idioma:** português.
- **Papel:** professor e apoiador, não executor. O objetivo do usuário é aprender os conceitos (Go, Clean Architecture, SOLID) e ficar autônomo.
- **Para cada decisão, explicar:** por que usamos, qual outro caminho existia, e o argumento para o caminho escolhido. Fechar com uma heurística reutilizável ("como reconhecer isso sozinho").
- **Quem escreve o código:** o usuário. O Claude explica e revisa. Exceção: infraestrutura longa (ex.: o teste de arquitetura), em que o Claude escreve e o usuário estende ou quebra de propósito.
- **Formato das respostas:** primeira linha é uma ação; passos numerados; estado atual ("passo X de Y") em toda resposta; uma decisão por vez; terminar com uma próxima ação concreta. O usuário costuma ativar a skill `i-have-adhd` para isso.
- **Antes de afirmar que algo funciona:** simular numa cópia em scratchpad ou rodar no repositório. Testar também o caso de falha.

## Andamento (seção 16 do DESAFIO)

Um PR por etapa, com testes e com o checklist do README atualizado.

| # | Etapa | Estado |
| --- | --- | --- |
| 0 | Fundação | ✅ PR #5 mesclado (ver pendências abaixo) |
| 1 | Domínio | ⏭️ próxima |
| 2 | Portas e casos de uso | ⬜ |
| 3 | Configuração e armazenamento local | ⬜ |
| 4 | Adapter AnimeFire | ⬜ |
| 5 | Adapter mpv | ⬜ |
| 6 | Fatia vertical (`animego play "<busca>" --episode N`) | ⬜ |
| 7 | Adapter AniList | ⬜ |
| 8 | Adapter AniSkip | ⬜ |
| 9 | Goyabu + agregador | ⬜ |
| 10 | TUI | ⬜ |
| 11 | Release | ⬜ |

### Pendências

- `cmd/animego/main.go` foi para a `main` como `package animego`; a correção para `package main` está local, sem commit.
- O checklist do README não tem um item para a Fundação e ainda marca o domínio como pronto, o que não é verdade pelo DESAFIO.

### Etapa 1: o que existe hoje

`internal/domain` tem só structs com `int` e tags JSON (`Anime`, `Season`, `Episode`, `Progress`). Faltam: construtores com validação, reidratação separada da criação, `time.Duration`, `ListEntry`, `SkipInterval`, `Stream`, os erros da seção 6 e a decisão sobre `Season` (seção 6.2). Ordem combinada: começar pelos erros do domínio.

## O que já está montado

- **Módulo:** `github.com/minimundosbr/animeGo-cli`, Go 1.27.1.
- **Makefile:** `build`, `test`, `race`, `vet`, `fmt`.
- **CI:** `.github/workflows/ci.yml` roda `gofmt`, `make vet` e `make race` em push na `main` e em PRs.
- **Regra de dependência:** `internal/archtest/deps_test.go` lê os imports com `go/parser` e compara com a tabela `rules`. Hoje cobre `domain` e `usecase`. Usa `go/parser` e não `go list` porque o `go list` roda em outro processo e o cache de testes não percebia mudanças (falso `ok (cached)`). Ao criar `adapter`, `ui` e `config`, adicionar a regra na tabela e provar com um import proibido.

## Ambiente do usuário

- Windows 10, VS Code, PowerShell.
- `make` instalado via `winget install ezwinports.make`. Depois de instalar ferramentas, reiniciar o VS Code inteiro (o terminal herda o PATH dele).
- `go test -race` não roda localmente (sem `gcc`); o CI em Linux cobre.
- Git com `core.autocrlf=true`; o `.gitattributes` (`* text=auto eol=lf`) corrige. Arquivos novos criados no VS Code podem nascer em CRLF: `make fmt` resolve os `.go`.
- Risco conhecido para as Etapas 5 e 6: no Windows o mpv usa *named pipe* no `--input-ipc-server`, não socket Unix. Só Linux é obrigatório em execução, mas o usuário desenvolve no Windows.

## Convenções de git

- Uma branch por etapa (`feature/<nome>`), a partir da `main` atualizada.
- Commits pequenos, mensagem em inglês com prefixo: `[feat]`, `[fix]`, `[docs]`, `[test]`, `[chore]`. Formato: `[prefixo] - Descrição no gerúndio`.
- `git add` com caminhos explícitos, nunca `git add .`.
- Depois do merge, a branch pode ser apagada (o PR guarda o registro).
