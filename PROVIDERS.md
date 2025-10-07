# 🚀 Nero - Model Provider System

## Provider/Model Format

Nero now uses a clean `provider/model` format with shortcuts!

### Shortcuts
- `o/` = OpenAI
- `g/` = Google Gemini  
- `c/` = Anthropic Claude
- `gq/` = Groq (ultra-fast!)
- `ol/` = Ollama (local)

### Examples
```yaml
# config.yaml
default_model: "gq/llama-3.3-70b"  # Groq with Llama 3.3
default_model: "o/gpt-4o"           # OpenAI GPT-4o
default_model: "ol/llama3.2"        # Local Ollama
default_model: "g/gemini-2.5-pro"   # Google Gemini
default_model: "c/claude-sonnet-4.5" # Claude

# Or use full provider names
default_model: "groq/llama-3.3-70b"
default_model: "openai/gpt-4o"
```

### Auto-Detection
If you don't use a prefix, Nero auto-detects:
- `gpt-*` → OpenAI
- `claude-*` → Anthropic
- `gemini-*` → Google
- `llama*`, `qwen*`, `deepseek*` → Ollama
- Everything else → Ollama (default)

### Fallback Strategy
**Always falls back to Ollama (local, no API key needed)**

1. Try requested provider
2. If fails → Use Ollama with warning
3. If Ollama also fails → Error

This means Nero **always works** even without API keys!

---

## Supported Providers

### ✅ **Groq** (NEW!)
**Ultra-fast inference on custom hardware**
- Models: `llama-3.3-70b-versatile`, `llama-3.1-70b-versatile`, `llama-3.1-8b-instant`
- Speed: Sub-second responses
- API Key: `groq_api_key`
- Shortcut: `gq/`

### ✅ **OpenAI**
- Models: `gpt-4.5`, `gpt-4o`, `gpt-4-turbo`, `gpt-4`
- Features: Text, vision, audio, function calling
- API Key: `openai_api_key`
- Shortcut: `o/`

### ✅ **Ollama** (LOCAL - NO API KEY!)
- Models: Any model you have pulled
- Popular: `llama3.2`, `qwen2.5`, `deepseek-r1`
- No API key needed
- Shortcut: `ol/`

### 🚧 **Google Gemini** (Prepared)
- Models: `gemini-2.5-pro`, `gemini-2.0-flash-exp`
- Needs SDK implementation
- API Key: `google_api_key`
- Shortcut: `g/`

### 🚧 **Anthropic Claude** (Prepared)
- Models: `claude-sonnet-4.5`, `claude-3.5-sonnet`
- Needs SDK fixes
- API Key: `anthropic_api_key`
- Shortcut: `c/`

---

## Configuration

### Config File: `~/.nero/config.yaml`

```yaml
# API Keys (all optional!)
groq_api_key: "gsk_..."      # Groq - ultra-fast!
openai_api_key: "sk-..."     # OpenAI
google_api_key: "..."        # Google Gemini
anthropic_api_key: "sk-ant..." # Claude

# Default model (uses Ollama by default - no API key!)
default_model: "llama3.2"

# Examples with shortcuts:
# default_model: "gq/llama-3.3-70b"   # Groq (fastest!)
# default_model: "o/gpt-4o"           # OpenAI
# default_model: "ol/llama3.2"        # Ollama local
```

### Environment Variables
You can also use environment variables:
```bash
export GROQ_API_KEY="gsk_..."
export OPENAI_API_KEY="sk-..."
```

---

## Usage

### Interactive Mode
```bash
$ nero

🤖 Nero - AI Assistant
Model: llama3.2 | Mood: helpful
(Use provider/model format: o/gpt-4o, gq/llama-3.3-70b, ol/llama3.2)

You: hi
Nero: Hello! How can I help you today?
```

### Direct Questions
```bash
# Using default model (Ollama)
nero "What's 2+2?"

# With specific model flag
nero -m "gq/llama-3.3-70b" "Explain quantum computing"
```

### Change Model in Config
```bash
# Edit config
nano ~/.nero/config.yaml

# Change this line:
default_model: "gq/llama-3.3-70b"  # Ultra-fast Groq!
```

---

## Why This Design?

### 1. **Always Works**
- Ollama fallback means no API keys required
- Local models always available
- No "401 Unauthorized" errors

### 2. **Clear and Explicit**
- `o/gpt-4o` is clearer than guessing from "gpt-4o"
- Know exactly which provider you're using
- Easy to switch: just change prefix

### 3. **Fast Options**
- Groq provides ultra-fast inference
- Great for development and testing
- Much cheaper than OpenAI

### 4. **Flexible**
- Use shortcuts (`gq/`) or full names (`groq/`)
- Auto-detection for common models
- Easy to add new providers

---

## Provider Comparison

| Provider | Speed | Cost | API Key? | Best For |
|----------|-------|------|----------|----------|
| **Groq** | ⚡⚡⚡ | 💰 | Yes | Fast responses, development |
| **Ollama** | ⚡⚡ | Free | No | Local, privacy, no costs |
| **OpenAI** | ⚡ | 💰💰💰 | Yes | Best quality, latest features |
| **Gemini** | ⚡⚡ | 💰 | Yes | Long context, multimodal |
| **Claude** | ⚡ | 💰💰 | Yes | Writing, analysis, coding |

---

## Next Steps

1. **Try Groq** - Get API key from https://console.groq.com
2. **Install Ollama** - Download from https://ollama.ai
3. **Add API Keys** - Edit `~/.nero/config.yaml`
4. **Test Models** - Try different providers with shortcuts

---

**Built with ❤️ for flexibility and speed!** 🚀
