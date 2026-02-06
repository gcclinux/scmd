# SCMD Features Overview

## Three Ways to Use SCMD

### 1. Interactive CLI Mode 🆕

**Launch:** `scmd.exe --interactive` or `scmd.exe -i`

**Best for:** Daily command-line usage, quick searches, natural language queries

**Features:**
- ✓ Natural language queries
- ✓ Direct keyword search
- ✓ Slash commands (/help, /search, /add, /list, /count, /exit)
- ✓ Quick shortcuts (help, clear, exit work without /)
- ✓ Persistent session
- ✓ Formatted output
- ✓ User-friendly prompts
- ✓ Built-in help

**Example:**
```
scmd> provide me with postgresql replication on master example
scmd> /add docker logs -f myapp | Follow application logs
scmd> /list
scmd> exit
```

### 2. Traditional CLI

**Launch:** `scmd.exe --search "pattern"` or `scmd.exe --save "cmd" "desc"`

**Best for:** Scripting, automation, one-off queries

**Features:**
- ✓ Single command execution
- ✓ Scriptable
- ✓ Exit after execution
- ✓ JSON output support
- ✓ Comma-separated patterns

**Example:**
```bash
scmd.exe --search "postgresql replication"
scmd.exe --save "docker ps -a" "List all containers"
```

### 3. Web Interface

**Launch:** `scmd.exe --web` or `scmd.exe --web -port 8080`

**Best for:** Team sharing, visual browsing, remote access

**Features:**
- ✓ Browser-based interface
- ✓ Visual search results
- ✓ Add commands via form
- ✓ Syntax highlighting for code
- ✓ SSL/HTTPS support
- ✓ Read-only mode (-block)
- ✓ Service mode (no browser launch)

**Example:**
```bash
scmd.exe --web -port 3333
scmd.exe --web -port 8080 -block
scmd.exe --ssl -port 443 cert.pem key.pem
```

## Feature Comparison

| Feature | Interactive CLI | Traditional CLI | Web Interface |
|---------|----------------|-----------------|---------------|
| Natural Language | ✓ | ✗ | ✗ |
| Pattern Search | ✓ | ✓ | ✓ |
| Add Commands | ✓ | ✓ | ✓ |
| List Commands | ✓ | ✗ | ✓ |
| Count Commands | ✓ | ✗ | ✗ |
| Persistent Session | ✓ | ✗ | ✓ |
| Scriptable | ✗ | ✓ | ✗ |
| Remote Access | ✗ | ✗ | ✓ |
| SSL/HTTPS | ✗ | ✗ | ✓ |
| Syntax Highlighting | ✓ | ✓ | ✓ |
| Multi-user | ✗ | ✗ | ✓ |

## Search Capabilities

### Pattern Matching

All modes support:
- Case-insensitive search (PostgreSQL ILIKE)
- Multiple patterns (comma-separated)
- Search in both commands and descriptions
- Ordered results by ID

**Examples:**
```
postgresql replication          → Single pattern
docker,kubernetes              → Multiple patterns
git branch,git merge           → Related patterns
```

### Natural Language (Interactive Mode Only)

Automatically extracts keywords from questions:

**Input:** "provide me with postgresql replication on master example"
**Extracted:** "postgresql replication master"

**Input:** "how to check docker container logs"
**Extracted:** "check docker container logs"

## Command Management

### Adding Commands

**Interactive Mode:**
```
scmd> /add docker ps -a | List all containers
```

**Traditional CLI:**
```bash
scmd.exe --save "docker ps -a" "List all containers"
```

**Web Interface:**
- Fill in form fields
- Click "Add Command" button

### Duplicate Detection

All modes check for duplicate commands before adding:
- Exact command match
- Case-sensitive comparison
- Prevents duplicate entries

## Output Formats

