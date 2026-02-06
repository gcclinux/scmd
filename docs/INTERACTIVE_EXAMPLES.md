# Interactive Mode Examples

## Real-World Usage Examples

### Example 1: Finding PostgreSQL Commands

**Your current workflow:**
```bash
.\scmd --search "postgresql replication"
```

**New interactive workflow:**
```
scmd.exe -i

scmd> provide me with postgresql replication on master example

Found 1 result(s) for: postgresql replication master
══════════════════════════════════════════════════════════════

ID: 785
Description: Postgresql Replication check on Master Server inside a docker container
Command:
$ docker exec POSTGRESQL psql -U ricardo -c "SELECT * FROM pg_stat_replication;"
──────────────────────────────────────────────────────────────

scmd> postgresql replication slave

Found 2 result(s) for: postgresql replication slave
══════════════════════════════════════════════════════════════

ID: 786
Description: Postgresql Replication check on SLAVE Server inside a docker container
Command:
$ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT * FROM pg_stat_wal_receiver;"
──────────────────────────────────────────────────────────────

ID: 787
Description: Postgresql Replication delay check on SLAVE Server inside a docker container
Command:
$ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;"
──────────────────────────────────────────────────────────────

scmd> exit
Goodbye!
```

### Example 2: Adding Commands While Working

**Scenario:** You discover a useful command and want to save it immediately

```
scmd> /add docker exec POSTGRESQL psql -U ricardo -c "SELECT version();" | Check PostgreSQL version

✓ Command added successfully!
  Command: docker exec POSTGRESQL psql -U ricardo -c "SELECT version();"
  Description: Check PostgreSQL version

scmd> /search postgresql version

Found 1 result(s) for: postgresql version
══════════════════════════════════════════════════════════════

ID: 1248
Description: Check PostgreSQL version
Command:
docker exec POSTGRESQL psql -U ricardo -c "SELECT version();"
──────────────────────────────────────────────────────────────
```

### Example 3: Natural Language Queries

**Different ways to ask the same thing:**

```
scmd> show me docker commands
scmd> give me docker examples
scmd> how to use docker
scmd> find docker commands
scmd> docker

# All extract "docker" and search for it!
```

### Example 4: Multiple Pattern Search

```
scmd> /search docker,kubernetes,postgresql

Found 47 result(s) for: docker,kubernetes,postgresql
══════════════════════════════════════════════════════════════
[Shows all commands related to docker, kubernetes, or postgresql]
```

### Example 5: Browsing Recent Commands

```
scmd> /list

Recent Commands (showing 10 of 1247):
══════════════════════════════════════════════════════════════

ID: 1238 - List all Docker containers
    docker ps -a

ID: 1239 - Show Kubernetes pods
    kubectl get pods

ID: 1240 - PostgreSQL backup command
    pg_dump -h localhost -U postgres mydb > backup.sql

[... more commands ...]

scmd> /count

Total commands in database: 1247
```

### Example 6: Daily Development Workflow

```
scmd.exe -i

# Morning: Check what commands you have
scmd> /count
Total commands in database: 1247

# Find docker compose commands
scmd> docker compose

Found 12 result(s) for: docker compose
[Results shown...]

# Add a new command you just learned
scmd> /add docker compose logs -f --tail=100 service_name | Follow last 100 lines of service logs

✓ Command added successfully!

# Search for kubernetes
scmd> kubernetes deployment

Found 8 result(s) for: kubernetes deployment
[Results shown...]

# Quick check of recent additions
scmd> /list

# End of day
scmd> exit
Goodbye!
```

### Example 7: Learning New Technologies

```
scmd> show me terraform commands

Found 15 result(s) for: terraform
══════════════════════════════════════════════════════════════
[Terraform commands shown...]

scmd> ansible playbook examples

Found 8 result(s) for: ansible playbook
══════════════════════════════════════════════════════════════
[Ansible commands shown...]

scmd> /add terraform init -backend-config="bucket=my-bucket" | Initialize Terraform with S3 backend

✓ Command added successfully!
```

### Example 8: Team Collaboration

**Team member 1 adds commands:**
```
scmd> /add kubectl rollout restart deployment/myapp | Restart Kubernetes deployment

✓ Command added successfully!
```

**Team member 2 (using same database) finds it:**
```
scmd> kubernetes restart

Found 1 result(s) for: kubernetes restart
══════════════════════════════════════════════════════════════

ID: 1249
Description: Restart Kubernetes deployment
Command:
kubectl rollout restart deployment/myapp
──────────────────────────────────────────────────────────────
```

