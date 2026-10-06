# animeGo-cli

🇺🇸 English | 🇧🇷 [Português](README.pt-br.md)

> ⚠️ **Work in progress.** This project is under active development. Domain layer is complete; use cases, interfaces and adapters are being built next. Nothing is runnable yet.

A terminal-based (TUI) CLI, written in Go, to search for anime, watch episodes through an external player, and automatically sync your watch progress with AniList.

## Features

- Interactive, keyboard-driven terminal interface — no web browser needed.
- Multi-source search with automatic fallback (starting with AnimeFire and Goyabu).
- Playback through [mpv](https://mpv.io/), controlled via IPC — volume, seek, pause, all from your usual player.
- Automatic opening/ending skip via [AniSkip](https://aniskip.com/).
- Progress tracking synced with [AniList](https://anilist.co/) (MyAnimeList support planned for a later phase).
- User-editable source configuration (`config.json`) — no need to rebuild the binary when a source's domain changes.

## Project status

This is being built in public, one deliberate step at a time, following Clean Architecture and SOLID principles as a learning and portfolio project.

The full requirements, guarantees, constraints and evaluation criteria are written as a technical challenge in [DESAFIO.md](DESAFIO.md) (Portuguese).

- [x] Domain entities (`internal/domain`) — `Anime`, `Season`, `Episode`, `Progress`
- [ ] Use cases and interfaces / ports (`internal/usecase`) — in progress
- [ ] AnimeFire adapter
- [ ] Goyabu adapter + multi-source aggregator (fallback)
- [ ] mpv adapter
- [ ] AniList adapter
- [ ] AniSkip adapter
- [ ] Terminal UI (bubbletea)
- [ ] Composition root (`cmd/animego/main.go`)

## Architecture

This project follows **Clean Architecture**: dependencies always point inward, toward the domain. Outer layers (adapters, frameworks) depend on inner layers (use cases, entities) — never the other way around. This is what lets a video source or the tracker be swapped out without touching the core business logic.

```
internal/
├── domain/     # Entities — no dependency on anything else
├── usecase/    # Business rules + ports (interfaces) consumed by adapters
├── adapter/    # Concrete implementations: AnimeFire, Goyabu, mpv, AniList, AniSkip
├── ui/         # Terminal UI (bubbletea)
└── config/     # User-editable configuration
```

A full diagram of the layers and the "Play Episode" use case flow is available here: [Architecture diagram](https://claude.ai/code/artifact/e1a415c3-19a3-4d8d-bd71-8b9c19f38535).

## Tech stack

- **Language:** Go
- **TUI:** [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) (+ `bubbles`, `lipgloss`)
- **Player:** [mpv](https://mpv.io/), controlled via JSON IPC
- **Video sources:** AnimeFire, Goyabu (PT-BR, with automatic fallback)
- **Tracker:** AniList (GraphQL API)
- **OP/ED skip:** AniSkip (public API)
- **Config format:** JSON

## Requirements

- Go 1.22+
- [mpv](https://mpv.io/) installed and available on your `PATH`

## Installation

Not available yet — the project doesn't build a usable binary at this stage. This section will be filled in once the composition root (`cmd/animego/main.go`) and the first working adapters are in place.

## Usage

Coming soon.

## Contributing

Contributions are welcome once the MVP is far enough along to be usable. A few conventions to keep in mind:

- Code identifiers (types, fields, functions) are in **English**.
- Clean Architecture layering is enforced: `domain` never imports `usecase`; `usecase` never imports a concrete `adapter`, only the ports (interfaces) it depends on.
- One file per entity in `internal/domain`, one file per use case in `internal/usecase`.

A `CONTRIBUTING.md` with more detailed guidelines will be added as the project matures.

## Related projects

- [alvarorichard/GoAnime](https://github.com/alvarorichard/GoAnime) — closest project in spirit, without AniList/MAL tracking.
- [Wraient/curd](https://github.com/Wraient/curd) — mpv + AniList + MyAnimeList + AniSkip, English-subbed sources. Recommended if you're looking for English subs instead of PT-BR sources.
- [pystardust/ani-cli](https://github.com/pystardust/ani-cli) — a classic of the genre.

## Legal / ethical disclaimer

This project does not host any content. It automates access to publicly available streaming websites, in a way roughly analogous to a human operating a browser. There is no affiliation with any of the referenced sites. Use is entirely at your own risk and responsibility.

## License

To be determined.