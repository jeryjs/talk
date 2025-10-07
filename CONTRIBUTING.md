# Contributing to Nero

Thank you for your interest in contributing to Nero! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/yourusername/nero.git
   cd nero
   ```
3. **Install dependencies**
   ```bash
   go mod download
   ```
4. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development

### Project Structure

```
nero/
├── cmd/nero/           # Main application entry point
├── internal/           # Internal packages
│   ├── agent/          # AI agent logic
│   ├── cli/            # CLI commands
│   ├── computer/       # Computer use capabilities
│   ├── config/         # Configuration management
│   ├── models/         # LLM provider implementations
│   ├── mood/           # Personality system
│   ├── tools/          # Tool implementations
│   └── ui/             # User interface components
└── pkg/                # Public packages (if any)
```

### Building

```bash
go build -o nero cmd/nero/main.go
```

Or use the build script:
```bash
./build.sh        # Linux/Mac
.\build.ps1       # Windows
```

### Running Tests

```bash
go test ./...
```

### Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions

## Adding New Features

### Adding a New LLM Provider

1. Create a new file in `internal/models/` (e.g., `myprovider.go`)
2. Implement the `Provider` interface
3. Register the provider in `init()`
4. Update documentation

Example:
```go
package models

type MyProvider struct {
    client *someClient
}

func NewMyProvider() *MyProvider {
    return &MyProvider{}
}

func (p *MyProvider) Name() string {
    return "myprovider"
}

// Implement other Provider methods...

func init() {
    Register(NewMyProvider())
}
```

### Adding a New Tool

1. Create a tool struct in `internal/tools/`
2. Implement the `Tool` interface
3. Register the tool in `init()`

Example:
```go
package tools

type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Does something useful"
}

func (t *MyTool) Parameters() interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param": map[string]interface{}{
                "type": "string",
                "description": "A parameter",
            },
        },
        "required": []string{"param"},
    }
}

func (t *MyTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
    // Implementation
    return "result", nil
}

func init() {
    Register(&MyTool{})
}
```

### Adding a New CLI Command

1. Create a command file in `internal/cli/` (e.g., `mycommand.go`)
2. Define the cobra command
3. Add it to the root command in `root.go`

## Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new features
3. **Ensure all tests pass**
4. **Update CHANGELOG.md** with your changes
5. **Submit the pull request** with a clear description

## Code Review

- Be open to feedback
- Respond to comments promptly
- Make requested changes
- Keep discussions professional and constructive

## Reporting Issues

When reporting issues, please include:
- Nero version (`nero version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or error messages

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open an issue for discussion or reach out to the maintainers.

---

Thank you for contributing to Nero! 🚀
