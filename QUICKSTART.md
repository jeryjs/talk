# 🤖 Nero - Quick Reference

## Chat Commands (Use these INSIDE Nero)

```
/help              Show all commands
/mood [name]       Change personality
/clear             Clear conversation
/quit              Exit
```

## Usage

```bash
# Start interactive chat
nero

# Quick question
nero "your question here"

# Show status
nero status
```

## Available Moods
- helpful (default) - Friendly assistant
- creative - Imaginative responses
- professional - Formal tone
- analytical - Logical & detailed
- concise - Brief answers
- sarcastic - Witty & dry
- empathetic - Understanding
- enthusiastic - Energetic

## Models (2025 Latest)

**Default**: Gemini 2.5 Pro ⭐

- OpenAI: gpt-4.5, gpt-4o, gpt-4-turbo
- Google: gemini-2.5-pro, gemini-2.0-flash-exp
- Ollama: llama3.1:405b, qwen2.5:235b, deepseek-r1

## Configuration

Edit `~/.nero/config.yaml`:
```yaml
google_api_key: "YOUR_KEY"
default_model: "gemini-2.5-pro"
default_mood: "helpful"
```

## Build

```bash
.\build.ps1          # Windows
./build.sh           # Linux/Mac
```

See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for full documentation.
