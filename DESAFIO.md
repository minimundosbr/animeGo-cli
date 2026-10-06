# Desafio Técnico — CLI de Anime com Sincronização de Progresso em Go

Implemente em **Go** uma aplicação de terminal (TUI) para buscar animes em fontes brasileiras, reproduzir episódios no **mpv** e sincronizar automaticamente o progresso com o **AniList**, seguindo **Clean Architecture** e **SOLID**.

> Este documento é o enunciado do desafio. O [README](README.md) descreve o produto e o andamento; aqui ficam os requisitos, as garantias, as restrições e os critérios de avaliação contra os quais a solução é medida.

## 1. Objetivo

A aplicação deve permitir que uma pessoa, sem sair do terminal:

1. busque um anime em uma ou mais fontes de vídeo;
2. escolha um episódio e assista no mpv, com pulo automático de abertura e encerramento;
3. tenha o progresso salvo localmente e enviado ao AniList quando o episódio for concluído;
4. avance para o próximo episódio sem procurar o anime de novo.

Demonstre que o núcleo do produto continua correto e intacto quando uma fonte cai, muda de domínio ou de HTML, quando a rede some no meio da sessão, quando o player é fechado abruptamente e quando o tracker limita requisições.

A avaliação considera **regra de dependência**, **modelagem de domínio**, **resiliência a falhas externas**, **correção da sincronização com o tracker**, **testabilidade** e **clareza das decisões documentadas**.

## 2. Ambiente de execução e falhas

As dependências externas deste projeto são, por natureza, **voláteis**. Prepare a solução para:

- uma fonte de vídeo fora do ar, lenta ou devolvendo HTML diferente do esperado;
- mudança de domínio de uma fonte (ex.: `animefire.xyz` → `animefire.abc`);
- uma busca que devolve o mesmo anime em várias fontes, com títulos diferentes;
- um anime da fonte que não corresponde de forma óbvia a uma entrada do AniList;
- ausência de rede ao concluir um episódio;
- limite de requisições (`429`) ou token revogado (`401`) no AniList;
- AniSkip sem dados para o episódio, ou anime sem ID do MyAnimeList;
- mpv não instalado, fechado pelo usuário ou encerrado por erro;
- `Ctrl+C` ou `SIGTERM` durante a reprodução;
- o usuário já ter assistido episódios em outro dispositivo.

Nenhuma dessas situações pode **derrubar a aplicação**, **regredir o progresso no AniList**, **enviar a mesma atualização duas vezes**, **perder uma atualização já concluída** ou **deixar processos e arquivos órfãos**.

## 3. Stack

### Tecnologias obrigatórias

