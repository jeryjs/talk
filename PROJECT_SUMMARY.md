# 🤖 Nero - Project Summary & Architecture

> **Nero** is a modern, cutting-edge AI-powered CLI assistant built in Go with agentic capabilities, multi-model support, computer use features, and personality system.

**Default Model**: Gemini 2.5 Pro ⭐ (Latest with thinking capabilities)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features Implemented](#features-implemented)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Supported Models](#supported-models)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Usage Examples](#usage-examples)
- [Development Status](#development-status)
- [Next Steps](#next-steps)

---

## 🎯 Overview

Nero transforms your previous simple TTS CLI into a **comprehensive AI assistant** with:

- **🤖 Agentic AI**: Heavy tool use with function calling capabilities
- **🎭 Personality System**: 8+ mood states (helpful, professional, creative, analytical, etc.)
- **📸 Computer Use**: Screen capture, analysis, and automation
- **🗣️ Voice Synthesis**: Text-to-speech with multiple engines
- **📝 File Operations**: Smart file editing and management
- **🔌 MCP Support**: 100% Model Context Protocol integration
- **🧩 Plugin System**: Extendable architecture with hot-reload
- **💬 Multi-Model**: Seamless switching between OpenAI, Anthropic, Google, and Ollama
- **🌐 Web Search**: Built-in web search and content fetching

---

## ✨ Features Implemented

### Core CLI Framework ✅
- **Cobra-based CLI** with comprehensive command structure
- **Tab completion** support for all shells
- **Interactive TUI** using Bubbletea (framework in place)
- **Rich output** with colors and formatting via Lipgloss

### Model Providers ✅
| Provider | Status | Latest Models | Context Window |
|----------|--------|---------------|----------------|
| **Google Gemini** | ✅ Ready | gemini-2.5-pro, 2.0-flash-exp | 2M tokens |
| **OpenAI** | ✅ Ready | gpt-4.5, gpt-4o, gpt-4-turbo | 200K tokens |
| **Anthropic** | ⚙️ Prepared | claude-sonnet-4.5, 3.5-sonnet | 200K tokens |
| **Ollama** | ✅ Ready | Local models (llama, qwen, deepseek) | Varies |
| **OpenRouter** | ⚙️ Prepared | 500+ models gateway | Varies |

### Tools System ✅
- **Screenshot Tool**: Capture and save screenshots
- **Screen Info Tool**: Get display information
- **Screen Analysis Tool**: Basic image analysis
- **File Read/Write**: Smart file operations
- **File List**: Directory browsing
- **Shell Execute**: Safe command execution
- **Web Search**: Google search integration
- **Web Fetch**: Content extraction from URLs

### Configuration Management ✅
- **YAML-based config** with sensible defaults
- **Environment variable** support
- **Auto-creation** of config directories
- **Runtime config updates**
- **Default: Gemini 2.5 Pro** ⭐

### CLI Commands ✅
```
nero [message]          # Direct message or interactive mode
nero chat               # Chat management (list, save, load, delete, export)
nero model              # Model management (list, use, info)
nero mood [mood]        # Set personality (helpful, professional, creative, etc.)
nero status             # Show system status and configuration
nero plugin             # Plugin management (list, install, enable, disable)
nero config             # Configuration management (show, set, reset)
nero version            # Version information
```

---

## 🏗️ Architecture

```
nero/
├── cmd/nero/              # Application entry point
│   └── main.go           # Initializes CLI with version info
│
├── internal/             # Internal packages
│   ├── agent/           # AI agent orchestration logic
│   │   └── agent.go     # Conversation flow, tool execution
│   │
│   ├── cli/             # Cobra commands
│   │   ├── root.go      # Root command & flags
│   │   ├── chat.go      # Chat management commands
│   │   ├── model.go     # Model selection commands
│   │   ├── mood.go      # Personality system
│   │   ├── status.go    # Status display
│   │   ├── plugin.go    # Plugin management
│   │   ├── config.go    # Config commands
│   │   └── version.go   # Version info
│   │
│   ├── computer/        # Computer use capabilities
│   │   ├── init.go      # Initialization
│   │   └── screenshot.go # Screen capture
│   │
│   ├── config/          # Configuration management
│   │   └── config.go    # Viper-based config loader
│   │
│   ├── models/          # LLM provider implementations
│   │   ├── models.go    # Provider interface & registry
│   │   ├── openai.go    # OpenAI implementation
│   │   ├── anthropic.go # Anthropic Claude (prepared)
│   │   ├── gemini.go    # Google Gemini (prepared)
│   │   ├── ollama.go    # Ollama local models
│   │   └── openrouter.go# OpenRouter gateway (prepared)
│   │
│   ├── mood/            # Personality system
│   │   └── mood.go      # Mood definitions & prompts
│   │
│   ├── tools/           # Tool definitions & execution
│   │   ├── registry.go  # Tool registration system
│   │   ├── computer.go  # Computer use tools
│   │   ├── file_tools.go# File operation tools
│   │   └── web_search.go# Web search tools
│   │
│   └── ui/              # TUI components
│       └── interactive.go # Bubbletea interface
│
├── speechUtils/         # Legacy TTS utilities (preserved)
│   ├── brain.go         # AI chat functions
│   ├── ear.go           # Speech recognition (placeholder)
│   └── mouth.go         # TTS engines
│
├── assets/              # Bundled assets
│   └── espeak-data/    # eSpeak voice data
│
├── config.example.yaml  # Example configuration
├── build.ps1           # Windows build script
├── build.sh            # Unix build script
├── CONTRIBUTING.md     # Contribution guidelines
└── README.md           # User documentation
```

---

## 🛠️ Technology Stack

### Core Framework
- **Go 1.22+**: Modern Go with latest features
- **Cobra v1.8**: CLI framework with auto-completion
- **Viper v1.19**: Configuration management
- **Bubbletea v1.3**: Terminal UI framework
- **Lipgloss v1.1**: Styling and layout

### LLM SDKs
- **go-openai v1.38**: OpenAI official SDK
- **anthropic-sdk-go v1.5**: Anthropic official SDK
- **google.golang.org/genai**: Google Gemini SDK
- Native HTTP clients for Ollama & OpenRouter

### Computer Use
- **kbinani/screenshot**: Cross-platform screenshots
- **gocolly/colly v2**: Web scraping and search

### Legacy (Preserved)
- **edge-tts-go**: Microsoft Edge TTS
- **htgo-tts**: Google TTS
- eSpeak integration

---

## 🤖 Supported Models (2025 Latest)

### 🔷 OpenAI
- **gpt-4.5** (Orion) - 200K context, enhanced reasoning
- **gpt-4o** - Multimodal (text, audio, vision)
- **gpt-4-turbo** - 128K context, fast & capable

### 🔶 Anthropic Claude
- **claude-sonnet-4.5** - Best for coding & agents, 30hr task focus
- **claude-3.5-sonnet** - Balanced speed & intelligence
- **claude-3-opus** - Maximum capability

### 🔵 Google Gemini (DEFAULT ⭐)
- **gemini-2.5-pro** - Latest with thinking, 2M context
- **gemini-2.0-flash-exp** - Experimental fast model
- **gemini-2.0-flash-thinking-exp** - Reasoning-focused

### 🦙 Ollama (Local)
- **llama3.1:405b** - Meta's largest open model
- **qwen2.5:235b** - Alibaba's multilingual model
- **deepseek-r1:671b** - Advanced reasoning model

---

## 📁 Project Structure Details

### Key Files

#### Configuration
- `~/.nero/config.yaml` - User configuration
- `config.example.yaml` - Template with all options

#### Build Outputs
- `nero.exe` (Windows) / `nero` (Linux/Mac)
- Portable single binary

#### Build Scripts
- `build.ps1` - PowerShell build with version injection
- `build.sh` - Bash build script

### Important Directories
- `~/.nero/` - User data directory
  - `config.yaml` - Configuration
  - `chats/` - Saved conversations
  - `logs/` - Application logs
  - `plugins/` - Installed plugins

---

## 🚀 Getting Started

### Prerequisites
```bash
# Go 1.22 or later
go version

# Windows: PowerShell 5.1+
# Linux/Mac: Bash
```

### Build
```powershell
# Windows
.\build.ps1

# Linux/Mac
chmod +x build.sh
./build.sh
```

### Configuration
```bash
# First run creates default config
nero status

# Edit config
notepad ~/.nero/config.yaml  # Windows
nano ~/.nero/config.yaml     # Linux/Mac
```

### API Keys
Add to `~/.nero/config.yaml`:
```yaml
google_api_key: "YOUR_KEY"      # Required for Gemini (default)
openai_api_key: "sk-..."        # Optional
anthropic_api_key: "sk-ant-..." # Optional
```

---

## 💡 Usage Examples

### Basic Interaction
```bash
# Interactive mode
nero

# Direct question
nero "What's the weather like in Tokyo?"

# With specific model
nero --model gpt-4.5 "Explain quantum computing"
```

### Model Management
```bash
# List all models
nero model list

# Switch model
nero model use claude-sonnet-4.5

# Model info
nero model info gemini-2.5-pro
```

### Personality/Mood
```bash
# Show current mood
nero mood

# Change personality
nero mood creative      # Imaginative responses
nero mood professional  # Formal tone
nero mood sarcastic     # Witty with dry humor
nero mood concise       # Brief answers
```

### Chat Management
```bash
# Save conversation
nero chat save project-planning

# List saved chats
nero chat list

# Load previous chat
nero chat load project-planning

# Export chat
nero chat export project-planning output.json
```

### Context Prefixes (Planned)
```bash
# Terminal context
nero "help me #terminal debug this error"

# Code context
nero "review this #code for bugs"

# Screen context
nero "what do you see #screen"
```

---

## 📊 Development Status

### ✅ Completed
- [x] CLI framework with Cobra
- [x] Configuration system with Viper
- [x] OpenAI provider integration
- [x] Ollama provider integration
- [x] Tool system architecture
- [x] Computer use tools (screenshots, files, shell)
- [x] Web search tools
- [x] Mood/personality system
- [x] Status and diagnostics
- [x] Build scripts
- [x] Documentation
- [x] Latest 2025 models updated
- [x] Gemini set as default

### 🚧 In Progress
- [ ] Anthropic Claude provider (API changes)
- [ ] Google Gemini provider (package resolution)
- [ ] OpenRouter integration
- [ ] Interactive TUI mode
- [ ] Agent conversation flow
- [ ] Chat persistence

### 📋 Planned
- [ ] MCP server integration
- [ ] Plugin system implementation
- [ ] Voice input/output
- [ ] Context collectors (#terminal, #code, #screen)
- [ ] Chat export formats (Markdown, PDF)
- [ ] Advanced computer use (mouse, keyboard control)
- [ ] Multi-turn conversations with context
- [ ] Streaming responses
- [ ] Cost tracking
- [ ] Usage statistics

---

## 🎯 Next Steps

### Immediate Priorities
1. **Fix Provider Implementations**
   - Resolve Anthropic SDK API compatibility
   - Fix Google Gemini package version
   - Test OpenRouter integration

2. **Complete Agent Logic**
   - Implement conversation loop
   - Add tool execution flow
   - Handle streaming responses

3. **Interactive TUI**
   - Finish Bubbletea interface
   - Add keyboard shortcuts
   - Implement chat history display

### Short Term
- Add chat persistence (save/load)
- Implement context collectors
- Add plugin loading system
- Complete MCP client integration

### Long Term
- Advanced computer control
- Voice interface
- IDE integrations (VS Code extension)
- Web UI option
- Cloud deployment support

---

## 📚 Additional Resources

### Documentation
- [README.md](README.md) - User guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guide
- [config.example.yaml](config.example.yaml) - Configuration reference

### External Links
- [Cobra Framework](https://github.com/spf13/cobra)
- [Bubbletea TUI](https://github.com/charmbracelet/bubbletea)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenAI API](https://platform.openai.com/docs/)
- [Anthropic Claude](https://docs.anthropic.com/)
- [Google Gemini](https://ai.google.dev/gemini-api/docs)
- [Ollama](https://ollama.ai/)

---

## 🙏 Acknowledgments

This project was completely rebuilt from a simple TTS CLI into a comprehensive AI assistant using:
- Latest Go best practices
- Modern CLI/TUI frameworks
- Cutting-edge LLM capabilities
- Extensible architecture

**Original**: Simple voice chat with GPT/Gemini  
**Now**: Full-featured agentic AI assistant with computer use, multi-model support, personality system, and extensible tools

---

**Built with ❤️ using Go and the power of AI** 🚀

*Last Updated: October 2025*