### Example 9: Complex Searches

```
# Find commands with multiple keywords
scmd> postgresql backup docker

Found 3 result(s) for: postgresql backup docker
══════════════════════════════════════════════════════════════
[Shows commands that match postgresql AND backup AND docker]

# Use slash command for exact pattern
scmd> /search "pg_dump"

Found 5 result(s) for: pg_dump
══════════════════════════════════════════════════════════════
[Shows all pg_dump commands]
```

### Example 10: Quick Reference

```
# Quick help
scmd> help

Available Commands:
──────────────────────────────────────────────────────────────
  /search <pattern>     - Search for commands matching pattern
  /add <cmd> | <desc>   - Add a new command (use | as separator)
  /list                 - List recent commands
  /count                - Show total number of commands
  help or ?             - Show this help message
  clear or cls          - Clear the screen
  exit, quit, or q      - Exit interactive mode
[...]

# Clear screen and start fresh
scmd> clear

# Continue working...
```

## Comparison: Traditional vs Interactive

### Traditional CLI (Multiple Commands)

```bash
# Search for docker
.\scmd.exe --search "docker"

# Search for kubernetes
.\scmd.exe --search "kubernetes"

# Add a command
.\scmd.exe --save "docker ps -a" "List all containers"

# Search again
.\scmd.exe --search "docker ps"
```

**4 separate command executions, 4 database connections**

### Interactive Mode (Single Session)

```
scmd.exe -i

scmd> docker
scmd> kubernetes
scmd> /add docker ps -a | List all containers
scmd> docker ps
scmd> exit
```

**1 session, 1 database connection, faster workflow**

## Tips for Maximum Productivity

### 1. Keep It Open
Leave interactive mode running in a terminal window for quick access throughout the day.

### 2. Use Natural Language
Don't overthink it - just type what you're looking for:
- "show me git commands"
- "postgresql backup"
- "how to restart nginx"

### 3. Add Commands Immediately
When you discover a useful command, add it right away:
```
scmd> /add <command> | <description>
```

### 4. Use /list to Browse
Periodically check what's been added:
```
scmd> /list
```

### 5. Combine with Traditional CLI
Use interactive mode for exploration, traditional CLI for scripts:

**Interactive:** Daily usage, learning, exploring
**Traditional CLI:** Scripts, automation, CI/CD

### 6. Leverage Tab Completion (Future)
Once implemented, tab completion will make it even faster!

## Common Patterns

### Pattern 1: Find and Add
```
scmd> docker logs
[Review results]
scmd> /add docker logs -f --tail=100 myapp | Follow last 100 lines of app logs
```

### Pattern 2: Multiple Related Searches
```
scmd> git branch
scmd> git merge
scmd> git rebase
scmd> git cherry-pick
```

### Pattern 3: Verify Addition
```
scmd> /add kubectl get pods -n production | List production pods
scmd> /search kubectl get pods
[Verify it was added]
```

### Pattern 4: Quick Count Check
```
scmd> /count
Total commands in database: 1247

[Add some commands...]

scmd> /count
Total commands in database: 1250
```

## Keyboard Shortcuts

- `Ctrl+C` - Exit interactive mode
- `Enter` on empty line - Safe (does nothing)
- Type `exit`, `quit`, or `q` - Exit gracefully

## Error Handling Examples

### No Results
```
scmd> xyz123abc

No results found for: xyz123abc
Try different keywords or use /search with comma-separated patterns
```

### Duplicate Command
```
scmd> /add docker ps -a | List containers

Error: This command already exists in the database
```

### Invalid Syntax
```
scmd> /add docker ps -a

Error: Use | to separate command and description
Example: /add docker ps -a | List all containers
```

## Integration Examples

### With Shell Scripts
```bash
#!/bin/bash
# Start interactive mode from script
scmd.exe -i
```

### With Aliases
```bash
# Add to .bashrc or .zshrc
alias scmd='scmd.exe -i'

# Now just type:
scmd
```

### With Screen/Tmux
```bash
# Create a dedicated window
screen -S scmd scmd.exe -i

# Or tmux
tmux new-session -s scmd 'scmd.exe -i'
```

## Summary

Interactive mode makes SCMD:
- **Faster** - One session, multiple queries
- **Easier** - Natural language support
- **More powerful** - Slash commands
- **More productive** - Add commands on the fly
- **More intuitive** - User-friendly prompts

**Start using it today:**
```bash
scmd.exe -i
```
