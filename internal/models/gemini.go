package models

import (
	"context"
	"encoding/json"
	"fmt"

	"nero/internal/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements Provider for Google Gemini
type GeminiProvider struct {
	client *genai.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider() *GeminiProvider {
	cfg := config.Get()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GoogleKey))
	if err != nil {
		// Return provider anyway, will error on use
		return &GeminiProvider{}
	}

	return &GeminiProvider{
		client: client,
	}
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// Chat sends a message and gets a response
func (p *GeminiProvider) Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
	if p.client == nil {
		return nil, fmt.Errorf("gemini client not initialized (check API key)")
	}

	model := p.client.GenerativeModel(opts.Model)
	model.SetTemperature(float32(opts.Temperature))
	model.SetMaxOutputTokens(int32(opts.MaxTokens))

	// Convert tools if provided
	if len(opts.Tools) > 0 {
		tools := make([]*genai.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			// Convert parameters to proper schema
			paramsJSON, _ := json.Marshal(tool.Function.Parameters)
			var schema *genai.Schema
			json.Unmarshal(paramsJSON, &schema)

			tools[i] = &genai.Tool{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        tool.Function.Name,
						Description: tool.Function.Description,
						Parameters:  schema,
					},
				},
			}
		}
		model.Tools = tools
	}

	// Start chat session
	chat := model.StartChat()

	// Add history (except last user message)
	for i := 0; i < len(messages)-1; i++ {
		msg := messages[i]
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		chat.History = append(chat.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Send last message
	lastMsg := messages[len(messages)-1]
	resp, err := chat.SendMessage(ctx, genai.Text(lastMsg.Content))
	if err != nil {
		return nil, fmt.Errorf("gemini chat failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}

	// Extract response
	result := &Response{
		Role: "assistant",
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result.Content += string(text)
		}
		if fc, ok := part.(genai.FunctionCall); ok {
			argsJSON, _ := json.Marshal(fc.Args)
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:   fc.Name, // Gemini doesn't provide ID
				Type: "function",
				Function: FunctionCall{
					Name:      fc.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	// Add usage stats if available
	if resp.UsageMetadata != nil {
		result.Usage = Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return result, nil
}

// Stream sends a message and streams the response
func (p *GeminiProvider) Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error) {
	if p.client == nil {
		ch := make(chan StreamChunk, 1)
		ch <- StreamChunk{Error: fmt.Errorf("gemini client not initialized")}
		close(ch)
		return ch, nil
	}

	model := p.client.GenerativeModel(opts.Model)
	model.SetTemperature(float32(opts.Temperature))
	model.SetMaxOutputTokens(int32(opts.MaxTokens))

	chat := model.StartChat()

	// Add history
	for i := 0; i < len(messages)-1; i++ {
		msg := messages[i]
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		chat.History = append(chat.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Send message and stream
	lastMsg := messages[len(messages)-1]
	iter := chat.SendMessageStream(ctx, genai.Text(lastMsg.Content))

	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() != "iterator exhausted" {
					ch <- StreamChunk{Error: err}
				} else {
					ch <- StreamChunk{Done: true}
				}
				return
			}

			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				continue
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					ch <- StreamChunk{
						Content: string(text),
						Role:    "assistant",
					}
				}
			}
		}
	}()

	return ch, nil
}

// ListModels returns available models
func (p *GeminiProvider) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{
		{
			ID:           "gemini-2.5-pro",
			Provider:     "gemini",
			ContextSize:  1048576,
			Capabilities: []string{"text", "vision", "thinking", "function_calling"},
		},
		{
			ID:           "gemini-2.5-flash",
			Provider:     "gemini",
			ContextSize:  1048576,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		{
			ID:           "gemini-2.5-flash-lite",
			Provider:     "gemini",
			ContextSize:  1048576,
			Capabilities: []string{"text", "vision", "thinking", "function_calling"},
		},
		{
			ID:           "gemini-2.0-flash-exp",
			Provider:     "gemini",
			ContextSize:  1000000,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
	}, nil
}

func init() {
	Register(NewGeminiProvider())
}
