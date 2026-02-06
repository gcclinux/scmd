# Intelligent Scoring System

## Overview

SCMD now uses an intelligent scoring system that ranks search results by how well they match your query. This ensures the most relevant commands appear first!

## How It Works

### 1. Word Extraction

Your query is broken down into meaningful words:

**Query:** "check Postgresql Replication delay on SLAVE Server"

**Extracted words:** `["check", "postgresql", "replication", "delay", "slave", "server"]`

Common words removed: "on", "the", "a", etc.

### 2. Scoring

Each command is scored based on how many query words it contains:

```
Score = (Matched Words / Total Query Words) × 100%
```

**Example:**

Query: "check postgresql replication delay slave server" (6 words)

Command: "Postgresql Replication delay check on SLAVE Server"
- Contains: postgresql ✓, replication ✓, delay ✓, check ✓, slave ✓, server ✓
- Score: 6/6 = **100%** ⭐

Command: "Docker container logs"
- Contains: none
- Score: 0/6 = **0%**

### 3. Smart Decision

**If score ≥ 50%:** Use PostgreSQL results (fast, accurate)
**If score < 50%:** Try vector search with Ollama (semantic understanding)

## Example Scenarios

### Scenario 1: Perfect Match

```
Query: "check postgresql replication slave"
Words: ["check", "postgresql", "replication", "slave"]

Result ID 787:
Description: "Postgresql Replication delay check on SLAVE Server"
Match: 4/4 words = 100% ✓

Decision: Use PostgreSQL results (no AI needed)
```

### Scenario 2: Good Match

```
Query: "postgresql replication monitoring"
Words: ["postgresql", "replication", "monitoring"]

Result ID 786:
Description: "Postgresql Replication check on SLAVE Server"
Match: 2/3 words = 67% ✓

Decision: Use PostgreSQL results + AI explanation
```

### Scenario 3: Weak Match

```
Query: "how to monitor database performance"
Words: ["monitor", "database", "performance"]

Result ID 786:
Description: "Postgresql Replication check on SLAVE Server"
Match: 0/3 words = 0% ✗

Decision: Try vector search (semantic similarity)
```

## Benefits

### 1. Faster Results
- Good matches (≥50%) skip vector search
- Direct PostgreSQL query is instant
- AI only invoked when needed

### 2. Better Accuracy
- Exact keyword matches ranked highest
- Relevant results appear first
- Less irrelevant results

### 3. Transparent Scoring
- See match percentage for each result
- Understand why results were returned
- Know which words matched

## Search Flow

```
User Query
    ↓
Extract Words (remove common words)
    ↓
Traditional PostgreSQL Search
    ↓
Score All Results
    ↓
Best Match ≥ 50%?
    ↓
YES → Use PostgreSQL Results
    ↓
    Invoke AI for Explanation (optional)
    ↓
    Show Results with Scores
    
NO → Try Vector Search
    ↓
    Combine Vector + Traditional
    ↓
    Invoke AI for Explanation
    ↓
    Show Results
```

## Output Format

```
scmd> check postgresql replication delay slave

✓ Found good matches in database

Found 2 result(s) for: check postgresql replication delay slave
(Showing 2 results with matches, filtered 8 with 0% match)
(Best match: 100% - 6/6 words matched)
══════════════════════════════════════════════════════════════

ID: 787 (Match: 100%)
Description: Postgresql Replication delay check on SLAVE Server
Command:
$ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;"
──────────────────────────────────────────────────────────────

ID: 786 (Match: 83%)
Description: Postgresql Replication check on SLAVE Server
Command:
$ docker exec POSTGRESQL-SLAVE psql -U ricardo -c "SELECT * FROM pg_stat_wal_receiver;"
──────────────────────────────────────────────────────────────
```

**Note:** Results with 0% match are automatically filtered out to show only relevant commands.

## Scoring Thresholds

| Score | Meaning | Action |
|-------|---------|--------|
| 100% | Perfect match | Use immediately, no AI needed |
| 75-99% | Excellent match | Use PostgreSQL, AI optional |
| 50-74% | Good match | Use PostgreSQL, AI recommended |
| 25-49% | Weak match | Try vector search |
| 0-24% | Poor match | Vector search only |

## Word Filtering

### Removed Words (Stop Words)
Common words that don't add meaning:
- Question words: how, what, where, when, why, which
- Articles: a, an, the
- Pronouns: i, you, me
- Verbs: is, are, can, do
- Others: to, for, with, on, in

### Kept Words
Meaningful words that help matching:
- Technical terms: postgresql, docker, kubernetes
- Actions: check, monitor, list, show
- Concepts: replication, backup, logs
- Specifics: slave, master, delay

### Minimum Length
Words must be ≥ 3 characters to be considered.

## Advanced Features

### 1. Partial Word Matching
"replication" matches "replicat", "replica", etc.

### 2. Case Insensitive
"PostgreSQL" = "postgresql" = "POSTGRESQL"

### 3. Multi-Field Search
Searches both command and description fields

### 4. Sorted Results
Results automatically sorted by score (highest first)

## Configuration

The 50% threshold is hardcoded but can be adjusted in `ollama.go`:

```go
// Check if we have good matches (50%+ score)
hasGoodMatches := HasGoodMatches(scored, 50)  // Change 50 to adjust
```

## Performance

### With Good Matches (≥50%)
- **Speed:** <100ms (PostgreSQL only)
- **Accuracy:** Excellent (exact matches)
- **AI:** Optional (for explanation)

### With Weak Matches (<50%)
- **Speed:** 1-3s (vector search + AI)
- **Accuracy:** Good (semantic similarity)
- **AI:** Always invoked

## Tips

### Get Better Matches
1. **Use specific terms:** "postgresql replication" vs "database sync"
2. **Include key words:** "check", "monitor", "list"
3. **Be concise:** Fewer words = higher match percentage

### Example Queries

**Good:**
- "postgresql replication slave"
- "docker logs container"
- "kubernetes deployment scale"

**Less Good:**
- "how do I check if my database is replicating properly"
- "show me some commands for containers"

## Summary

✅ **Intelligent scoring** - Ranks by relevance
✅ **50% threshold** - Smart AI invocation
✅ **Transparent** - See match percentages
✅ **Fast** - Skip AI when not needed
✅ **Accurate** - Best results first

Your searches are now smarter and faster! 🚀
