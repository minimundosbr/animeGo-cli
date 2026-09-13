# Documentação:
- [Arquitetura inicial](https://claude.ai/code/artifact/e1a415c3-19a3-4d8d-bd71-8b9c19f38535)


### APIs
- [API aniskip](https://api.aniskip.com/api-docs#/)
- [API anilist](https://docs.anilist.co/)

### Arquitetura atual:
```
animego-cli/
├── go.mod
├── README.md
├── .gitignore
├── cmd/
│   └── animego/
│       └── main.go              # composition root (por último)
├── internal/
│   ├── domain/                  # 1º — núcleo, zero dependências
│   │   ├── anime.go
│   │   ├── season.go
│   │   ├── episode.go
│   │   └── progress.go
│   │
│   ├── usecase/                 # 2º — casos de uso + portas (interfaces)
│   │   ├── search_anime.go
│   │   ├── authenticate_tracker.go
│   │   ├── update_list_status.go
│   │   ├── play_episode.go
│   │   ├── advance_episode.go
│   │   ├── anime_source.go      # interface AnimeSource
│   │   ├── player.go            # interface Player
│   │   ├── skip_provider.go     # interface SkipProvider
│   │   └── tracker_provider.go  # interface TrackerProvider
│   │
│   ├── adapter/                 # 3º — implementações concretas (por último entre os três)
│   │   ├── source/
│   │   │   ├── animefire/
│   │   │   ├── goyabu/
│   │   │   └── aggregator/
│   │   ├── player/
│   │   │   └── mpv/
│   │   ├── tracker/
│   │   │   └── anilist/
│   │   └── skip/
│   │       └── aniskip/
│   │
│   ├── ui/
│   └── config/

```


### Projetos utilizados:
- [aniskip](https://github.com/synacktraa/ani-skip)
- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [anilist](https://github.com/AniList/docs)

### Referências para o projeto:
- [curd](https://github.com/Wraient/curd/tree/main)
- [ani-cli](https://github.com/pystardust/ani-cli)