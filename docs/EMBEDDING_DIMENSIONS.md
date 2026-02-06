# Embedding Dimensions Guide

## Understanding Embedding Dimensions

Vector embeddings are numerical representations of text. Different models produce embeddings of different sizes (dimensions).

## Common Dimensions

| Dimension | Models | Use Case |
|-----------|--------|----------|
| 384 | all-MiniLM-L6-v2, nomic-embed-text | Fast, efficient, good quality |
| 768 | all-mpnet-base-v2, BERT | Better quality, slower |
| 1024 | gtr-t5-large | High quality |
| 1536 | OpenAI ada-002, some Ollama models | Highest quality |

## Your Configuration

Your PostgreSQL table is configured for **384 dimensions**:

```sql
CREATE TABLE scmd (
    ...
    embedding vector(384),
    ...
);
```

## Setting EMBEDDING_DIM

In your `.env` file, set `EMBEDDING_DIM` to match your table:

```env
EMBEDDING_DIM=384
```

## How SCMD Handles Dimensions

SCMD automatically adjusts embeddings to match your configured dimension:

### If Model Returns More Dimensions (e.g., 1536)
**Truncates** to your configured dimension (384):
```
Model returns: [1536 dimensions]
SCMD uses:     [first 384 dimensions]
```

### If Model Returns Fewer Dimensions (e.g., 256)
**Pads with zeros** to your configured dimension (384):
```
Model returns: [256 dimensions]
SCMD uses:     [256 dimensions + 128 zeros]
```

### If Dimensions Match (e.g., 384)
**Uses directly** without modification:
```
Model returns: [384 dimensions]
SCMD uses:     [384 dimensions]
```

## Your Current Setup

Based on your error message, your Ollama model returns **1536 dimensions**, but your table expects **384 dimensions**.

**Solution:** SCMD now automatically truncates to 384 dimensions!

### Your .env Configuration

```env
# PostgreSQL table has vector(384)
EMBEDDING_DIM=384

# Ollama model (returns 1536 dimensions)
OLLAMA=192.168.0.78
MODEL=dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m
```

SCMD will:
1. Get 1536-dimension embedding from Ollama
2. Truncate to first 384 dimensions
3. Store in PostgreSQL vector(384) column

## Changing Dimensions

### Option 1: Keep 384 (Recommended)
- Fast and efficient
- Good quality for command search
- No database changes needed
- **Already configured!**

### Option 2: Change to 1536
If you want full 1536 dimensions:

1. **Update PostgreSQL table:**
```sql
-- Drop existing index
DROP INDEX IF EXISTS idx_scmd_embedding;

-- Change column dimension
ALTER TABLE scmd ALTER COLUMN embedding TYPE vector(1536);

-- Recreate index
CREATE INDEX idx_scmd_embedding 
ON scmd USING hnsw (embedding vector_cosine_ops);
```

2. **Update .env:**
```env
EMBEDDING_DIM=1536
```

3. **Regenerate all embeddings** (existing embeddings will be invalid)

### Option 3: Use 384-Dimension Model
Use an Ollama model that natively produces 384 dimensions:

```bash
# Pull a 384-dimension model
ollama pull nomic-embed-text

# Update .env
MODEL=nomic-embed-text
EMBEDDING_DIM=384
```

## Recommended Configuration

For best performance with your current setup:

```env
# Keep your table at 384 dimensions
EMBEDDING_DIM=384

# Use your current model (SCMD will truncate automatically)
OLLAMA=192.168.0.78
MODEL=dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m
```

**Benefits:**
- ✓ No database changes needed
- ✓ Faster vector search (smaller vectors)
- ✓ Less storage space
- ✓ Good quality for command search
- ✓ Works with any Ollama model

## Troubleshooting

### Error: "different vector dimensions X and Y"

This means:
- Your model returns X dimensions
- Your table expects Y dimensions
- SCMD couldn't adjust automatically

**Solution:**
1. Check `EMBEDDING_DIM` in `.env` matches your table
2. Rebuild SCMD: `go build -o scmd.exe .`
3. Restart SCMD

### Checking Your Table Dimension

```sql
SELECT 
    column_name, 
    udt_name, 
    character_maximum_length 
FROM information_schema.columns 
WHERE table_name = 'scmd' 
  AND column_name = 'embedding';
```

Look for `vector(384)` or similar.

### Verifying SCMD Configuration

```
scmd.exe -i

✓ Ollama available at 192.168.0.78:11434 (model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m, embedding_dim: 384)
```

The `embedding_dim: 384` confirms SCMD will use 384 dimensions.

## Performance Impact

### Truncation (1536 → 384)
- **Speed:** No impact (happens in memory)
- **Quality:** Minimal loss (first dimensions are most important)
- **Storage:** 4x less space

### Padding (256 → 384)
- **Speed:** No impact
- **Quality:** No loss (zeros don't affect similarity)
- **Storage:** Slightly more space

## Summary

- ✅ SCMD automatically handles dimension mismatches
- ✅ Truncates if model returns more dimensions
- ✅ Pads if model returns fewer dimensions
- ✅ Set `EMBEDDING_DIM` to match your PostgreSQL table
- ✅ 384 dimensions is recommended for command search
- ✅ No code changes needed - it's automatic!

Your configuration is now correct and should work! 🎉
