# SCMD — Search Command

> An AI-powered command search and management tool with CLI, Interactive, and Web interfaces.

**Version:** 2.0.6 | **Language:** Go 1.23 | **License:** MIT
**Author:** Ricardo Wagemaker | **Repo:** github.com/gcclinux/scmd

---

## What Is SCMD?

SCMD is a personal/team command knowledge base. Store, search, and retrieve CLI commands, code snippets, and documentation using keyword search, vector similarity, or natural language — from a terminal, interactive shell, or browser.

---

## Three Interfaces

### 1. Interactive CLI (`--cli`)
- Natural language queries ("show me postgresql replication examples")
- 13 slash commands: `/search`, `/add`, `/list`, `/count`, `/delete`, `/show`, `/help`, `/import`, `/run`, `/ai`, `/config`, `/embeddings`, `/generate`
- AI-powered explanations with context-aware responses
- Feedback loop: save or retry AI answers
- Markdown rendering in terminal

### 2. Traditional CLI
- `--search "pattern"` — keyword search with AND/OR logic
- `--save "command" "description"` — store new entries
- `--import file.md` — import markdown documents
- `--copydb` — export database to JSON
- JSON output, scriptable, automation-friendly

### 3. Web UI (`--web` / `--ssl`)
- Browser-based search with real-time results
- Add commands via form with duplicate detection
- AI-generated explanations inline
- Syntax highlighting for code blocks
- SSL/HTTPS support with custom certificates
- Read-only mode (`-block`), headless service mode (`-service`)
- Custom port binding (`-port`)
- Session-based authentication (email + API key, 24h sessions)

---

## AI Integration

### Dual Provider Architecture
| Provider | Type | Use Case |
|----------|------|----------|
| **Ollama** | Local/Remote | Privacy-first, self-hosted, no API costs |
| **Google Gemini** | Cloud API | High quality, free tier (1,500 req/day) |

- Configurable preferred provider via `agent` field
- Automatic fallback: Ollama → Gemini (or vice versa)
- Separate models for chat and embeddings

### Smart Search Pipeline
```
User Query
  → Keyword Extraction (NLP stop-word removal)
    → PostgreSQL ILIKE Search
      → Score & Rank Results (word-match %)
        → Score ≥ 60%? → Return ranked results
        → Score < 60%? → Vector Similarity Search (pgvector / cosine)
          → Still low? → Pure AI Chat Response
```

### Embeddings
- Generate vector embeddings for all stored commands
- Configurable dimensions (384, 768, etc.)
- Batch generation with progress tracking
- Stats dashboard (`--embedding-stats`)
- Supports pgvector (PostgreSQL) and in-memory cosine similarity (SQLite)

---

## Database Support

| Feature | PostgreSQL | SQLite |
|---------|-----------|--------|
| Type | Client-server | File-based |
| Vector Search | pgvector extension | Go cosine similarity |
| Multi-user | Yes | Single-user |
| Setup | `--create-db-postgresql` | `--create-db-sqlite` |
| Location | Remote/local server | `~/.scmd/scmd.db` |
| Scalability | Enterprise | Lightweight |

### Schema
- `id` — Auto-increment primary key
- `key` — Command/code snippet (TEXT)
- `data` — Description/documentation (TEXT)
- `embedding` — Vector (optional, for semantic search)

---

## Search Capabilities

### Pattern Matching
- **AND logic** (spaces): `postgresql replication slave` — all words must match
- **OR logic** (commas): `docker,kubernetes` — any pattern matches
- **Combined**: `postgresql replication,docker backup`
- Case-insensitive, partial word matching
- Searches both command and description fields

### Intelligent Scoring
- Word-match percentage ranking
- Automatic filtering below 60% threshold
- Best-match prioritization
- NLP keyword extraction (removes stop words like "show me", "how to", "please")

---

## Deployment Options

### Platforms
- Windows (AMD64)
- Linux (AMD64, ARM64)
- macOS (Intel, Apple Silicon)

### Docker
- Pre-built image: `gcclinux/scmd:latest`
- Docker Compose included
- Multi-stage build (Go builder → Debian slim runtime)

### Self-Hosted
- Single binary, zero runtime dependencies
- Config at `~/.scmd/config.json`
- Interactive setup wizards for database and AI providers
- Auto-upgrade from GitHub releases (`--upgrade`)

