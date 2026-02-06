# Ollama Integration Guide

## Overview

SCMD now features intelligent AI-powered search using Ollama! When Ollama is available, SCMD automatically enhances your searches with:

- **Vector similarity search** - Find semantically similar commands, not just keyword matches
- **AI-generated explanations** - Get context and explanations for commands
- **Natural language understanding** - Ask questions naturally
- **Automatic fallback** - Seamlessly falls back to traditional search if Ollama is unavailable

## Configuration

### 1. Install and Run Ollama

First, install Ollama from [ollama.ai](https://ollama.ai) and start the server.

### 2. Configure .env File

Add these settings to your `.env` file:

```env
# Ollama Configuration
OLLAMA=192.168.0.78    # Ollama server IP (or localhost)
MODEL=dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m  # Your Ollama model

# Embedding Configuration (must match your vector dimension)
EMBEDDING_MODEL=all-MiniLM-L6-v2
EMBEDDING_DIM=384
```

### 3. Database Setup

Your PostgreSQL table must have the `embedding` column:

```sql
ALTER TABLE scmd ADD COLUMN embedding vector(384);

-- Create index for fast vector search
CREATE INDEX idx_scmd_embedding 
ON scmd USING hnsw (embedding vector_cosine_ops);
```

## How It Works

### Automatic Detection

When you start SCMD interactive mode, it automatically:
1. Checks if Ollama is available
2. Tests the connection
3. Enables AI features if successful
4. Falls back to traditional search if not

```
scmd.exe -i

✓ Ollama available at 192.168.0.78:11434 (model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m)
```

### Vector Similarity Search

Instead of just matching keywords, Ollama creates embeddings (vector representations) of your query and finds semantically similar commands.

**Traditional Search:**
```
Query: "postgresql replication"
Matches: Commands containing both "postgresql" AND "replication"
```

**Vector Search with Ollama:**
```
Query: "postgresql replication"
Matches: Commands about:
  - PostgreSQL replication
  - Database replication
  - Master-slave setup
  - Streaming replication
  - Replication monitoring
  (Even if they don't contain exact keywords!)
```

### AI-Generated Explanations

When Ollama is active, you get AI-generated explanations along with your search results:

```
scmd> how to check postgresql replication

🤖 AI Assistant:
══════════════════════════════════════════════════════════════
To check PostgreSQL replication, you can use the following commands:

1. On the Master server, use `SELECT * FROM pg_stat_replication;` 
   to see active replication connections and their status.

2. On the Slave server, use `SELECT * FROM pg_stat_wal_receiver;` 
   to check the WAL receiver status.

3. To check replication lag, run the delay check command which 
   calculates the time difference between now and the last 
   transaction replay.

All these commands can be executed inside Docker containers 
using `docker exec`.
══════════════════════════════════════════════════════════════

Found 3 result(s) for: how to check postgresql replication
══════════════════════════════════════════════════════════════
[Command results follow...]
```

## Usage Examples

### Example 1: Natural Language Query

```
scmd> show me how to backup postgresql database

🤖 AI Assistant:
══════════════════════════════════════════════════════════════
To backup a PostgreSQL database, you can use pg_basebackup for 
replication backups. The command shown uses Docker to run 
pg_basebackup with the following key options:
- -h: specifies the host
- -U: specifies the user
- -D: specifies the data directory
- -Fp: plain format
- -Xs: stream WAL
- -P: show progress
- -R: write recovery configuration
══════════════════════════════════════════════════════════════

Found 1 result(s):
[Results...]
```

### Example 2: Semantic Search

```
scmd> container logs

# Finds commands about:
# - docker logs
# - kubectl logs
# - container debugging
# - log streaming
# Even if they don't contain "container logs" exactly!
```

### Example 3: Conceptual Search

```
scmd> monitor database performance

# Finds commands about:
# - pg_stat_replication
# - replication lag
# - database status checks
# - performance monitoring
```

## Slash Commands

### /ai - Check Ollama Status

```
scmd> /ai

✓ Ollama is available and active
  Host: 192.168.0.78
  Model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m

AI-enhanced search is automatically used when available.
Features:
  - Vector similarity search for better relevance
  - AI-generated explanations and context
  - Automatic fallback to traditional search if needed
```

## Fallback Behavior

SCMD is designed to work seamlessly with or without Ollama:

### When Ollama is Available
- ✓ Vector similarity search
- ✓ AI-generated explanations
- ✓ Better semantic understanding
- ✓ More relevant results

### When Ollama is Unavailable
- ✓ Traditional keyword search (AND/OR logic)
- ✓ Fast and reliable
- ✓ No AI explanations
- ✓ Exact keyword matching

**Automatic Fallback:**
If Ollama fails during a search, SCMD automatically falls back to traditional search:

```
⚠ Ollama search failed, falling back to traditional search: connection refused
Found 4 result(s) for: postgresql replication
[Results using traditional search...]
```

## Performance

### Vector Search
- **Speed:** Slightly slower than keyword search (embedding generation + vector similarity)
- **Accuracy:** Much better semantic understanding
- **Best for:** Natural language queries, conceptual searches

### Traditional Search
- **Speed:** Very fast (direct PostgreSQL query)
- **Accuracy:** Exact keyword matching
- **Best for:** Specific keyword searches, known commands

## Troubleshooting

### Ollama Not Detected

```
scmd> /ai

⚠ Ollama is not available
  Host: 192.168.0.78
  Model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m

Make sure Ollama is running and accessible.
```

**Solutions:**
1. Check if Ollama is running: `curl http://192.168.0.78:11434/api/tags`
2. Verify OLLAMA setting in `.env`
3. Check firewall settings
4. Ensure model is pulled: `ollama pull dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m`

### Slow Searches

If searches are slow with Ollama:
1. Use a smaller/faster model
2. Run Ollama locally instead of remote
3. Ensure vector index exists: `CREATE INDEX idx_scmd_embedding ON scmd USING hnsw (embedding vector_cosine_ops);`

### No AI Explanations

If you get results but no AI explanations:
- Ollama chat API might be failing
- Check Ollama logs
- Try a different model
- SCMD will still show search results

### Vector Dimension Mismatch

```
Error: vector dimension mismatch
```

**Solution:** SCMD automatically adjusts dimensions! Ensure `EMBEDDING_DIM` in `.env` matches your table:
- Table has `vector(384)` → Set `EMBEDDING_DIM=384`
- Table has `vector(768)` → Set `EMBEDDING_DIM=768`

SCMD will automatically:
- **Truncate** if model returns more dimensions (e.g., 1536 → 384)
- **Pad** if model returns fewer dimensions (e.g., 256 → 384)

See [EMBEDDING_DIMENSIONS.md](EMBEDDING_DIMENSIONS.md) for details.

## Advanced Configuration

### Using Different Models

You can use any Ollama model that supports embeddings:

```env
# Small and fast
MODEL=llama2:7b

# Better quality
MODEL=mistral:latest

# Specialized
MODEL=codellama:latest
```

### Adjusting Search Results

Edit `ollama.go` to change the number of results:

```go
// Default: 5 results
results, err = SearchWithOllama(query, 5)

// More results: 10
results, err = SearchWithOllama(query, 10)
```

### Custom Prompts

Edit the system prompt in `ollama.go` to customize AI behavior:

```go
systemPrompt := `You are a helpful assistant that helps users find and understand command-line commands.
[Customize this to change AI behavior]`
```

## Benefits

### 1. Better Search Results
Find commands based on meaning, not just keywords.

### 2. Learning Tool
AI explanations help you understand commands better.

### 3. Natural Language
Ask questions naturally without worrying about exact keywords.

### 4. Semantic Understanding
Find related commands even with different terminology.

### 5. Zero Downtime
Automatic fallback ensures SCMD always works.

## Comparison

| Feature | Without Ollama | With Ollama |
|---------|---------------|-------------|
| Search Type | Keyword matching | Semantic similarity |
| AI Explanations | No | Yes |
| Natural Language | Limited | Full support |
| Speed | Very fast | Slightly slower |
| Accuracy | Good | Excellent |
| Fallback | N/A | Automatic |

## Best Practices

1. **Use natural language** - Take advantage of semantic search
2. **Check /ai status** - Verify Ollama is working
3. **Be patient** - Vector search takes a bit longer
4. **Trust the fallback** - If Ollama fails, traditional search still works
5. **Update embeddings** - Regenerate embeddings when adding many commands

## Future Enhancements

Planned features:
- Command suggestions based on context
- Multi-turn conversations
- Command composition assistance
- Automatic command categorization
- Learning from user interactions

## Summary

Ollama integration brings AI-powered search to SCMD:
- ✅ Vector similarity search
- ✅ AI-generated explanations
- ✅ Natural language understanding
- ✅ Automatic fallback
- ✅ Zero configuration (if Ollama is running)
- ✅ Backward compatible

**Try it now:**
```bash
scmd.exe -i
scmd> show me postgresql replication commands
```

Experience the future of command search! 🚀