| Responsabilidade | Tecnologia |
| --- | --- |
| Linguagem | Go 1.22 ou superior; declare a versão em `go.mod` |
| Dependências | Go Modules, com `go.mod` e `go.sum` versionados |
| Interface | [bubbletea](https://github.com/charmbracelet/bubbletea), `bubbles` e `lipgloss` |
| HTTP | `net/http` da biblioteca padrão, com `http.Client` configurado (timeouts, `User-Agent`) |
| Parsing de HTML | `golang.org/x/net/html` ou `goquery`, à escolha |
| Player | [mpv](https://mpv.io/), controlado por [JSON IPC](https://mpv.io/manual/stable/#json-ipc) |
| Tracker | [AniList](https://docs.anilist.co/), API GraphQL com OAuth 2.0 |
| Pulo de OP/ED | [AniSkip](https://api.aniskip.com/api-docs), API pública |
| Configuração | JSON editável pelo usuário (`config.json`) |
| Logs | `log/slog`, gravando em arquivo |
| Testes | `testing`, `net/http/httptest` e `go test`, incluindo execução com `-race` |
| Integração contínua | GitHub Actions |

### Composição da aplicação

A injeção de dependências é **manual**, feita por construtores em um único **composition root** (`cmd/animego/main.go`). Frameworks de DI (Fx, Wire, dig) não são permitidos: o objetivo é deixar explícito, em código legível, quem depende de quem.

O `main.go` deve:

- carregar e validar a configuração antes de construir qualquer adapter;
- instanciar os adapters, injetá-los nos casos de uso e os casos de uso na TUI;
- criar um `context.Context` raiz cancelado por `SIGINT`/`SIGTERM` (`signal.NotifyContext`);
- garantir, no encerramento, que o mpv foi finalizado, o socket IPC removido, o progresso salvo e o terminal restaurado.

### Build

O binário deve compilar com `CGO_ENABLED=0` para Linux, macOS e Windows (`amd64` e `arm64`), mesmo que, em tempo de execução, apenas Linux seja obrigatório (seção 15).

## 4. Arquitetura

### Camadas e regra de dependência

```
cmd/animego/          # composition root — único lugar que conhece todas as camadas
internal/
├── domain/           # entidades e value objects — só biblioteca padrão
├── usecase/          # casos de uso + portas (interfaces)
├── adapter/
│   ├── source/       # animefire, goyabu, aggregator
│   ├── player/       # mpv
│   ├── tracker/      # anilist
│   ├── skip/         # aniskip
│   └── storage/      # arquivos locais: token, progresso, fila de sincronização, cache
├── config/           # leitura, validação e valores padrão do config.json
└── ui/               # TUI (bubbletea)
```

| Pacote | Pode importar | Não pode importar |
| --- | --- | --- |
| `domain` | biblioteca padrão (sem `net/*`, `os/exec`, `database/*`) | qualquer outro pacote interno ou de terceiros |
| `usecase` | `domain`, biblioteca padrão | `adapter/*`, `ui`, `config`, bubbletea, clientes HTTP |
| `adapter/*` | `domain`, `usecase` (para implementar as portas), bibliotecas externas | `ui`, outros adapters (exceto `aggregator`, que compõe fontes pela porta) |
| `ui` | `domain`, `usecase` | `adapter/*` |
| `cmd/animego` | todos | — |

A regra deve ser **verificada por teste automatizado** (por exemplo, com `go list -deps` ou `golang.org/x/tools/go/packages`), e não apenas pela convenção.

### Fluxo de referência: Reproduzir Episódio

```
TUI ──► PlayEpisode (caso de uso)
          1. AnimeSource   → resolve o link do vídeo (com fallback entre fontes)
          2. SkipProvider  → intervalos de OP/ED (falha não bloqueia)
          3. Player        → reproduz a partir da posição salva
          4. Progress      → atualizado durante a reprodução e no encerramento
          5. Tracker       → sincroniza ao concluir (ou enfileira, se offline)
```

O diagrama completo está no [documento de arquitetura](https://claude.ai/code/artifact/e1a415c3-19a3-4d8d-bd71-8b9c19f38535).

## 5. Garantias obrigatórias

1. `domain` não depende de nada além da biblioteca padrão; `usecase` nunca importa um adapter concreto.
2. Toda operação de I/O recebe `context.Context`, respeita cancelamento e tem timeout definido.
3. O progresso no AniList **nunca regride** por ação automática: só aumenta, ou muda por ação explícita do usuário.
4. Cada conclusão de episódio gera **no máximo uma** atualização efetiva no tracker, mesmo com retry, reabertura do episódio ou reinício da aplicação.
5. Atualizações não enviadas (offline, `429`, `5xx`) são **persistidas em disco** e reenviadas depois, sobrevivendo ao reinício.
6. A falha de uma fonte de vídeo não interrompe a busca nem a reprodução enquanto outra fonte puder atender.
7. A falha do AniSkip nunca impede a reprodução.
8. A TUI nunca bloqueia: todo I/O roda em `tea.Cmd`, e sair de uma tela cancela o trabalho pendente dela.
9. Trocar o domínio de uma fonte exige apenas editar o `config.json`, sem recompilar.
10. Escritas em disco são atômicas (arquivo temporário + `rename`): um encerramento abrupto não corrompe estado.
11. Tokens ficam em arquivo com permissão `0600` (ou no keyring do sistema) e nunca aparecem em logs, erros ou na tela.
12. Nenhum processo `mpv` ou socket IPC permanece após o encerramento da aplicação, inclusive por `Ctrl+C`.
13. Não há estado global mutável nem `init()` com efeitos colaterais fora do `main`.

## 6. Modelo de domínio

### Encapsulamento e erros

Modele entidades com construtores que validam a entrada (`NewEpisode`, `NewProgress`…) e métodos explícitos de transição. Valores zero ou inválidos devem ser rejeitados pelo construtor, e as invariantes preservadas em todas as operações públicas.

Separe **criação** (com validação e regras) de **reidratação** (reconstrução a partir do disco ou de uma API), que não deve reaplicar transições.

Erros de domínio devem ser classificáveis com `errors.Is`/`errors.As`. Defina ao menos:

| Erro | Quando |
| --- | --- |
| `ErrNotFound` | Anime, episódio ou entrada da lista inexistente |
| `ErrSourceUnavailable` | Fonte fora do ar, timeout ou erro `5xx` |
| `ErrSourceChanged` | A fonte respondeu, mas o HTML/JSON não tem a estrutura esperada |
| `ErrUnauthorized` | Token ausente, inválido ou revogado |
| `ErrRateLimited` | Limite de requisições atingido; deve carregar o tempo de espera |
| `ErrInvalidTransition` | Mudança de status ou progresso que viola as regras da seção 6.5 |

`panic` não representa erro de negócio nem falha de fonte.

### 6.1. Anime

Representa uma obra. Precisa carregar:

- identidade no tracker (`AniListID`) e, quando existir, no MyAnimeList (`MalID`, exigido pelo AniSkip);
- referências por fonte (ex.: `animefire → "slug-do-anime"`), pois o mesmo anime tem identificadores diferentes em cada fonte;
- títulos (principal, nativo e alternativos), sinopse, formato (TV, filme, OVA…) e total de episódios, que pode ser **desconhecido** enquanto o anime estiver em exibição.

Documente como é feita a **correspondência** entre o resultado de uma fonte e a entrada do AniList: normalização de títulos, critério de confiança e o que acontece quando há ambiguidade. Quando a confiança for baixa, o usuário deve confirmar a escolha, que é persistida para não ser perguntada de novo.

### 6.2. Season

No AniList, **cada temporada é uma entrada (Media) independente**, ligada às outras por relações `PREQUEL`/`SEQUEL`. Nas fontes, temporadas costumam aparecer como páginas separadas.

Decida e justifique em `ARCHITECTURE.md` se `Season` continua sendo uma entidade, vira apenas um agrupamento de apresentação ou é removida. A decisão deve manter o progresso sincronizado com a entrada correta do AniList.

### 6.3. Episode

- Número positivo; documente como trata episódios especiais ou fracionados (ex.: `6.5`) quando a fonte os listar.
- Título opcional.
- Duração em `time.Duration`, que pode ser desconhecida até o player informar.
- Referência na fonte, suficiente para resolver o link do vídeo.

### 6.4. Progress

Progresso de reprodução de **um episódio**, guardado localmente para retomar de onde parou.

- Posição em `time.Duration`, sempre dentro de `[0, duração]`.
- Um episódio é considerado **concluído** quando a posição atinge o limiar configurado (padrão: **85%** da duração) **ou** quando o encerramento (ED) informado pelo AniSkip começa, o que ocorrer primeiro.
- Conclusão é **monotônica**: uma vez concluído, o episódio continua concluído, mesmo que o usuário volte a cena.
- A conclusão é o evento que dispara a sincronização com o tracker; a posição em si nunca é enviada ao AniList.

### 6.5. ListEntry

Estado do anime na lista do usuário, espelhando o AniList.

| Status | Significado |
| --- | --- |
| `PLANNING` | Planeja assistir |
| `CURRENT` | Assistindo |
| `COMPLETED` | Concluído |
| `PAUSED` | Pausado |
| `DROPPED` | Abandonado |
| `REPEATING` | Reassistindo |

Regras:

- `progress` é o número de episódios assistidos: nunca negativo e nunca maior que o total, quando o total for conhecido.
- Concluir um episódio de um anime em `PLANNING`, `PAUSED`, `DROPPED` ou sem entrada move o status para `CURRENT`.
- Atingir o último episódio conhecido move o status para `COMPLETED`.
- Em `COMPLETED`, assistir de novo só altera o progresso se o usuário escolher explicitamente `REPEATING`.
- Ao combinar estado local e remoto, prevalece o **maior** progresso. A automação nunca diminui o progresso remoto.

### 6.6. SkipInterval

Intervalo de pulo com tipo (`OPENING`, `ENDING` e, opcionalmente, `RECAP`), início e fim.

- `início < fim` e ambos dentro da duração do episódio, quando conhecida.
- Intervalos inválidos vindos da API são descartados com log, sem falhar a reprodução.

### 6.7. Stream

Resultado da resolução de um episódio em uma fonte: URL do vídeo, qualidade e os **cabeçalhos HTTP necessários** (ex.: `Referer`, `User-Agent`), que o adapter do player repassa ao mpv. Documente o tempo de validade do link, se a fonte o expirar.

## 7. Portas

As portas ficam em `internal/usecase`, são pequenas (**Interface Segregation**) e falam apenas em tipos do domínio. Nenhum tipo do AniList, do mpv ou do HTML de uma fonte pode aparecer nelas.

As assinaturas abaixo são uma **referência**. Podem mudar, desde que a mudança seja justificada.

```go
type AnimeSource interface {
    Name() string
    Search(ctx context.Context, query string) ([]domain.SourceResult, error)
    Episodes(ctx context.Context, ref domain.SourceRef) ([]domain.Episode, error)
    Stream(ctx context.Context, ep domain.EpisodeRef) (domain.Stream, error)
}

type Player interface {
    Play(ctx context.Context, s domain.Stream, opts PlayOptions) (PlaybackSession, error)
}

type PlaybackSession interface {
    Events() <-chan PlaybackEvent // posição, duração, pausa, fim
    Seek(ctx context.Context, to time.Duration) error
    Close() error
}

type SkipProvider interface {
    SkipTimes(ctx context.Context, malID, episode int, length time.Duration) ([]domain.SkipInterval, error)
}

type Tracker interface {
    Viewer(ctx context.Context) (domain.User, error)
    SearchMedia(ctx context.Context, query string) ([]domain.Anime, error)
    Entry(ctx context.Context, animeID int) (domain.ListEntry, error)
    SaveEntry(ctx context.Context, e domain.ListEntry) (domain.ListEntry, error)
}

type TokenStore interface { /* Load, Save, Delete */ }
type ProgressStore interface { /* posição local por episódio */ }
type SyncQueue interface { /* atualizações pendentes para o tracker */ }
type Clock interface { Now() time.Time }
```

O agregador de fontes deve **implementar `AnimeSource`** (padrão *Composite*): para os casos de uso, uma fonte e um agregador de fontes são intercambiáveis (**Liskov**).

## 8. Casos de uso

Um arquivo por caso de uso em `internal/usecase`, cada um como uma struct construída com as portas de que precisa.

### 8.1. SearchAnime

- A consulta é aparada; com menos de 2 caracteres, retorna erro de validação, sem chamada externa.
- Consulta as fontes habilitadas **em paralelo**, cada uma com timeout próprio.
- Resultados equivalentes de fontes diferentes são agrupados, preservando a referência de cada fonte.
- Falha parcial devolve os resultados disponíveis e informa quais fontes falharam. Falha total devolve um erro que agrega as causas (`errors.Join`).

### 8.2. AuthenticateTracker

- Conduz o fluxo OAuth do AniList usando o `clientId` informado pelo usuário no `config.json`. Nenhum segredo fica versionado.
- Use o *implicit grant* com o redirect de PIN do AniList, ou um servidor local em `localhost`; documente a escolha.
- Valida o token com a consulta `Viewer` antes de persistir.
- Oferece logout, que remove o token.
- Um `401` em qualquer operação posterior leva à reautenticação **sem descartar** a fila de sincronização.

### 8.3. PlayEpisode

Segue o fluxo da seção 4:

1. resolve o `Stream` na fonte preferida e, em caso de falha, nas seguintes, na ordem de prioridade;
2. busca os intervalos de pulo com timeout curto (sugestão: 3 s) e cache local; sem `MalID` ou sem dados, reproduz sem pulo;
3. inicia o player na posição salva do episódio, se houver e se ele não estiver concluído;
4. salva a posição periodicamente (sugestão: a cada 10 s) e no encerramento da sessão;
5. ao detectar a conclusão (seção 6.4), registra a atualização de `ListEntry` e tenta sincronizar; se falhar de forma transitória, enfileira.

O pulo automático acontece **uma única vez por intervalo e por sessão**: se o usuário voltar para rever a abertura, ela não é pulada de novo.

### 8.4. AdvanceEpisode

- Determina o próximo episódio a partir do atual e da lista da fonte.
- No último episódio conhecido, informa que não há próximo. Opcionalmente, sugere a sequência (`SEQUEL`) pelo AniList.
- Em animes em exibição (total desconhecido), confia na lista da fonte.

### 8.5. UpdateListStatus

- Altera o status por ação explícita do usuário, validando as regras da seção 6.5.
- Lê o estado remoto antes de gravar e aplica a regra de "maior progresso prevalece".
- Usa a mesma fila de sincronização do `PlayEpisode`.

### 8.6. SyncPendingUpdates

- Executado na inicialização e periodicamente enquanto houver pendências.
- Retry com backoff exponencial e jitter. Em `429`, respeita `Retry-After`.
- Atualizações do mesmo anime são **coalescidas**: só a de maior progresso é enviada.
- Uma atualização confirmada pelo tracker sai da fila; nenhuma sai da fila sem confirmação ou sem ser descartada por erro permanente registrado em log.

## 9. Fontes de vídeo

### Configuração

As fontes são definidas no `config.json`, nunca no código:

```json
{
  "version": 1,
  "player": { "path": "mpv", "extraArgs": [] },
  "tracker": { "provider": "anilist", "clientId": "SEU_CLIENT_ID" },
  "playback": {
    "completionThreshold": 0.85,
    "autoSkip": { "opening": true, "ending": false }
  },
  "sources": [
    { "name": "animefire", "enabled": true, "priority": 1, "baseUrl": "https://animefire.example", "timeout": "10s" },
    { "name": "goyabu",    "enabled": true, "priority": 2, "baseUrl": "https://goyabu.example",    "timeout": "10s" }
  ]
}
```

- O arquivo fica no diretório de configuração do usuário (`os.UserConfigDir()/animego/config.json`) e pode ser sobrescrito por `ANIMEGO_CONFIG`.
- Na primeira execução, um arquivo padrão é gerado e o usuário é informado do caminho.
- Campos desconhecidos são rejeitados, e os erros de validação indicam o campo (ex.: `sources[1].timeout: duração inválida "10"`).
- O campo `version` permite migrar o formato no futuro.

### Adapters

- Cada fonte é um pacote independente que implementa `AnimeSource`.
- O parsing é testado contra **fixtures** de HTML reais versionadas em `testdata/`.
- Mudança de estrutura gera `ErrSourceChanged` com contexto suficiente para diagnóstico, e não `panic` ou resultado vazio silencioso.
- Seja um cliente educado: `User-Agent` identificável, no máximo 2 requisições simultâneas por host e nenhum retry agressivo.

### Agregador

- Ordena as fontes por `priority` e ignora as desabilitadas.
- Na busca, consulta todas em paralelo. Na resolução de stream, tenta em ordem até uma funcionar.
- Cancela o trabalho das demais fontes quando o resultado já foi obtido ou o contexto foi cancelado, sem vazar goroutines.

## 10. Player (mpv)

- Inicia o mpv com `--input-ipc-server` apontando para um socket exclusivo da sessão, além de `--start`, `--force-media-title` e os cabeçalhos do `Stream` (`--http-header-fields`/`--referrer`).
- Observa `time-pos`, `duration`, `pause` e o evento `end-file` por `observe_property`.
- Executa o pulo com `set_property time-pos` quando a posição entra em um intervalo habilitado.
- Conexão ao socket com retry e timeout, pois o mpv leva alguns instantes para criá-lo.
- mpv ausente do `PATH` gera uma mensagem com instruções de instalação, e não um erro genérico de `exec`.
- Fechamento do mpv pelo usuário encerra a sessão normalmente, salvando a posição.
- No encerramento da aplicação, o processo é finalizado e o socket, removido.

## 11. Tracker (AniList)

- Endpoint GraphQL `https://graphql.anilist.co`. Consultas e mutações (`Viewer`, `Media`, `MediaList`, `SaveMediaListEntry`) ficam no adapter, nunca no caso de uso.
- Respeite os cabeçalhos de limite (`X-RateLimit-Remaining`, `Retry-After`) e transforme `429` em `ErrRateLimited`.
- Não sincronize a cada segundo de reprodução: apenas na conclusão de episódio e em mudanças explícitas de status.
- **MyAnimeList está fora do escopo**, mas a solução deve permitir adicioná-lo como um novo adapter de `Tracker` sem alterar os casos de uso. Isso é avaliado (**Open/Closed**).

## 12. AniSkip

- O AniSkip identifica o anime pelo **ID do MyAnimeList**, obtido pelo campo `idMal` do AniList.
- Converta os tipos da API (`op`, `ed`, `mixed-op`, `mixed-ed`, `recap`) para os tipos do domínio. Documente o mapeamento.
- Use cache local com TTL para não repetir a consulta a cada reprodução do mesmo episódio.

## 13. Interface de terminal

### Telas mínimas

| Tela | Conteúdo |
| --- | --- |
| Início | Usuário autenticado (ou convite para login) e lista "Assistindo" do AniList |
| Busca | Campo de busca, resultados com título e fontes disponíveis |
| Episódios | Lista com episódios concluídos marcados e posição salva do episódio em andamento |
| Reprodução | Anime, episódio, posição, status da sincronização e atalhos para próximo/anterior/sair |
| Ajuda | Todos os atalhos |

### Atalhos sugeridos

`/` busca · `enter` seleciona · `n` próximo episódio · `p` episódio anterior · `s` alterar status · `esc` voltar · `q` sair · `?` ajuda

### Requisitos

- A TUI conversa **apenas com casos de uso**, nunca com adapters.
- `Update` é puro em relação a I/O: chamadas externas viram `tea.Cmd`.
- Estados de carregamento e de erro são visíveis; erros não fatais aparecem como aviso, e a aplicação continua utilizável.
- Funciona em um terminal de 80×24 e respeita `NO_COLOR`.
- Nada é escrito em `stdout`/`stderr` enquanto a TUI está ativa; logs vão para arquivo.

## 14. Observabilidade

- Logs JSON com `log/slog` em `os.UserCacheDir()/animego/logs/`, com nível definido por `ANIMEGO_LOG_LEVEL` ou `--debug`.
- Campos de contexto quando disponíveis: `operation`, `source`, `animeId`, `episode`, `attempt`, `duration`, `error`.
- Nunca registre token, cabeçalho `Authorization` ou URL que contenha credencial.
- Comando `animego --version` com versão, commit e data de build (`-ldflags`).

## 15. O que pode e o que não pode

### Permitido

| Item | Observação |
| --- | --- |
| Suportar apenas Linux em tempo de execução | macOS e Windows são diferenciais |
| Escolher a biblioteca de parsing de HTML | `x/net/html` ou `goquery` |
| Usar `cobra` ou `flag` para subcomandos simples | `login`, `logout`, `config path`, `--version` |
| Bibliotecas utilitárias pequenas | Justificadas em `ARCHITECTURE.md` |
| Frameworks de mock | Preferência por fakes escritos à mão; frameworks são aceitos se justificados |
| Cache local de buscas, episódios e skip times | Com TTL e sem afetar a correção |
| Apenas conteúdo legendado em PT-BR | Seleção de dublado é diferencial |
| Reinterpretar uma regra ambígua | Desde que a interpretação esteja documentada |

### Proibido

| Item | Motivo |
| --- | --- |
| `domain` importar algo fora da biblioteca padrão, ou `usecase` importar um adapter | Quebra a regra de dependência |
| Domínio de fonte fixo no código | Toda mudança de domínio exigiria recompilar |
| Frameworks de DI | O composition root deve ser explícito |
| Navegador headless (chromedp, rod, Playwright) | Mantém o binário simples e portátil |
| Contornar captcha, DRM ou desafios anti-bot | Limite ético do projeto |
| Hospedar, baixar ou redistribuir vídeos | Fora do escopo e do propósito do projeto |
| Segredos versionados ou registrados em log | Segurança |
| `go test ./...` depender de internet | Testes padrão devem ser reprodutíveis offline |
| Estado global mutável, `init()` com efeitos colaterais | Dificulta testes e esconde dependências |
| `panic` para erro de negócio ou de fonte | Erros devem ser tratáveis |
| Ignorar erros sem justificativa (`_ = f()`) | Falhas silenciosas |
| Escrever no terminal enquanto a TUI está ativa | Corrompe a interface |

## 16. Etapas esperadas

A ordem abaixo segue o checklist do [README](README.md) e prioriza uma **fatia vertical funcionando** antes da TUI. Cada etapa deve ser entregue em um PR próprio, com testes e com o checklist do README atualizado.

| # | Etapa | Pronto quando |
| --- | --- | --- |
| 0 | **Fundação** | `go.mod` com módulo e versão; todos os arquivos com `package`; `.gitignore`; `Makefile`; CI rodando `gofmt`, `go vet` e `go test -race`; teste da regra de dependência |
| 1 | **Domínio** | `Anime`, `Season` (ou a decisão da seção 6.2), `Episode`, `Progress`, `ListEntry`, `SkipInterval`, `Stream` e erros, com construtores, invariantes e testes |
| 2 | **Portas e casos de uso** | Interfaces da seção 7 e casos de uso da seção 8 testados com fakes, sem nenhum adapter real |
| 3 | **Configuração e armazenamento local** | `config.json` com validação e geração padrão; token, progresso e fila persistidos com escrita atômica |
| 4 | **Adapter AnimeFire** | Busca, episódios e stream testados com fixtures |
| 5 | **Adapter mpv** | Reprodução, eventos, seek e encerramento testados contra um servidor IPC falso |
| 6 | **Fatia vertical** | `animego play "<busca>" --episode N` funciona de ponta a ponta, sem TUI e sem tracker |
| 7 | **Adapter AniList** | Login, leitura e gravação da lista, fila offline e tratamento de `401`/`429` |
| 8 | **Adapter AniSkip** | Pulo automático com cache e degradação silenciosa |
| 9 | **Goyabu + agregador** | Fallback entre fontes, busca paralela e erros agregados |
| 10 | **TUI** | Telas e atalhos da seção 13, conversando só com casos de uso |
| 11 | **Release** | README com instalação e uso, `ARCHITECTURE.md` completo, binários gerados pelo CI |

## 17. Verificação obrigatória

### Testes unitários

Cubra as invariantes e transições de `Progress`, `ListEntry` e `SkipInterval`, o limiar de conclusão, a regra de "maior progresso prevalece", a coalescência da fila e cada caso de uso com fakes, incluindo os caminhos de erro.

### Testes de adapters

- Fontes: `httptest.Server` servindo fixtures de `testdata/`, incluindo HTML alterado, `404`, `5xx` e timeout.
- AniList: respostas GraphQL gravadas, incluindo `401`, `429` com `Retry-After` e erro de GraphQL com status `200`.
- AniSkip: episódio com dados, sem dados e intervalos inválidos.
- mpv: servidor IPC falso em socket Unix emitindo eventos de posição e fim de arquivo.
- Configuração: arquivo ausente, inválido, com campo desconhecido e com duração malformada.

### Cenários obrigatórios

1. A primeira fonte está fora do ar: busca e reprodução acontecem pela segunda, com aviso na TUI.
2. O `baseUrl` de uma fonte é alterado no `config.json` e a aplicação passa a usá-lo sem recompilar.
3. Uma fixture com HTML alterado gera `ErrSourceChanged`; a aplicação continua utilizável.
4. O usuário fecha o mpv no meio do episódio; ao reabri-lo, a reprodução retoma da posição salva.
5. O episódio passa do limiar de conclusão: o AniList recebe **exatamente uma** atualização, mesmo que o episódio seja reaberto e concluído de novo.
6. Sem rede no momento da conclusão: a atualização é enfileirada e sincronizada na próxima execução com rede.
7. O progresso remoto é maior que o local (assistido em outro dispositivo): nada regride.
8. O AniList responde `429`: o `Retry-After` é respeitado e a atualização não se perde.
9. Anime sem `MalID` ou sem dados no AniSkip: reprodução normal, sem erro.
10. O usuário volta para rever a abertura: ela não é pulada de novo.
11. Token revogado (`401`): a reautenticação é solicitada e a fila permanece intacta.
12. `Ctrl+C` durante a reprodução: mpv encerrado, socket removido, posição salva e terminal restaurado.
13. Conclusão do último episódio: o status muda para `COMPLETED`.
14. A aplicação é encerrada durante a escrita de um arquivo de estado: o arquivo anterior continua válido.

Execute `go test -race ./...`. Testes que usam rede real ficam atrás da build tag `e2e` e não fazem parte da verificação padrão.

## 18. Critérios de avaliação

| Critério | Pontos | Evidência esperada |
| --- | ---: | --- |
| Arquitetura e regra de dependência | 20 | Camadas respeitadas e verificadas por teste; composition root explícito; portas sem vazamento de tecnologia |
| Modelagem de domínio | 15 | Construtores com validação, invariantes, transições explícitas e erros classificáveis |
| Resiliência | 15 | Fallback entre fontes, timeouts, cancelamento, degradação do AniSkip, encerramento limpo |
| Sincronização com o tracker | 15 | Sem regressão, sem duplicidade, fila persistente, `401`/`429` tratados |
| Player e pulo de OP/ED | 10 | IPC robusto, retomada de posição, pulo único por intervalo, sem processos órfãos |
| Testes | 10 | Fixtures reais, fakes, cenários da seção 17, `-race` |
| Experiência na TUI | 5 | Navegação fluida, sem travamentos, estados de carregamento e erro claros |
| Observabilidade | 5 | Logs úteis para diagnóstico, sem dados sensíveis |
| Documentação | 5 | Execução reproduzível e decisões explicadas |
| **Total** | **100** | |

São **eliminatórios**: `domain` ou `usecase` importando infraestrutura; TUI chamando adapter diretamente; `go test ./...` dependendo de internet; token versionado ou exibido em log; regressão automática de progresso no AniList; atualização duplicada no tracker; `panic` causado por falha de fonte; domínio de fonte fixo no código; TUI travando durante I/O; processo `mpv` órfão após o encerramento; download, redistribuição ou contorno de proteções de conteúdo; ausência de testes do domínio.

**Diferenciais** (não somam pontos, mas pesam na conversa): adapter de MyAnimeList como prova do Open/Closed; suporte a macOS e Windows; token no keyring do sistema; *fuzzing* dos parsers de HTML (`go test -fuzz`); verificação de vazamento de goroutines (`goleak`); comando `animego doctor` (mpv, configuração, fontes e token); release automatizado com GoReleaser; histórico de episódios assistidos.

## 19. Entrega

Entregue o código, as fixtures, o `config.example.json` e instruções suficientes para outra pessoa compilar e usar a aplicação a partir de um checkout limpo.

O `README.md` da solução deve explicar pré-requisitos, instalação, primeiro uso (geração do `config.json` e login no AniList), atalhos, localização dos arquivos (config, estado, logs) e comandos de teste.

O `ARCHITECTURE.md` deve registrar, no mínimo, as decisões sobre:

- correspondência entre fonte e AniList, e o tratamento de `Season`;
- regra de conclusão de episódio e política de sincronização (quando, retry, coalescência e conflito);
- ordem, timeouts e fallback das fontes;
- formato e localização da configuração e do estado local;
- armazenamento do token e fluxo OAuth escolhido;
- protocolo IPC com o mpv e plataformas suportadas;
- modelo de concorrência da TUI;
- taxonomia de erros;
- limitações, interpretações adotadas e trabalho não concluído.

Disponibilize os comandos abaixo, ou equivalentes documentados:

```sh
make build            # go build -o bin/animego ./cmd/animego
go test ./...
go test -race ./...
go vet ./...
gofmt -l .            # deve retornar vazio
go test -tags=e2e ./...   # opcional: usa as fontes e APIs reais
```

Commits pequenos, com mensagem descritiva e padrão consistente (o repositório usa `[feat]`, `[fix]`…), e um PR por etapa da seção 16.

## 20. Princípios demonstrados

Esta seção serve de roteiro para defender a solução em uma entrevista. Cada princípio precisa apontar para um trecho concreto do código.

| Princípio | Onde deve aparecer |
| --- | --- |
| **Single Responsibility** | Um caso de uso por arquivo; o adapter de fonte só faz scraping, e a regra de conclusão fica no domínio |
| **Open/Closed** | Nova fonte = novo adapter + entrada no `config.json`; novo tracker = novo adapter, sem tocar nos casos de uso |
| **Liskov Substitution** | O agregador e uma fonte isolada são intercambiáveis como `AnimeSource` |
| **Interface Segregation** | Portas pequenas; o caso de uso de busca não conhece o `Player` |
| **Dependency Inversion** | Casos de uso dependem de interfaces declaradas por eles mesmos; adapters as implementam |
| **Ports & Adapters** | Toda tecnologia externa (HTML, GraphQL, IPC, disco) atrás de uma porta |
| **Composite** | O agregador de fontes |
| **Composition Root** | `cmd/animego/main.go` é o único ponto que conhece todas as implementações |
| **Fail-safe / degradação** | AniSkip e fontes secundárias falham sem derrubar a experiência principal |
| **Idempotência** | Fila coalescida e regra de "maior progresso prevalece" |

### Perguntas que a solução deve permitir responder

- Como você adicionaria o MyAnimeList? Quais arquivos mudariam?
- O que acontece se o AnimeFire mudar o HTML amanhã? Como você descobriria?
- Por que a TUI não chama o adapter diretamente? O que você perderia se chamasse?
- Como você garante que o progresso não é enviado duas vezes, nem regride?
- Como você testa a integração com o mpv sem abrir um player de verdade?
- Por que injeção de dependências manual, e não um framework?
- O que acontece com uma atualização pendente se a aplicação for encerrada à força?
- Qual decisão você tomaria diferente hoje, e por quê?

## Aviso legal

Este projeto não hospeda conteúdo. Ele automatiza o acesso a sites de streaming publicamente disponíveis, de forma análoga a uma pessoa usando um navegador. Não há afiliação com nenhum dos sites citados. O uso é de inteira responsabilidade de quem executa a aplicação.
