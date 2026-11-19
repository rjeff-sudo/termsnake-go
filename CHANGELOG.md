## v1.0.0 - 2026-11-19
### Added
- Initial release: terminal Snake game with tcell UI SQLite persistence
- 6 game screens menu difficulty play pause game-over high-scores
- Direction queue limited to 2 validated entries
- Speed scaling Easy 150ms Normal 100ms Hard 60ms min 40ms
- Tick every 5 foods reduces interval by 5ms min 40ms
- Name entry 3-12 chars default PLAYER for qualifying scores
- SQLite scores stored ~/.local/share/termsnake/termsnake.db
- Platforms linux darwin windows amd64 arm64
