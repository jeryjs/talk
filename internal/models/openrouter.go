package models

import (
	"context"
	"io"

	"nero/internal/config"

	"github.com/sashabaranov/go-openai"
)

// OpenRouterProvider implements Provider for OpenRouter (universal API gateway)
type OpenRouterProvider struct {
	client *openai.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider() *OpenRouterProvider {
	cfg := config.Get()

	// OpenRouter uses OpenAI-compatible API
	config := openai.DefaultConfig(cfg.OpenRouterKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	client := openai.NewClientWithConfig(config)

	return &OpenRouterProvider{
		client: client,
	}
}

// Name returns the provider name
func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

// Chat sends a message and gets a response
func (p *OpenRouterProvider) Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
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
func (p *OpenRouterProvider) Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error) {
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

// ListModels returns available models (popular ones from OpenRouter)
func (p *OpenRouterProvider) ListModels(ctx context.Context) ([]Model, error) {
	// OpenRouter has 500+ models, list most popular ones
	return []Model{
		{
			ID:           "anthropic/claude-sonnet-4.5",
			Provider:     "openrouter",
			ContextSize:  200000,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "google/gemini-2.5-pro",
			Provider:     "openrouter",
			ContextSize:  2000000,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "openai/gpt-4.5",
			Provider:     "openrouter",
			ContextSize:  200000,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "meta-llama/llama-3.3-70b-instruct",
			Provider:     "openrouter",
			ContextSize:  128000,
			Capabilities: []string{"text", "function_calling"},
		},
		{
			ID:           "deepseek/deepseek-r1",
			Provider:     "openrouter",
			ContextSize:  128000,
			Capabilities: []string{"text", "reasoning"},
		},
	}, nil
}

func init() {
	Register(NewOpenRouterProvider())
}
