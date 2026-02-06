# Gemini API Integration

## Overview
SCMD now supports Google Gemini API for generating embeddings, providing a reliable cloud-based alternative to Ollama.

## Why Gemini?
- ✅ Always available (cloud-based)
- ✅ No local installation required
- ✅ High-quality embeddings
- ✅ Fast and reliable
- ✅ Free tier available

## Configuration

### 1. Get Gemini API Key
1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Copy the key

### 2. Update .env File
```env
# Gemini API Configuration (Primary)
GEMINIAPI=your_api_key_here
GEMINIMODEL=text-embedding-004

# Embedding dimension (must match your PostgreSQL table)
EMBEDDING_DIM=384
```

### 3. Available Models
- `text-embedding-004` (Recommended) - Latest embedding model
- `embedding-001` - Previous generation

## Provider Priority

SCMD uses a fallback system:

1. **Gemini API** (Primary) - If `GEMINIAPI` is configured
2. **Ollama** (Fallback) - If Gemini fails or unavailable
3. **Text-only** (Last resort) - If no embedding provider available

## Usage

### Adding Commands
```bash
# Automatically generates embeddings using Gemini
scmd --save "docker ps -a" "List all containers"

# Output:
# ✓ Gemini API available (model: text-embedding-004, embedding_dim: 384)
# ✓ Generated embedding using Gemini API
# returned: ( true )
```

### Interactive Mode
```bash
scmd --cli

# Output:
# ✓ Gemini API available (model: text-embedding-004, embedding_dim: 384)
# ✓ Ollama available at localhost:11434 (model: llama2, embedding_dim: 384)
```

### Vector Search
When searching, SCMD will:
1. Try traditional keyword search first
2. If no good matches, use Gemini for vector similarity search
3. Return semantically similar commands

```bash
scmd --search "container management"

# Will find commands like:
# - docker ps -a
# - docker container ls
# - kubectl get pods
```

## Benefits Over Ollama

| Feature | Gemini | Ollama |
|---------|--------|--------|
| Availability | Always (cloud) | Requires local server |
| Setup | API key only | Install + model download |
| Speed | Fast | Depends on hardware |
| Quality | High | Varies by model |
| Cost | Free tier available | Free (local compute) |

## Embedding Dimensions

Gemini's `text-embedding-004` produces **768-dimensional** embeddings by default.

SCMD automatically adjusts to your configured dimension:

```env
# Your PostgreSQL table has vector(384)
EMBEDDING_DIM=384
```

SCMD will:
- Truncate 768 → 384 dimensions automatically
- No quality loss for command search use case

## Troubleshooting

### Error: "Gemini API returned status 403"
- Check your API key is correct
- Verify API key has embedding permissions
- Check if you've exceeded free tier limits

### Error: "Gemini API returned status 429"
- You've hit rate limits
- Wait a few seconds and retry
- Consider upgrading to paid tier

### Fallback to Ollama
If Gemini fails, SCMD automatically tries Ollama:
```
⚠ Gemini embedding failed: API error, trying Ollama...
✓ Generated embedding using Ollama
```

### No Embedding Provider
If both fail, commands are saved without embeddings:
```
⚠ No embedding provider available, saving without vector
```

Commands without embeddings still work with traditional text search!

## Cost Considerations

### Gemini Free Tier
- 1,500 requests per day
- More than enough for personal use

### Typical Usage
- Adding 1 command = 1 embedding request
- Searching = 1 embedding request per search
- ~100 commands/day = well within free tier

## Migration from Ollama

If you were using Ollama, no changes needed:

1. Add `GEMINIAPI` to `.env`
2. Gemini becomes primary
3. Ollama remains as fallback
4. Existing embeddings still work

## Best Practices

1. **Use Gemini for production** - More reliable
2. **Keep Ollama as fallback** - For offline scenarios
3. **Set EMBEDDING_DIM=384** - Good balance of speed/quality
4. **Monitor API usage** - Stay within free tier

## Example .env Configuration

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=scmd_user
DB_PASS=secure_password
DB_NAME=scmd_db
TB_NAME=scmd

# Gemini (Primary)
GEMINIAPI=AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
GEMINIMODEL=text-embedding-004

# Ollama (Fallback)
OLLAMA=localhost
MODEL=llama2

# Embeddings
EMBEDDING_DIM=384
```

## Summary

✅ Gemini API is now the default embedding provider
✅ More reliable than Ollama (always available)
✅ Easy setup (just API key)
✅ Automatic fallback to Ollama if needed
✅ Free tier sufficient for most users
✅ No code changes required - just update .env
