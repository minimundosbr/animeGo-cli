# animeGo-cli

🇧🇷 Português | 🇺🇸 [English](README.en.md)

> ⚠️ **Projeto em desenvolvimento.** Em construção ativa. A camada de domínio está pronta; casos de uso, interfaces e adapters são os próximos passos. Nada ainda é executável.

Uma CLI em terminal (TUI), escrita em Go, para buscar animes, assistir episódios através de um player externo e sincronizar automaticamente o progresso com o AniList.

## Funcionalidades

- Interface interativa via teclado, direto no terminal — sem precisar de navegador.
- Busca em múltiplas fontes com fallback automático (começando com AnimeFire e Goyabu).
- Reprodução via [mpv](https://mpv.io/), controlado por IPC — volume, avançar, pausar, tudo pelo seu player de sempre.
- Pulo automático de abertura/encerramento via [AniSkip](https://aniskip.com/).
- Progresso sincronizado com o [AniList](https://anilist.co/) (suporte a MyAnimeList planejado para uma fase futura).
- Configuração de fontes editável pelo usuário (`config.json`) — sem precisar recompilar o binário quando o domínio de uma fonte muda.

## Status do projeto

Este projeto está sendo construído publicamente, um passo deliberado de cada vez, seguindo Clean Architecture e princípios SOLID como projeto de aprendizado e portfólio.

Os requisitos completos, garantias, restrições e critérios de avaliação estão escritos como um desafio técnico em [DESAFIO.md](DESAFIO.md).

- [x] Entidades de domínio (`internal/domain`) — `Anime`, `Season`, `Episode`, `Progress`
- [ ] Casos de uso e interfaces/portas (`internal/usecase`) — em andamento
- [ ] Adapter do AnimeFire
- [ ] Adapter do Goyabu + agregador multi-fonte (fallback)
- [ ] Adapter do mpv
- [ ] Adapter do AniList
- [ ] Adapter do AniSkip
- [ ] Interface de terminal (bubbletea)
- [ ] Composition root (`cmd/animego/main.go`)

## Arquitetura

Este projeto segue **Clean Architecture**: as dependências sempre apontam para dentro, em direção ao domínio. Camadas externas (adapters, frameworks) dependem das camadas internas (casos de uso, entidades) — nunca o contrário. É isso que permite trocar uma fonte de vídeo ou o tracker sem tocar na lógica de negócio principal.

```
internal/
├── domain/     # Entidades — sem depender de mais nada
├── usecase/    # Regras de negócio + portas (interfaces) consumidas pelos adapters
├── adapter/    # Implementações concretas: AnimeFire, Goyabu, mpv, AniList, AniSkip
├── ui/         # Interface de terminal (bubbletea)
└── config/     # Configuração editável pelo usuário
```

Um diagrama completo das camadas e do fluxo do caso de uso "Reproduzir Episódio" está disponível aqui: [Diagrama de arquitetura](https://claude.ai/code/artifact/e1a415c3-19a3-4d8d-bd71-8b9c19f38535).

## Stack técnico

- **Linguagem:** Go
- **TUI:** [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) (+ `bubbles`, `lipgloss`)
- **Player:** [mpv](https://mpv.io/), controlado via socket IPC (JSON)
- **Fontes de vídeo:** AnimeFire, Goyabu (PT-BR, com fallback automático)
- **Tracker:** AniList (API GraphQL)
- **Skip de abertura/encerramento:** AniSkip (API pública)
- **Formato de config:** JSON

## Requisitos

- Go 1.22+
- [mpv](https://mpv.io/) instalado e disponível no `PATH`

## Instalação

Ainda não disponível — o projeto não gera um binário utilizável nesta fase. Esta seção será preenchida assim que o composition root (`cmd/animego/main.go`) e os primeiros adapters estiverem funcionando.

## Uso

Em breve.

## Como contribuir

Contribuições serão bem-vindas assim que o MVP estiver suficientemente pronto para uso. Algumas convenções a manter em mente:

- Identificadores de código (tipos, campos, funções) são em **inglês**.
- A separação do Clean Architecture é reforçada: `domain` nunca importa `usecase`; `usecase` nunca importa um `adapter` concreto, só as portas (interfaces) das quais depende.
- Um arquivo por entidade em `internal/domain`, um arquivo por caso de uso em `internal/usecase`.

Um `CONTRIBUTING.md` com diretrizes mais detalhadas será adicionado conforme o projeto amadurecer.

## Projetos relacionados

- [alvarorichard/GoAnime](https://github.com/alvarorichard/GoAnime) — projeto mais próximo em espírito, sem tracking via AniList/MAL.
- [Wraient/curd](https://github.com/Wraient/curd) — mpv + AniList + MyAnimeList + AniSkip, fontes legendadas em inglês. Recomendado para quem procura legendado em inglês em vez de fontes PT-BR.
- [pystardust/ani-cli](https://github.com/pystardust/ani-cli) — um clássico do gênero.

## Aviso legal / ético

Este projeto não hospeda nenhum conteúdo. Ele automatiza o acesso a sites de streaming publicamente disponíveis, de forma análoga a uma pessoa operando um navegador. Não há afiliação com nenhum dos sites referenciados. O uso é de responsabilidade exclusiva de quem executa o software.

## Licença

A definir.