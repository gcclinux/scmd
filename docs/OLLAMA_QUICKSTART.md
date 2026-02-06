# Ollama Quick Start

## 5-Minute Setup

### Step 1: Install Ollama

Visit [ollama.ai](https://ollama.ai) and install Ollama for your platform.

### Step 2: Pull a Model

```bash
ollama pull llama2
# or
ollama pull dagbs/dolphin-2.9.3-qwen2-1.5b:q4_k_m
```

### Step 3: Start Ollama

```bash
ollama serve
```

### Step 4: Configure SCMD

Edit your `.env` file:

```env
OLLAMA=localhost
MODEL=llama2
EMBEDDING_DIM=384
```

### Step 5: Start SCMD

```bash
scmd.exe -i
```

You should see:
```
✓ Ollama available at localhost:11434 (model: llama2)
```

### Step 6: Try It!

```
scmd> show me postgresql replication commands

🤖 AI Assistant:
══════════════════════════════════════════════════════════════
[AI-generated explanation appears here]
══════════════════════════════════════════════════════════════

Found 4 result(s):
[Results...]
```

## That's It!

You now have AI-powered command search! 🎉

## Troubleshooting

### Ollama Not Detected?

```bash
# Test Ollama manually
curl http://localhost:11434/api/tags

# Should return JSON with available models
```

### Model Not Found?

```bash
# List available models
ollama list

# Pull the model if missing
ollama pull llama2
```

### Slow Performance?

Try a smaller model:
```bash
ollama pull llama2:7b
```

Update `.env`:
```env
MODEL=llama2:7b
```

## Next Steps

- Read [OLLAMA_INTEGRATION.md](OLLAMA_INTEGRATION.md) for full documentation
- Try natural language queries
- Check `/ai` command for status
- Experiment with different models

## Without Ollama

SCMD works perfectly fine without Ollama! It will automatically use traditional keyword search if Ollama is not available.
