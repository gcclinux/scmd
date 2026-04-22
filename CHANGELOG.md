# Changelog

All notable changes to this project will be documented in this file.

## [2.1.0] - 2026-04-22

### Added
- **Execute correct answer in scmd-cli** — users can select numbered code blocks `[1]`, `[2]`, `[3]` from AI responses and execute them directly.
- **MCP server initial version** — scmd moved from direct PostgreSQL queries to MCP server architecture.
- **AI Persona search focus** — implemented persona-aware search that tailors results based on the active AI persona context.

## [2.0.7] - 2026-04-15

### Added
- **Model Context Protocol (MCP)** server integration (`--mcp`).
- Purpose-built interface for **AI Agents** to interact with the SCMD knowledge base.
- Exposes tools to AI assistants: `search_commands`, `add_command`, and `get_stats`.
- Support for official **Go MCP SDK** over stdio transport.
- Comprehensive [MCP-walkthrough.md](docs/MCP-walkthrough.md) for assistant setup (Claude Desktop, Cursor, etc.).

### Changed
- Upgraded project to **Go 1.25.0** to support enhanced MCP protocol capabilities.

## [2.0.6] - 2026-04-04

### Added
- Six new specialized AI personas for focused system administration and command help:
  - `/ubuntu` — Focused on Ubuntu Linux specific commands, patches, and fixes.
  - `/debian` — Focused on Debian and derived distributions.
  - `/fedora` — Focused on Fedora ecosystem and DNF administration.
  - `/windows` — Focused on standard Windows management and CMD tools.
  - `/powershell` — Focused on PowerShell cmdlets, scripting, and automation.
  - `/archlinux` — Focused on Arch Linux, Pacman, and rolling release maintenance.
- Interactive feedback for persona-based queries (saves AI answers to database).

### Changed
- Rebalanced the interactive help menu layout for better readability.
- Consolidated slash commands to prioritize utility and core features.

### Removed
- `/help next` — Simplified help menu navigation.
- `/count` — Integrated database state visibility into standard list commands.

## [2.0.5] - 2026-03-24

### Added
- Standalone interactive setup commands for first-time application configuration
- `--create-db-postgresql` — interactive PostgreSQL database setup (prompts for host, port, user, password, db/table name)
- `--create-db-sqlite` — interactive SQLite database setup (lightweight, no server required, stored in ~/.scmd/)
- `--server-ollama` — interactive Ollama AI server setup (prompts for host, chat model, embedding model, dimension)
- `--server-gemini` — interactive Gemini AI server setup (prompts for API key, chat model, embedding model, dimension)
- SQLite database support as an alternative to PostgreSQL (pure Go driver, no CGo)
- SQLite-compatible query layer with cosine similarity vector search in Go
- `SaveConfig` and `CurrentConfig` helpers for reading/writing ~/.scmd/config.json
- `db_type` field in config.json to select between `postgresql` and `sqlite`

## [2.0.1] - 2026-02-23

### Added
- Several features added to interactive CLI
- Enhanced natural language query support
- Improved slash commands functionality

## [2.0.0] - 2026-02-19

### Changed
- **BREAKING:** Migrated from SQLite (tardigrade.db) to PostgreSQL database
- Complete database architecture overhaul
- Improved performance and scalability

### Added
- PostgreSQL integration with pgvector support
- Environment-based configuration (.env file)
- Migration guide documentation

## [1.3.8] - 2024-03-20

### Changed
- Paused game development
- Added favicon to web interface

## [1.3.7] - 2023-12-16

### Added
- Started command game page feature

## [1.3.6] - 2023-09-09

### Added
- `-block` flag to disable add commands page

## [1.3.5] - 2023-08-29

### Added
- TLS capabilities for HTTPS support
- SSL certificate configuration options

## [1.3.3] - 2023-04-05

### Changed
- Minor cosmetics on help page

## [1.3.2] - 2023-04-01

### Added
- Help page creation
- Search login functionality

## [1.3.1] - 2023-03-26

### Added
- Check if command already exists before adding
- Cosmetic improvements

## [1.3.0] - 2023-03-19

### Added
- Option to save or display functions

## [1.2.0] - 2023-03-12

### Added
- Option to specify custom port for Web UI

## [1.1.0] - 2023-03-05

### Added
- Binary upgrade option in the menu

## [1.0.2] - 2023-02-26

### Changed
- Minor cosmetic changes in the search UI

## [1.0.1] - 2023-02-19

### Changed
- Recompiled with updated tardigrade-mod v0.2.0

## [1.0.0] - 2023-02-18

### Added
- Initial SCMD CLI release
- Web UI interface
- Basic command search and storage functionality
