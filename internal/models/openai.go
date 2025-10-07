package models

import (
	"context"
	"fmt"
	"io"

	"nero/internal/config"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements Provider for OpenAI
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	cfg := config.Get()
	client := openai.NewClient(cfg.OpenAIKey)

	return &OpenAIProvider{
		client: client,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Chat sends a message and gets a response
func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
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
	}

	// Add tools if provided
	if len(opts.Tools) > 0 {
		tools := make([]openai.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		req.Tools = tools
	}

	// Send request
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
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
func (p *OpenAIProvider) Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error) {
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

	// Add tools if provided
	if len(opts.Tools) > 0 {
		tools := make([]openai.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		req.Tools = tools
	}

	// Create stream
	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
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
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]Model, error) {
	// Latest OpenAI models as of 2025
	return []Model{
		{
			ID:           "gpt-5",
			Provider:     "openai",
			ContextSize:  400000,
			Capabilities: []string{"text", "vision", "audio", "function_calling", "reasoning"},
		},
		{
			ID:           "gpt-5-mini",
			Provider:     "openai",
			ContextSize:  400000,
			Capabilities: []string{"text", "vision", "function_calling", "reasoning"},
		},
		{
			ID:           "gpt-5-nano",
			Provider:     "openai",
			ContextSize:  400000,
			Capabilities: []string{"text", "vision", "function_calling", "reasoning"},
		},
		{
			ID:           "gpt-4.1",
			Provider:     "openai",
			ContextSize:  1047576,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "gpt-4o",
			Provider:     "openai",
			ContextSize:  128000,
			Capabilities: []string{"text", "vision", "audio", "function_calling"},
		},
	}, nil
}

func init() {
	Register(NewOpenAIProvider())
}
