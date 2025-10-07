package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool represents an executable tool
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Parameters returns the JSON schema for parameters
	Parameters() interface{}

	// Execute runs the tool
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

// Registry holds all registered tools
type Registry struct {
	tools map[string]Tool
}

var registry = &Registry{
	tools: make(map[string]Tool),
}

// Register registers a tool
func Register(tool Tool) {
	registry.tools[tool.Name()] = tool
}

// Get gets a tool by name
func Get(name string) (Tool, error) {
	tool, ok := registry.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// List returns all registered tools
func List() []Tool {
	tools := make([]Tool, 0, len(registry.tools))
	for _, tool := range registry.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute executes a tool by name
func Execute(ctx context.Context, name string, args json.RawMessage) (string, error) {
	tool, err := Get(name)
	if err != nil {
		return "", err
	}
	return tool.Execute(ctx, args)
}