### Interactive Mode
```
Found 1 result(s) for: postgresql replication
══════════════════════════════════════════════════════════════

ID: 785
Description: Postgresql Replication check on Master Server
Command:
$ docker exec POSTGRESQL psql -U ricardo -c "SELECT * FROM pg_stat_replication;"
──────────────────────────────────────────────────────────────
```

### Traditional CLI
```json
{"id": 785, "key": "$ docker exec POSTGRESQL psql...", "data": "Postgresql Replication check..."}
```

### Web Interface
```
----------------------------------------------------------------------
# ID: 785
# Description: "Postgresql Replication check on Master Server"
# Command: $ docker exec POSTGRESQL psql -U ricardo -c "SELECT * FROM pg_stat_replication;"
```

## Database Backend

All modes use the same PostgreSQL backend:
- Shared database connection
- Consistent data across all interfaces
- Real-time updates
- ACID compliance

**Configuration:** `.env` file
```env
DB_HOST=192.168.1.4
DB_PORT=5432
DB_USER=user_name
DB_PASS=password
DB_NAME=database_name
TB_NAME=scmd
```

## Use Cases

### Interactive Mode

**Daily Development:**
```
scmd> show me docker commands
scmd> kubernetes deployment
scmd> /add kubectl scale deployment myapp --replicas=3 | Scale deployment
```

**Learning & Exploration:**
```
scmd> how to use git rebase
scmd> postgresql backup commands
scmd> /list
```

### Traditional CLI

**Shell Scripts:**
```bash
#!/bin/bash
result=$(scmd.exe --search "backup")
echo "$result" | jq '.[] | .key'
```

**Automation:**
```bash
# Add command from script
scmd.exe --save "$(cat command.txt)" "$(cat description.txt)"
```

### Web Interface

**Team Knowledge Base:**
- Share commands across team
- Browse commands visually
- Add commands via form
- Access from any device

**Documentation:**
- Embed in internal wiki
- Reference in documentation
- Share via URL
- Read-only mode for viewers

## Advanced Features

### SSL/HTTPS (Web Interface)

```bash
scmd.exe --ssl -port 443 certificate.pem privkey.pem
scmd.exe --ssl -port 443 -service certificate.pem privkey.pem
```

### Service Mode (Web Interface)

Run without launching browser:
```bash
scmd.exe --web -port 3333 -service
scmd.exe --ssl -port 443 -service cert.pem key.pem
```

### Read-Only Mode (Web Interface)

Disable command addition:
```bash
scmd.exe --web -block
scmd.exe --web -port 8080 -block
```

### Code Detection

All modes automatically detect and format code:
- Functions
- Scripts
- Multi-line commands
- Proper indentation

## Performance

### Interactive Mode
- Fast startup (single DB connection)
- Instant searches
- No overhead between commands
- Efficient for multiple queries

### Traditional CLI
- Quick execution
- Minimal overhead
- Ideal for scripting
- One connection per command

### Web Interface
- Persistent connection
- Multiple concurrent users
- Efficient connection pooling
- Suitable for team use

## Migration from SQLite

All features work with PostgreSQL:
- Same command syntax
- Same search patterns
- Same output formats
- Enhanced performance
- Better scalability

## Future Enhancements

Planned features:
- Command history (interactive mode)
- Auto-completion
- Fuzzy search
- Command editing
- Batch operations
- Export/import
- Favorites/bookmarks
- API endpoints
- Command versioning
- Tags and categories

## Getting Started

1. **Configure database** - Edit `.env` file
2. **Choose your mode:**
   - Interactive: `scmd.exe -i`
   - CLI: `scmd.exe --search "pattern"`
   - Web: `scmd.exe --web`
3. **Start searching!**

## Documentation

- [INTERACTIVE_MODE.md](INTERACTIVE_MODE.md) - Interactive CLI guide
- [POSTGRESQL_MIGRATION.md](POSTGRESQL_MIGRATION.md) - Database setup
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [UPGRADE_SUMMARY.md](UPGRADE_SUMMARY.md) - What's new
- [README.md](README.md) - Main documentation
