package models

import (
	"context"
	"io"

	"nero/internal/config"

	"github.com/sashabaranov/go-openai"
)

// GroqProvider implements Provider for Groq (OpenAI-compatible, ultra-fast inference)
type GroqProvider struct {
	client *openai.Client
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider() *GroqProvider {
	cfg := config.Get()

	// Groq uses OpenAI-compatible API
	config := openai.DefaultConfig(cfg.GroqKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client := openai.NewClientWithConfig(config)

	return &GroqProvider{
		client: client,
	}
}

// Name returns the provider name
func (p *GroqProvider) Name() string {
	return "groq"
}

// Chat sends a message and gets a response
func (p *GroqProvider) Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
	// Convert messages
	oaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		oaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	// Convert tools if provided
	var tools []openai.Tool
	if len(opts.Tools) > 0 {
		tools = make([]openai.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:       opts.Model,
		Messages:    oaiMessages,
		MaxTokens:   opts.MaxTokens,
		Temperature: float32(opts.Temperature),
		TopP:        float32(opts.TopP),
		Tools:       tools,
	}

	// Send request
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	choice := resp.Choices[0]

	// Convert response
	result := &Response{
		Content:      choice.Message.Content,
		Role:         choice.Message.Role,
		FinishReason: string(choice.FinishReason),
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Convert tool calls
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return result, nil
}

// Stream sends a message and streams the response
func (p *GroqProvider) Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error) {
	// Convert messages
	oaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		oaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:       opts.Model,
		Messages:    oaiMessages,
		MaxTokens:   opts.MaxTokens,
		Temperature: float32(opts.Temperature),
		TopP:        float32(opts.TopP),
		Stream:      true,
	}

	// Create stream
	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create channel
	ch := make(chan StreamChunk)

	// Start goroutine to read stream
	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				ch <- StreamChunk{Done: true}
				return
			}
			if err != nil {
				ch <- StreamChunk{Error: err}
				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			delta := response.Choices[0].Delta
			chunk := StreamChunk{
				Content: delta.Content,
				Role:    delta.Role,
			}

			// Handle tool calls
			if len(delta.ToolCalls) > 0 {
				tc := delta.ToolCalls[0]
				chunk.ToolCall = &ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}

			ch <- chunk
		}
	}()

	return ch, nil
}

// ListModels returns available models
func (p *GroqProvider) ListModels(ctx context.Context) ([]Model, error) {
	// Groq's latest models (ultra-fast inference)
	return []Model{
		{
			ID:           "llama-3.1-8b-instant",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "function_calling"},
		},
		{
			ID:           "llama-3.3-70b-versatile",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "function_calling"},
		},
		{
			ID:           "openai/gpt-oss-120b",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "openai/gpt-oss-20b",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "meta-llama/llama-guard-4-12b",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "guard"},
		},
		{
			ID:           "groq/compound",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "function_calling", "tools"},
		},
		{
			ID:           "groq/compound-mini",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "function_calling", "tools"},
		},
		{
			ID:           "moonshotai/kimi-k2-instruct-0905",
			Provider:     "groq",
			ContextSize:  262144,
			Capabilities: []string{"text", "function_calling"},
		},
		{
			ID:           "qwen/qwen3-32b",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "function_calling"},
		},
		{
			ID:           "meta-llama/llama-4-maverick-17b-128e-instruct",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "meta-llama/llama-4-scout-17b-16e-instruct",
			Provider:     "groq",
			ContextSize:  131072,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
	}, nil
}

func init() {
	Register(NewGroqProvider())
}
