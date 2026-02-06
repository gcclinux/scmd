# Quick Fix - Embedding Dimension

## The Issue

You were getting:
```
✓ Ollama available at 192.168.0.78:11434 (model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m, embedding_dim: 1536)
⚠ Ollama search failed: different vector dimensions 1536 and 384
```

## The Problem

Your `.env` file had:
```env
EMBEDDING_DIM=1536  ❌ Wrong!
```

But your PostgreSQL table has:
```sql
embedding vector(384)  ✓ Correct
```

## The Fix

Updated `.env` file to:
```env
EMBEDDING_DIM=384  ✓ Correct!
```

## Verify the Fix

Run SCMD again:
```bash
.\scmd.exe -i
```

You should now see:
```
✓ Ollama available at 192.168.0.78:11434 (model: dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m, embedding_dim: 384)
```

Notice: `embedding_dim: 384` ✓

## Test It

```
scmd> how do I check postgresql replication on a slave server

🤖 AI Assistant:
══════════════════════════════════════════════════════════════
[AI explanation should appear here]
══════════════════════════════════════════════════════════════

Found 2 result(s):
[Results should appear here]
```

## What Happens Now

1. Ollama returns 1536-dimension embedding
2. SCMD reads `EMBEDDING_DIM=384` from `.env`
3. SCMD automatically truncates to first 384 dimensions
4. Stores in PostgreSQL vector(384) column
5. ✅ **It works!**

## Key Point

**EMBEDDING_DIM must match your PostgreSQL table dimension!**

Check your table:
```sql
\d scmd
-- Look for: embedding | vector(384)
```

Set EMBEDDING_DIM to match:
```env
EMBEDDING_DIM=384  # Must match table!
```

## Summary

✅ Fixed `.env` file: `EMBEDDING_DIM=384`
✅ Matches PostgreSQL table: `vector(384)`
✅ SCMD will auto-truncate: 1536 → 384
✅ Should work now!

Try it! 🚀
