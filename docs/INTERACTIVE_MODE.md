# SCMD Interactive CLI Mode

## Overview

The new Interactive CLI Mode provides a powerful, user-friendly interface for searching and managing commands. It supports both natural language queries and traditional search patterns.

## Starting Interactive Mode

Launch interactive mode using any of these commands:

```bash
scmd.exe --interactive
scmd.exe -i
scmd.exe --cli
```

## Features

### 1. Natural Language Queries

Ask questions naturally, and SCMD will extract keywords and search for you:

```
scmd> provide me with postgresql replication example
scmd> show me docker commands
scmd> how to check kubernetes pods
scmd> find git commands for branches
```

The system automatically removes common question words and extracts the relevant search terms.

### 2. Direct Keyword Search

Type keywords directly without any command prefix:

```
scmd> postgresql replication
scmd> docker,kubernetes
scmd> git branch
```

### 3. Slash Commands

Use slash commands for specific actions:

#### /help or /? - Show help
```
scmd> /help
scmd> /?
```

#### /search - Search for commands
```
scmd> /search postgresql replication
scmd> /search docker,kubernetes
```

#### /add - Add a new command
```
scmd> /add docker ps -a | List all containers
scmd> /add kubectl get pods | Show all Kubernetes pods
```

Note: Use the pipe symbol `|` to separate command and description.

#### /list - Show recent commands
```
scmd> /list
```

Shows the 10 most recent commands in the database.

#### /count - Show total commands
```
scmd> /count
```

Displays the total number of commands in the database.

#### /clear or /cls - Clear screen
```
scmd> /clear
scmd> /cls
```

#### /exit, /quit, or /q - Exit interactive mode
```
scmd> /exit
scmd> /quit
scmd> /q
```

### 4. Quick Shortcuts (Without Slash)

For convenience, these commands also work without the slash:

- `help` or `?` - Show help message
- `clear` or `cls` - Clear the screen
- `exit`, `quit`, or `q` - Exit interactive mode

## Example Session

```
╔════════════════════════════════════════════════════════════════╗
║          SCMD Interactive CLI - PostgreSQL Edition            ║
║                    Version 1.3.8                              ║
╚════════════════════════════════════════════════════════════════╝

Type 'help' for available commands or 'exit' to quit
You can search using:
  - Natural language: provide me with postgresql replication example
  - Search command: /search postgresql replication
  - Direct pattern: postgresql replication

scmd> provide me with postgresql replication on master example

Found 1 result(s) for: postgresql replication master
══════════════════════════════════════════════════════════════

ID: 785
Description: Postgresql Replication check on Master Server inside a docker container
Command:
$ docker exec POSTGRESQL psql -U ricardo -c "SELECT * FROM pg_stat_replication;"
──────────────────────────────────────────────────────────────

scmd> /search docker,kubernetes

Found 15 result(s) for: docker,kubernetes
══════════════════════════════════════════════════════════════
[Results displayed here...]

scmd> /add docker logs -f mycontainer | Follow container logs in real-time

✓ Command added successfully!
  Command: docker logs -f mycontainer
  Description: Follow container logs in real-time

scmd> /count

Total commands in database: 1247

scmd> exit
Goodbye!
```

## Search Behavior

### Natural Language Processing

The system removes common question words and phrases:
- "show me", "give me", "provide me with"
- "how to", "how do i"
- "what is", "what are"
- "can you", "please"
- "i need", "i want"
- "looking for", "search for"
- "example", "examples"
- "command", "commands"

### Pattern Matching

- Case-insensitive search using PostgreSQL ILIKE
- Searches both command text and descriptions
- Supports comma-separated patterns for multiple terms
- Results are ordered by ID

## Comparison with Traditional CLI

### Traditional CLI
```bash
# Search
scmd.exe --search "postgresql replication"

# Add command
scmd.exe --save "docker ps -a" "List all containers"
```

### Interactive Mode
```
scmd> postgresql replication
scmd> /add docker ps -a | List all containers
```

## Benefits

1. **No need to remember exact syntax** - Ask questions naturally
2. **Faster workflow** - Stay in one session for multiple queries
3. **Immediate feedback** - See results instantly
4. **Easy command management** - Add, search, and list in one place
5. **User-friendly** - Clear prompts and formatted output

## Tips

1. **Use natural language** for quick searches without worrying about syntax
2. **Use /search** when you want to be explicit about searching
3. **Use /add** to quickly add commands you discover
4. **Use /list** to browse recent additions
5. **Use comma-separated patterns** to search multiple terms at once

## Keyboard Shortcuts

- `Ctrl+C` - Exit interactive mode (alternative to typing 'exit')
- `Enter` on empty line - Does nothing (safe to press)

## Error Handling

The interactive mode provides clear error messages:

- **No results found** - Suggests trying different keywords
- **Duplicate command** - Warns when trying to add existing command
- **Database errors** - Shows connection or query errors
- **Invalid syntax** - Provides usage examples

## Integration with Existing Features

Interactive mode uses the same PostgreSQL backend as:
- Traditional CLI search (`--search`)
- Traditional CLI save (`--save`)
- Web interface

All commands added in interactive mode are immediately available in all interfaces.

## Performance

- **Fast startup** - Connects to PostgreSQL once at launch
- **Instant searches** - Direct database queries
- **Efficient** - Reuses database connection for all operations
- **Responsive** - No delays between commands

## Future Enhancements

Potential future features:
- Command history (up/down arrows)
- Auto-completion
- Fuzzy search
- Command editing
- Batch operations
- Export/import commands
- Favorites/bookmarks

## Troubleshooting

### "Failed to connect to database"
Ensure your `.env` file has correct PostgreSQL credentials and the database is running.

### "No results found"
Try:
- Different keywords
- Broader search terms
- Check if data exists with `/count`

### Commands not appearing
Verify:
- Command was added successfully (look for ✓ message)
- Database connection is stable
- Table name in `.env` is correct

## Support

For issues or questions about interactive mode:
1. Type `help` in interactive mode
2. Check this documentation
3. Review POSTGRESQL_MIGRATION.md for database setup
4. Test connection with `test_connection.go`
