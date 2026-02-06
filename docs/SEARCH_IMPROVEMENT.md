# Search Improvement - AND Logic

## What Changed

The search functionality has been improved to support **intelligent AND/OR logic**.

### Before (Old Behavior)

Space-separated words were treated as a single pattern:
```
scmd> postgresql replication slave
→ Searched for the exact phrase "postgresql replication slave"
→ No results (because no record has this exact phrase)
```

### After (New Behavior)

Space-separated words use **AND logic** - all words must be present:
```
scmd> postgresql replication slave
→ Searches for records containing ALL three words
→ Finds ID 786 and ID 787 (both contain all three words)
```

## How It Works Now

### AND Logic (Space-Separated)

**All words must be present** in the command or description.

```
scmd> postgresql replication slave

Finds:
✓ ID 786: "Postgresql Replication check on SLAVE Server..."
✓ ID 787: "Postgresql Replication delay check on SLAVE Server..."

Does NOT find:
✗ ID 774: Has "postgresql" and "replication" but not "slave"
✗ ID 785: Has "postgresql" and "replication" but not "slave"
```

### OR Logic (Comma-Separated)

**Any pattern can match** - unchanged from before.

```
scmd> docker,kubernetes,postgresql

Finds:
✓ All docker commands
✓ All kubernetes commands  
✓ All postgresql commands
```

### Combined Logic

You can mix both!

```
scmd> postgresql replication slave,docker backup

Finds:
✓ Records with (postgresql AND replication AND slave)
✓ OR records with (docker AND backup)
```

## Real-World Examples

### Example 1: Your Use Case

**Search:** `postgresql replication`
```
Found 4 results:
- ID 774: Postgresql replication backup...
- ID 785: Postgresql Replication check on Master...
- ID 786: Postgresql Replication check on SLAVE...
- ID 787: Postgresql Replication delay check on SLAVE...
```

**Search:** `postgresql replication slave`
```
Found 2 results:
- ID 786: Postgresql Replication check on SLAVE...
- ID 787: Postgresql Replication delay check on SLAVE...
```

**Search:** `postgresql replication slave delay`
```
Found 1 result:
- ID 787: Postgresql Replication delay check on SLAVE...
```

### Example 2: Docker Commands

**Search:** `docker`
```
Found 500+ results (too many!)
```

**Search:** `docker logs`
```
Found 15 results (better!)
```

**Search:** `docker logs follow`
```
Found 3 results (perfect!)
```

### Example 3: Multiple Technologies

**Search:** `docker,kubernetes`
```
Found 200+ results (all docker OR kubernetes commands)
```

**Search:** `docker logs,kubernetes logs`
```
Found 20 results (docker log commands OR kubernetes log commands)
```

## Benefits

### 1. More Precise Results

**Before:**
```
scmd> postgresql replication slave
→ No results (too specific as a phrase)
```

**After:**
```
scmd> postgresql replication slave
→ 2 results (exactly what you wanted!)
```

### 2. Natural Narrowing

Start broad, then narrow down:
```
scmd> postgresql          → 100 results
scmd> postgresql backup   → 10 results
scmd> postgresql backup docker → 3 results
```

### 3. Intuitive Behavior

Matches how people naturally think:
- "Find commands about postgresql replication on slave"
- All three concepts should be present

### 4. Flexible Searching

- Use spaces to narrow (AND)
- Use commas to broaden (OR)
- Combine both for complex queries

## Technical Details

### SQL Query Generation

**Space-separated (AND):**
```sql
SELECT * FROM scmd 
WHERE (key ILIKE '%postgresql%' OR data ILIKE '%postgresql%')
  AND (key ILIKE '%replication%' OR data ILIKE '%replication%')
  AND (key ILIKE '%slave%' OR data ILIKE '%slave%')
```

**Comma-separated (OR):**
```sql
SELECT * FROM scmd 
WHERE (key ILIKE '%docker%' OR data ILIKE '%docker%')
   OR (key ILIKE '%kubernetes%' OR data ILIKE '%kubernetes%')
```

### Case Insensitivity

All searches are case-insensitive using PostgreSQL's ILIKE:
```
postgresql = POSTGRESQL = PostgreSQL = PoStGrEsQl
```

### Partial Matching

Uses wildcard matching (%) for flexibility:
```
replic → matches replication, replica, replicate, etc.
```

## Migration Notes

### Existing Scripts

If you have scripts using the traditional CLI, they will benefit from the improved search:

**Before:**
```bash
scmd.exe --search "postgresql replication"
# Might have missed some results
```

**After:**
```bash
scmd.exe --search "postgresql replication"
# Now finds all records with both words
```

### Comma Behavior Unchanged

Comma-separated searches work exactly as before:
```bash
scmd.exe --search "docker,kubernetes"
# Same behavior - OR logic
```

## Troubleshooting

### Getting Too Many Results?

Add more specific words:
```
scmd> docker
→ Too many results

scmd> docker compose logs
→ Much better!
```

### Getting No Results?

Remove some words:
```
scmd> postgresql replication slave delay master backup
→ No results (too specific)

scmd> postgresql replication slave delay
→ Found 1 result!
```

### Want OR Logic?

Use commas:
```
scmd> postgresql,mysql,mongodb
→ All database commands
```

## Summary

| Feature | Before | After |
|---------|--------|-------|
| Space-separated | Single phrase search | AND logic (all words) |
| Comma-separated | OR logic | OR logic (unchanged) |
| Precision | Less precise | More precise |
| Flexibility | Limited | High |
| Natural language | Partial | Full support |

**The search is now smarter and more intuitive!** 🎉

## See Also

- [SEARCH_GUIDE.md](SEARCH_GUIDE.md) - Complete search guide
- [INTERACTIVE_MODE.md](INTERACTIVE_MODE.md) - Interactive mode documentation
- [FEATURES.md](FEATURES.md) - All features overview
