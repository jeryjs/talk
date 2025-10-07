# ✨ Nero - Complete Provider & Model Switching Guide

## 🎯 New Features

### 1. **Five Full Providers** ✅
- ✅ **OpenAI** - GPT models
- ✅ **Groq** - Ultra-fast Llama models
- ✅ **Gemini** - Google's latest with thinking
- ✅ **OpenRouter** - 500+ models gateway
- ✅ **Ollama** - Local models (fallback)

### 2. **Switch Models Mid-Chat** 🔄
Use `/model` command to change models without restarting!

---

## 🚀 Provider/Model Format

### Shortcuts
```
o/   = Ollama
g/   = Google Gemini
c/   = Anthropic Claude
oa/  = OpenAI
gq/  = Groq
or/  = OpenRouter
```

### Examples
```bash
# Config file
default_model: "gq/llama-3.3-70b"      # Groq (fastest)
default_model: "g/gemini-2.5-pro"      # Gemini (smartest)
default_model: "oa/gpt-4o"             # OpenAI (balanced)
default_model: "or/claude-sonnet-4.5"  # OpenRouter gateway
default_model: "o/llama3.2"            # Ollama (local, free)
```

---

## 🔄 Switch Models Mid-Chat

### Usage
```
You: Hi there

Nero: Hello! How can I help you?

You: /model gq/llama-3.3-70b

✓ Switched to model: gq/llama-3.3-70b
⚠️  Conversation history cleared

You: /model g/gemini-2.5-pro

✓ Switched to model: g/gemini-2.5-pro
⚠️  Conversation history cleared
```

### Why Clear History?
Different models have different:
- Context formats
- Tool calling schemas
- Token limits
- Capabilities

So history is cleared when switching for safety.

---

## 🤖 Available Models

### OpenAI (o/)
```
o/gpt-4.5              - Latest Orion (200K context)
o/gpt-4o               - Multimodal (128K context)
o/gpt-4-turbo          - Fast & capable (128K)
```

### Groq (gq/) - ULTRA FAST ⚡
```
gq/llama-3.3-70b-versatile    - Best balance
gq/llama-3.1-70b-versatile    - Stable version
gq/llama-3.1-8b-instant       - Fastest
gq/mixtral-8x7b-32768         - Good for code
```

### Google Gemini (g/)
```
g/gemini-2.5-pro              - Latest with thinking (2M context!)
g/gemini-2.0-flash-exp        - Experimental fast
g/gemini-2.0-flash-thinking-exp - Reasoning focused
g/gemini-1.5-pro              - Stable version
```

### OpenRouter (or/) - GATEWAY TO 500+ MODELS
```
or/anthropic/claude-sonnet-4.5      - Claude via OpenRouter
or/google/gemini-2.5-pro            - Gemini via OpenRouter
or/openai/gpt-4.5                   - GPT via OpenRouter
or/meta-llama/llama-3.3-70b         - Meta Llama
or/deepseek/deepseek-r1             - DeepSeek reasoning
```

### Ollama (ol/) - LOCAL & FREE
```
ol/llama3.2            - Default local model
ol/qwen2.5             - Multilingual
ol/deepseek-r1         - Reasoning
```

---

## 📚 Chat Commands

```
/help                  - Show all commands
/mood [name]           - Change personality
/model [provider/model] - Switch AI model ✨ NEW!
/clear                 - Clear conversation
/quit                  - Exit
```

---

## 🎭 Example Session

```bash
$ nero

⚙️  Using model: llama3.2 from provider: ollama

🤖 Nero - AI Assistant
Mood: helpful | Provider: O | Model: llama3.2
(Shortcuts: o/gpt-4o, gq/llama-3.3-70b, ol/llama3.2)

Type /help for commands, Ctrl+C to quit
────────────────────────────────────────────────────────────

You: Hi! What model are you?

Nero: I'm running on Llama 3.2 via Ollama, a local model.

You: /model gq/llama-3.3-70b

⚙️  Using model: llama-3.3-70b from provider: groq
✓ Switched to model: gq/llama-3.3-70b
⚠️  Conversation history cleared

You: Now what model?

Nero: I'm now running on Llama 3.3 70B via Groq's ultra-fast inference!

You: /model g/gemini-2.5-pro

⚙️  Using model: gemini-2.5-pro from provider: gemini
✓ Switched to model: g/gemini-2.5-pro
⚠️  Conversation history cleared

You: And now?

Nero: I'm Google Gemini 2.5 Pro with thinking capabilities and 2M context window!

You: /quit
👋 Goodbye!
```

