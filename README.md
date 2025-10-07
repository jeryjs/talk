# Nero - AI-Powered CLI Assistant

> A cutting-edge, agentic AI assistant with computer use capabilities, MCP support, and personality modes.

## 🌟 Features

### Core Capabilities
- **🤖 Agentic AI**: Heavy tool use with function calling
- **🎭 Personality System**: Dynamic mood-based interactions
- **📸 Computer Use**: Screen capture and vision capabilities
- **🗣️ Voice**: Text-to-speech synthesis
- **📝 File Operations**: Smart file editing and management
- **🔌 MCP Support**: 100% Model Context Protocol integration
- **🧩 Plugin System**: Extendable with hot-reload plugins
- **💬 Multi-Model**: Switch between OpenAI, Anthropic, Gemini, and more

### Advanced Features
- **Rich CLI Experience**: Tab completions, syntax highlighting
- **Context Prefixes**:
  - `/` - Commands (`/mood`, `/status`, `/help`, `/models`, `/chat`)
  - `#` - Resource access (`#terminal`, `#code`, `#screen`)
  - `@` - Mentions and references
- **Chat Management**: Save, load, and manage conversations
- **Web Search**: Built-in Google/Bing search capabilities

## 🚀 Quick Start

```bash
# Install
go install

# Run interactive mode
nero

# Direct command
nero "What's on my screen?"

# Set mood
nero /mood happy

# Use context
nero "help me #terminal debug this"
```

## 📦 Installation

### Prerequisites
- Go 1.22+
- Windows/Linux/macOS

### Build from Source
```bash
git clone https://github.com/yourusername/nero.git
cd nero
go mod download
go build -o nero cmd/nero/main.go
```

## 🎯 Usage Examples

### Basic Conversation
```bash
$ nero
Nero> Hello! How can I help you today?
You> What's the weather like?
Nero> Let me check that for you...
```

### Computer Use
```bash
$ nero "take a screenshot and describe what you see"
$ nero "click on the submit button"
$ nero "type 'hello world' in the active window"
```

### Resource Context
```bash
# Include terminal context
$ nero "help me #terminal fix this error"

# Include code context  
$ nero "review this #code for bugs"

# Include screen
$ nero "what do you see #screen"
```

### Mood System
```bash
$ nero /mood happy
✓ Mood set to: happy

$ nero /mood professional  
✓ Mood set to: professional

$ nero /mood sarcastic
✓ Mood set to: sarcastic
```

### Model Management
```bash
$ nero /models
Available models:
  - gpt-4-turbo
  - claude-3-opus
  - gemini-pro
  
$ nero /models use claude-3-opus
✓ Switched to: claude-3-opus
```

### Chat Management
```bash
$ nero /chat list
$ nero /chat save project-planning
$ nero /chat load project-planning
$ nero /chat delete project-planning
```

## ⚙️ Configuration

Create `config/config.yaml`:

```yaml
# API Keys
openai_api_key: "sk-..."
anthropic_api_key: "sk-ant-..."
google_api_key: "..."

# Default Settings
default_model: "gpt-4-turbo"
default_mood: "helpful"
max_tokens: 4096
temperature: 0.7

# Features
enable_voice: true
enable_computer_use: true
enable_mcp: true

# MCP Servers
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
  - name: "github"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "ghp_..."

# Plugins
plugins_dir: "./plugins"
auto_reload: true
```

## 🔌 Plugin Development

Create custom plugins in Go:

```go
package main

import (
    "github.com/yourusername/nero/pkg/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Init() error {
    // Initialize your plugin
    return nil
}

func (p *MyPlugin) Execute(ctx plugin.Context) error {
    // Plugin logic
    ctx.Reply("Hello from my plugin!")
    return nil
}

var Plugin MyPlugin
```

## 🏗️ Architecture

```
nero/
├── cmd/nero/           # Main application entry
├── internal/
│   ├── agent/          # AI agent logic
│   ├── chat/           # Chat management
│   ├── computer/       # Computer use capabilities
│   ├── config/         # Configuration
│   ├── mcp/            # MCP client implementation
│   ├── models/         # LLM providers
│   ├── mood/           # Personality system
│   ├── plugins/        # Plugin manager
│   ├── resources/      # Context collectors
│   ├── tools/          # Tool definitions
│   └── ui/             # TUI components
├── pkg/                # Public packages
└── plugins/            # Plugin binaries
```

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework  
- [RobotGo](https://github.com/go-vgo/robotgo) - Desktop automation
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Model Context Protocol
- And many more amazing open-source projects!

## 📞 Support

- 🐛 [Report Bug](https://github.com/yourusername/nero/issues)
- 💡 [Request Feature](https://github.com/yourusername/nero/issues)
- 💬 [Discussions](https://github.com/yourusername/nero/discussions)

---

Made with ❤️ by the Nero team
