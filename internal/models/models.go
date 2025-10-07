package models

import (
	"context"
	"fmt"
)

// Message represents a conversation message
type Message struct {
	Role       string `json:"role"` // user, assistant, system, tool
	Content    string `json:"content"`
	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// Provider interface for LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Chat sends a message and gets a response
	Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error)

	// Stream sends a message and streams the response
	Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error)

	// ListModels returns available models
	ListModels(ctx context.Context) ([]Model, error)
}

// Options for LLM requests
type Options struct {
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	Tools       []Tool
	ToolChoice  string
}

// Response from LLM
type Response struct {
	Content      string
	Role         string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        Usage
}

// StreamChunk for streaming responses
type StreamChunk struct {
	Content  string
	Role     string
	ToolCall *ToolCall
	Done     bool
	Error    error
}

// ToolCall represents a function call from the LLM
type ToolCall struct {
	ID       string
	Type     string
	Function FunctionCall
}

// FunctionCall details
type FunctionCall struct {
	Name      string
	Arguments string
}

// Tool definition
type Tool struct {
	Type     string
	Function Function
}

// Function definition
type Function struct {
	Name        string
	Description string
	Parameters  interface{}
}

// Usage statistics
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Model information
type Model struct {
	ID           string
	Provider     string
	ContextSize  int
	Capabilities []string
}

// Registry holds all registered providers
type Registry struct {
	providers map[string]Provider
}

var registry = &Registry{
	providers: make(map[string]Provider),
}

// Register registers a provider
func Register(provider Provider) {
	registry.providers[provider.Name()] = provider
}

// Get gets a provider by name
func Get(name string) (Provider, error) {
	provider, ok := registry.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// List returns all registered providers
func List() []string {
	names := make([]string, 0, len(registry.providers))
	for name := range registry.providers {
		names = append(names, name)
	}
	return names
}