---

## ⚙️ Configuration

### Config File: `~/.nero/config.yaml`

```yaml
# API Keys (all optional - Ollama works without!)
openai_api_key: "sk-..."
groq_api_key: "gsk_..."          # Get from console.groq.com
google_api_key: "..."            # Get from ai.google.dev
openrouter_api_key: "sk-or-..."  # Get from openrouter.ai

# Default model
default_model: "llama3.2"  # Local Ollama (no API key!)

# Try these with API keys:
# default_model: "gq/llama-3.3-70b"     # Groq (fastest!)
# default_model: "g/gemini-2.5-pro"     # Gemini (smartest!)
# default_model: "o/gpt-4o"             # OpenAI (balanced)
# default_model: "or/claude-sonnet-4.5" # OpenRouter gateway

default_mood: "helpful"
```

---

## 🔑 Get API Keys

### Groq (Recommended - Ultra Fast & Free Tier!)
1. Visit: https://console.groq.com
2. Sign up (free)
3. Get API key: `gsk_...`
4. Add to config: `groq_api_key: "gsk_..."`

### Google Gemini
1. Visit: https://ai.google.dev
2. Get API key
3. Add to config: `google_api_key: "..."`

### OpenRouter (Access 500+ Models!)
1. Visit: https://openrouter.ai
2. Sign up
3. Get API key: `sk-or-...`
4. Add to config: `openrouter_api_key: "sk-or-..."`

### OpenAI
1. Visit: https://platform.openai.com
2. Get API key: `sk-...`
3. Add to config: `openai_api_key: "sk-..."`

---

## 🎯 Provider Comparison

| Provider | Speed | Cost | Setup | Models | Best For |
|----------|-------|------|-------|--------|----------|
| **Ollama** | ⚡⚡ | Free | Easy | Any local | Privacy, offline, free |
| **Groq** | ⚡⚡⚡ | $ | Easy | Llama, Mixtral | Speed, development |
| **Gemini** | ⚡⚡ | $ | Easy | Gemini family | Long context, thinking |
| **OpenRouter** | ⚡ | $ | Easy | 500+ models | Model variety |
| **OpenAI** | ⚡ | $$$ | Easy | GPT family | Quality, features |

---

## 💡 Pro Tips

### 1. **Start Local, Scale Up**
- Use Ollama by default (free, private)
- Switch to Groq for speed when needed
- Use Gemini for long documents
- Use OpenAI for production

### 2. **Compare Models**
```
You: How do I center a div?

Nero: [Ollama response]

You: /model gq/llama-3.3-70b

Nero: [Groq response - same question, different style]

You: /model g/gemini-2.5-pro

Nero: [Gemini response - with thinking]
```

### 3. **Use OpenRouter for Variety**
- Access 500+ models through one API
- Try DeepSeek, Qwen, Mistral, etc.
- Compare different providers

### 4. **Groq for Speed**
- 10-100x faster than OpenAI
- Great for development/testing
- Free tier available

---

## 🐛 Troubleshooting

### Model Switch Failed?
```
❌ Failed to switch model: provider not available

# Solution: Check API key in config
nano ~/.nero/config.yaml
```

### Always Falling Back to Ollama?
```
⚠️  Provider 'gemini' not available, falling back to Ollama

# Solution: Make sure Ollama is installed
# Download from: https://ollama.ai
```

### Model Not Found?
```
❌ Error: model does not exist

# Solution: Check model name
# Use /help to see available models
# Or check provider docs
```

---

## 🚀 Quick Start

```bash
# 1. Install Ollama (optional but recommended)
# Download from https://ollama.ai

# 2. Pull a model
ollama pull llama3.2

# 3. Start Nero
nero

# 4. Try switching models
You: /model ol/llama3.2     # Local
You: /model gq/llama-3.3-70b # Fast (needs API key)
You: /model g/gemini-2.5-pro # Smart (needs API key)
```

---

**All 5 providers working! Switch models anytime!** 🎉

**Recommendation**: Start with Ollama (free), add Groq API key (fast & free tier), then explore others!