---

## Configuration

Located at `~/.scmd/config.json`:

```json
{
  "agent": "ollama | gemini",
  "db_type": "postgresql | sqlite",
  "db_host": "localhost",
  "db_port": "5432",
  "db_user": "...",
  "db_pass": "...",
  "db_name": "...",
  "tb_name": "scmd",
  "gemini_api": "...",
  "gemini_model": "gemini-2.5-flash-lite",
  "gemini_embedding_model": "gemini-embedding-001",
  "ollama": "localhost",
  "model": "ministral-3:3b",
  "embedding_model": "qwen2.5-coder:1.5b",
  "embedding_dim": "384"
}
```

Environment variables override config file values.

---

## Security

- Session-based web authentication (email + API key)
- 24-hour session expiry with automatic cleanup
- HTTP-only cookies with SameSite protection
- Read-only mode to prevent unauthorized additions
- SSL/TLS with custom certificate support
- Config file permissions (0600)

---

## Content Management

- Store CLI commands, code snippets, scripts, documentation
- Import markdown files with automatic title extraction
- Duplicate detection on add/import
- Delete by ID, list recent, count totals
- Export entire database to JSON (`--copydb`)
- Multi-line command and code block support

---

## Version Management

- Check current vs. remote version (`--version`)
- One-command binary upgrade (`--upgrade`)
- Platform-aware download (auto-detects OS/arch)
- Changelog tracking

---

## Potential Use Cases

| Use Case | Description |
|----------|-------------|
| **DevOps Command Library** | Store and search Docker, Kubernetes, Terraform, Ansible commands |
| **Team Knowledge Base** | Shared PostgreSQL backend for team-wide command reference |
| **Personal Cheat Sheet** | Quick lookup for frequently forgotten CLI syntax |
| **Onboarding Tool** | New team members search for project-specific commands and procedures |
| **Documentation Hub** | Import markdown runbooks and SOPs, search with natural language |
| **AI-Assisted Learning** | Ask natural language questions, get AI explanations with real command context |
| **Code Snippet Manager** | Store reusable code blocks with syntax highlighting |
| **Offline Reference** | SQLite + Ollama for fully offline, air-gapped environments |
| **CI/CD Integration** | Script-friendly CLI with JSON output for pipeline automation |
| **Self-Hosted Wiki Alternative** | Lightweight web UI for internal command documentation |
| **Incident Response** | Quickly find recovery commands during outages |
| **Multi-Language Support** | Store commands for any language/tool (bash, SQL, Python, PowerShell, etc.) |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                    SCMD v2.0.6                  │
├─────────────┬──────────────┬────────────────────┤
│ Interactive  │ Traditional  │     Web UI         │
│    CLI       │    CLI       │  (HTTP/HTTPS)      │
├─────────────┴──────────────┴────────────────────┤
│              Core Engine                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ Search   │ │ Scoring  │ │ Keyword Extract  │ │
│  │ Engine   │ │ System   │ │ (NLP)            │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
├──────────────────────────────────────────────────┤
│              AI Layer                            │
│  ┌──────────────────┐ ┌───────────────────────┐ │
│  │ Ollama (Local)   │ │ Gemini (Cloud)        │ │
│  │ Chat + Embedding │ │ Chat + Embedding      │ │
│  └──────────────────┘ └───────────────────────┘ │
├──────────────────────────────────────────────────┤
│              Data Layer                          │
│  ┌──────────────────┐ ┌───────────────────────┐ │
│  │ PostgreSQL       │ │ SQLite                │ │
│  │ + pgvector       │ │ + cosine similarity   │ │
│  └──────────────────┘ └───────────────────────┘ │
└──────────────────────────────────────────────────┘
```

---

## Key Numbers

- **13** slash commands in interactive mode
- **2** AI providers (Ollama + Gemini)
- **2** database backends (PostgreSQL + SQLite)
- **3** interfaces (Interactive CLI, Traditional CLI, Web UI)
- **5** platforms (Win AMD64, Linux AMD64, Linux ARM64, macOS Intel, macOS ARM)
- **60%** score threshold for smart search fallback to AI
- **24h** session duration for web authentication
- **384** default embedding dimensions
