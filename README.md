# Termsnake

A polished terminal Snake game in Go with tcell, SQLite high scores, CI/CD, and Docker.

## Overview

Termsnake is a terminal-based Snake game that runs in any modern terminal (Linux, macOS, Windows Terminal). It's written in Go as a learning project covering core language features: structs, slices, goroutines, channels, tickers, packages, and testing.

The project follows a phased roadmap:
- **v1** — single-player terminal game with local persistence, tests, and release binaries
- **v1.1** — wrap-around mode, color themes, config file
- **v2** — online leaderboard, WebSocket multiplayer, DevOps showcase

## Features (v1)

- Snake moves on a grid; arrow keys and WASD control it
- Food spawns on a random empty cell; eating grows the snake and adds score
- Game over on wall or self collision
- Pause/resume (P/Space), restart (R), quit (Q/Esc)
- Speed increases as score rises (150ms Easy → 100ms Normal → 60ms Hard, minimum 40ms)
- Persistent local high scores (top 10, stored in SQLite)
- Main menu, difficulty selection (Easy, Normal, Hard)
- Wrap-around (no walls) mode (v1.1)
- Color themes (v1.1)
- Config file (~/.config/termsnake/config.toml) (v1.1)

## Screens (6)

1. **Main menu** — title art, Play, High Scores, Quit
2. **Difficulty select** — Easy, Normal, Hard
3. **Game screen** — HUD plus board
4. **Pause overlay** — Resume, Restart, Quit to menu
5. **Game over** — final score, new-high-score banner, name entry if it qualifies, Restart or Menu
6. **High scores** — top 10 table with score, difficulty, date

## Controls

| Key | Action |
|-----|--------|
| Arrow keys / WASD | Move |
| P or Space | Pause/resume |
| R | Restart (on game over) |
| Enter | Confirm in menus |
| ↑/↓ | Navigate menus |
| Q or Esc | Back/quit |
| Ctrl+C | Quit immediately (terminal restored) |

## Terminal requirements

- Minimum size: **44 columns x 24 rows** (smaller shows a friendly message)
- Board: 20×20 cells, each cell 2 characters wide so it looks square
- Uses only the 16 ANSI colors so it works everywhere

## Installation

```bash
# Clone
git clone https://github.com/rjeff-sudo/termsnake-go.git
cd termsnake-go

# Build
make build       # produces bin/termsnake
make run         # runs via `go run ./cmd/termsnake`

# Or run directly
go run ./cmd/termsnake
```

## Build from source

```bash
go mod tidy        # fetch dependencies
make build         # CGO_ENABLED=0 binary in bin/termsnake
make test          # go test ./...
make test-race     # go test -race ./...
make vet           # go vet ./...
make lint          # golangci-lint run
```

## Cross-compiled releases

On tags `v*`, GitHub Actions builds for linux/darwin/windows (amd64 and arm64) and attaches binaries to the release.

## Docker

```bash
docker build -t termsnake .
docker run -it termsnake
```

## Configuration

A config file is auto-created at `~/.config/termsnake/config.toml` on first run. It stores the player name and settings. Permissions are set to 0600.

## Development roadmap

- **v1.1** — wrap-around mode, color themes, config file
- **v2** — online leaderboard via Go HTTP API, token-based identity, WebSocket live spectating or two-player mode, Dockerized server, CI/CD, monitoring

## License

MIT

## Project structure

```
termsnake-go/
├── cmd/termsnake/main.go        # game client entrypoint
├── internal/game/              # pure logic: board, snake, food, rules, tick
├── internal/input/             # key events to game commands
├── internal/ui/                # tcell rendering, screens, themes
├── internal/storage/           # SQLite scores and settings
├── migrations/                 # SQL files
├── Dockerfile                  # multi-stage build
├── Makefile
├── .github/workflows/ci.yml
├── go.mod
└── README.md
```

## Success criteria

- Runs at a steady tick with no flicker on Linux, macOS, and Windows Terminal
- Zero crashes on terminal resize or Ctrl+C (terminal always restored)
- Core logic covered by unit tests (collision, growth, food spawn)
- Cross-compiled binaries published on GitHub Releases