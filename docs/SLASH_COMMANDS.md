# Slash Commands Quick Reference

## All Available Slash Commands

### Help & Information

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show help message | `scmd> /help` |
| `/?` | Show help message (alias) | `scmd> /?` |

### Search & Query

| Command | Description | Example |
|---------|-------------|---------|
| `/search <pattern>` | Search for commands | `scmd> /search postgresql replication` |
| `/list` | List recent 10 commands | `scmd> /list` |
| `/count` | Show total commands | `scmd> /count` |

### Add Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/add <cmd> \| <desc>` | Add new command | `scmd> /add docker ps -a \| List containers` |

### Screen & Navigation

| Command | Description | Example |
|---------|-------------|---------|
| `/clear` | Clear the screen | `scmd> /clear` |
| `/cls` | Clear the screen (alias) | `scmd> /cls` |

### Exit

| Command | Description | Example |
|---------|-------------|---------|
| `/exit` | Exit interactive mode | `scmd> /exit` |
| `/quit` | Exit interactive mode (alias) | `scmd> /quit` |
| `/q` | Exit interactive mode (short) | `scmd> /q` |

## Quick Shortcuts (Without Slash)

For convenience, these commands work **with or without** the slash:

| Without Slash | With Slash | Description |
|---------------|------------|-------------|
| `help` or `?` | `/help` or `/?` | Show help message |
| `clear` or `cls` | `/clear` or `/cls` | Clear the screen |
| `exit`, `quit`, or `q` | `/exit`, `/quit`, or `/q` | Exit interactive mode |

## Usage Examples

### Getting Help

```
scmd> /help
[Shows complete help]

scmd> help
[Also shows complete help]

scmd> /?
[Also shows complete help]
```

### Searching

```
scmd> /search postgresql replication
[Searches for commands]

scmd> postgresql replication
[Also searches - no slash needed for direct search]
```

### Adding Commands

```
scmd> /add docker logs -f myapp | Follow application logs
✓ Command added successfully!
```

### Listing Commands

```
scmd> /list
Recent Commands (showing 10 of 1247):
[Shows last 10 commands]
```

### Counting Commands

```
scmd> /count
Total commands in database: 1247
```

### Clearing Screen

```
scmd> /clear
[Screen cleared]

scmd> clear
[Also clears screen]

scmd> /cls
[Also clears screen]
```

### Exiting

```
scmd> /exit
Goodbye!

scmd> exit
Goodbye!

scmd> /quit
Goodbye!

scmd> /q
Goodbye!
```

## Command Syntax Rules

### /search

**Syntax:** `/search <pattern>`

**Pattern types:**
- Space-separated (AND): `/search postgresql replication slave`
- Comma-separated (OR): `/search docker,kubernetes`
- Combined: `/search postgresql replication,docker backup`

**Examples:**
```
scmd> /search docker
scmd> /search docker logs
scmd> /search docker,kubernetes
scmd> /search postgresql replication slave
```

### /add

**Syntax:** `/add <command> | <description>`

**Rules:**
- Use `|` (pipe) to separate command and description
- Both command and description are required
- Checks for duplicates before adding

**Examples:**
```
scmd> /add docker ps -a | List all containers
scmd> /add kubectl get pods -n production | List production pods
scmd> /add git branch -d feature-branch | Delete local branch
```

### /list

**Syntax:** `/list`

**Behavior:**
- Shows last 10 commands
- Displays ID, description, and command preview
- Commands longer than 80 chars are truncated

**Example:**
```
scmd> /list

Recent Commands (showing 10 of 1247):
══════════════════════════════════════════════════════════════

ID: 1238 - List all Docker containers
    docker ps -a

ID: 1239 - Show Kubernetes pods
    kubectl get pods
[...]
```

### /count

**Syntax:** `/count`

**Behavior:**
- Shows total number of commands in database
- Quick way to check database size

**Example:**
```
scmd> /count

Total commands in database: 1247
```

## Error Messages

### Unknown Command

```
scmd> /unknown
Unknown command: /unknown
Type '/help' for available commands
```

### Missing Arguments

```
scmd> /search
Usage: /search <pattern>

scmd> /add
Usage: /add <command> | <description>
Example: /add docker ps -a | List all containers
```

### Invalid Syntax

```
scmd> /add docker ps -a
Error: Use | to separate command and description
Example: /add docker ps -a | List all containers
```

## Tips

1. **Use tab completion** (future feature) - Will make slash commands faster
2. **Shortcuts are faster** - Use `help` instead of `/help` for quick access
3. **Direct search is easiest** - Just type keywords without `/search`
4. **Use /search for exact control** - When you want to be explicit
5. **Combine commands** - Search, then add, then list in one session

## Comparison: Slash vs Non-Slash

### When to Use Slash Commands

- When you want to be explicit: `/search postgresql`
- When adding commands: `/add` (required)
- When listing/counting: `/list`, `/count` (required)
- For consistency: Some prefer always using slashes

### When to Use Without Slash

- Quick help: `help` is faster than `/help`
- Quick exit: `exit` is faster than `/exit`
- Quick clear: `clear` is faster than `/clear`
- Direct search: `postgresql replication` is faster than `/search postgresql replication`

## Summary

| Category | Commands |
|----------|----------|
| **Help** | `/help`, `/?`, `help`, `?` |
| **Search** | `/search <pattern>` or direct keywords |
| **Add** | `/add <cmd> \| <desc>` |
| **List** | `/list` |
| **Count** | `/count` |
| **Clear** | `/clear`, `/cls`, `clear`, `cls` |
| **Exit** | `/exit`, `/quit`, `/q`, `exit`, `quit`, `q` |

**Total slash commands:** 7 main commands + 3 aliases = 10 slash commands
**Total shortcuts:** 3 commands (help, clear, exit) work without slash

All commands are case-sensitive for the slash, but the keywords after are case-insensitive.
