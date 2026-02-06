# SCMD Search Guide

## Understanding Search Behavior

SCMD now supports intelligent search with two different modes: **AND** logic and **OR** logic.

### AND Logic (Space-Separated)

When you search with **space-separated words**, ALL words must be present in the result.

**Example:**
```
scmd> postgresql replication slave
```

This will find records where the command OR description contains:
- "postgresql" **AND**
- "replication" **AND**
- "slave"

**Results:**
```
ID: 786
Description: Postgresql Replication check on SLAVE Server inside a docker container
Command: $ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT * FROM pg_stat_wal_receiver;"

ID: 787
Description: Postgresql Replication delay check on SLAVE Server inside a docker container
Command: $ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;"
```

Both results contain all three words: "postgresql", "replication", and "slave" (or "SLAVE").

### OR Logic (Comma-Separated)

When you search with **comma-separated patterns**, ANY pattern can match.

**Example:**
```
scmd> docker,kubernetes,postgresql
```

This will find records that contain:
- "docker" **OR**
- "kubernetes" **OR**
- "postgresql"

**Results:** All commands related to docker, kubernetes, or postgresql.

### Combined Logic

You can combine both! Use commas for OR and spaces within each pattern for AND.

**Example:**
```
scmd> postgresql replication,docker backup
```

This will find records that contain:
- ("postgresql" **AND** "replication") **OR**
- ("docker" **AND** "backup")

## Search Examples

### Example 1: Narrow Down Results

**Search:** `postgresql replication`
**Finds:** 4 results (all postgresql replication commands)

**Search:** `postgresql replication slave`
**Finds:** 2 results (only slave-related replication commands)

**Search:** `postgresql replication slave delay`
**Finds:** 1 result (the specific delay check command)

### Example 2: Multiple Topics

**Search:** `docker,kubernetes`
**Finds:** All docker commands + all kubernetes commands

**Search:** `docker logs,kubernetes logs`
**Finds:** Docker log commands + Kubernetes log commands

### Example 3: Specific Commands

**Search:** `git branch`
**Finds:** Commands containing both "git" and "branch"

**Search:** `git branch delete`
**Finds:** Commands about deleting git branches

**Search:** `git branch,git merge,git rebase`
**Finds:** Commands about branches OR merges OR rebases

## Search Tips

### 1. Start Broad, Then Narrow

```
scmd> postgresql
[Too many results]

scmd> postgresql replication
[Better - 4 results]

scmd> postgresql replication slave
[Perfect - 2 results]

scmd> postgresql replication slave delay
[Exact - 1 result]
```

### 2. Use Commas for Multiple Topics

```
scmd> docker,kubernetes,terraform
[All commands for these three tools]
```

### 3. Combine for Complex Searches

```
scmd> docker compose,kubernetes deployment
[Docker compose commands OR Kubernetes deployment commands]
```

### 4. Case Doesn't Matter

```
scmd> POSTGRESQL REPLICATION
scmd> postgresql replication
scmd> PostgreSQL Replication
```

All three searches return the same results!

### 5. Word Order Doesn't Matter

```
scmd> postgresql replication slave
scmd> slave replication postgresql
scmd> replication slave postgresql
```

All three searches return the same results!

## Search Behavior Details

### What Gets Searched

Both the **command** and **description** fields are searched:

**Example Record:**
```
ID: 786
Description: Postgresql Replication check on SLAVE Server inside a docker container
Command: $ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT * FROM pg_stat_wal_receiver;"
```

**Search:** `postgresql replication slave`

**Matches because:**
- "Postgresql" is in the description ✓
- "Replication" is in the description ✓
- "SLAVE" is in both description and command ✓

### Partial Word Matching

SCMD uses partial matching (ILIKE with %), so:

**Search:** `replic`
**Finds:** Commands with "replication", "replica", "replicate", etc.

**Search:** `post`
**Finds:** Commands with "postgresql", "postgres", "post", etc.

### Special Characters

Special characters are treated as literal characters:

**Search:** `pg_stat_replication`
**Finds:** Commands containing "pg_stat_replication"

## Common Search Patterns

### Pattern 1: Technology + Action

```
scmd> docker restart
scmd> kubernetes scale
scmd> postgresql backup
scmd> nginx reload
```

### Pattern 2: Multiple Technologies

```
scmd> docker,kubernetes
scmd> postgresql,mysql,mongodb
scmd> git,svn
```

### Pattern 3: Specific Feature

```
scmd> postgresql replication master
scmd> docker compose logs
scmd> kubernetes deployment rollout
```

### Pattern 4: Troubleshooting

```
scmd> error logs
scmd> check status
scmd> debug connection
```

## Interactive Mode vs Traditional CLI

### Interactive Mode

```
scmd> postgresql replication slave
[Results displayed with formatting]

scmd> /search postgresql replication slave
[Same results]
```

### Traditional CLI

```bash
# Space-separated (AND logic)
scmd.exe --search "postgresql replication slave"

# Comma-separated (OR logic)
scmd.exe --search "docker,kubernetes"
```

Both modes use the same search logic!

## Troubleshooting Searches

### No Results Found

**Problem:** `postgresql replication slave delay master`
**Result:** No results

**Why?** Too specific - no single record contains ALL five words.

**Solution:** Remove some words:
```
scmd> postgresql replication slave delay
[Found 1 result]
```

### Too Many Results

**Problem:** `docker`
**Result:** 500+ results

**Why?** Too broad - many commands use docker.

**Solution:** Add more specific words:
```
scmd> docker logs
[Found 15 results]

scmd> docker logs follow
[Found 3 results]
```

### Wrong Results

**Problem:** Searching for "master" but getting "mastercard" results

**Solution:** Add more context words:
```
scmd> postgresql master
[More relevant results]
```

## Advanced Search Techniques

### 1. Exclude False Positives

If you're getting unwanted results, add more specific terms:

```
scmd> log
[Too many results including "login", "logical", etc.]

scmd> docker log
[Better - only docker logging commands]
```

### 2. Find Related Commands

Use OR logic to find related commands:

```
scmd> backup,restore,dump
[All backup-related commands]
```

### 3. Find Commands by Tool

```
scmd> kubectl
scmd> docker exec
scmd> psql
```

### 4. Find Commands by Action

```
scmd> check status
scmd> list all
scmd> show running
```

## Search Performance

- **Fast:** PostgreSQL indexes make searches instant
- **Efficient:** Case-insensitive matching (ILIKE)
- **Scalable:** Works with millions of commands
- **Smart:** AND/OR logic for precise results

## Summary

| Search Type | Syntax | Logic | Example |
|-------------|--------|-------|---------|
| AND (all words) | Space-separated | All words must match | `postgresql replication slave` |
| OR (any pattern) | Comma-separated | Any pattern can match | `docker,kubernetes` |
| Combined | Mix both | Complex queries | `docker logs,kubernetes logs` |

**Remember:**
- Spaces = AND (narrow down)
- Commas = OR (broaden)
- Case doesn't matter
- Order doesn't matter
- Partial matching works

Happy searching! 🔍
