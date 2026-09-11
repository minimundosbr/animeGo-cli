Documentação:
[Arquitetura inicial](https://claude.ai/code/artifact/e1a415c3-19a3-4d8d-bd71-8b9c19f38535)

### Arquitetura atual:
```
animego-cli/
├── cmd/
│   └── animego/
│       └── main.go              # bootstrap: monta os adapters concretos
│                                 # e injeta nos casos de uso.
├── internal/
│   ├── domain/                  # Entidades: Anime, Season, Episode, Progress
│   ├── usecase/                 # SearchAnime, PlayEpisode, AdvanceEpisode,
│   │                            # UpdateListStatus, AuthenticateTracker
│   ├── adapter/
│   │   ├── source/
│   │   │   ├── animefire/
│   │   │   ├── goyabu/
│   │   │   └── aggregator/      # MultiSourceAggregator
│   │   ├── player/
│   │   │   └── mpv/
│   │   ├── tracker/
│   │   │   ├── anilist/
│   │   │   └── myanimelist/
│   │   └── skip/
│   │       └── aniskip/
│   ├── ui/                      # telas do bubbletea
│   └── config/                  # leitura do arquivo de config editável
├── go.mod
└── README.md
```

Referências para o projeto:
[curd](https://github.com/Wraient/curd/tree/main)
[ani-cli](https://github.com/pystardust/ani-cli)