package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nero/internal/config"
	"nero/internal/models"
	"nero/internal/tools"
)

// Agent represents an AI agent
type Agent struct {
	provider      models.Provider
	model         models.Model
	conversation  []models.Message
	maxIterations int
}

// New creates a new agent
func New() (*Agent, error) {
	cfg := config.Get()

	// Parse provider/model format
	// Supports: "provider/model", "shortcut/model", or just "model"
	// Shortcuts: g/ = gemini, op/ = openai, op/ = ollama, c/ = claude, gq/ = groq
	providerName, modelID := parseModelString(cfg.Model)

	provider, err := models.Get(providerName)
	if err != nil {
		// Always fallback to Ollama (local, no API key needed)
		fmt.Printf("⚠️  Provider '%s' not available, falling back to Ollama (local)\n", providerName)
		provider, err = models.Get("ollama")
		if err != nil {
			return nil, fmt.Errorf("no providers available (even Ollama failed): %w", err)
		}
		providerName = "ollama"
		modelID = "llama3.2" // Default Ollama model
	}

	// Create model struct
	model := models.Model{
		ID:       modelID,
		Provider: providerName,
	}

	return &Agent{
		provider:      provider,
		model:         model,
		conversation:  make([]models.Message, 0),
		maxIterations: 10,
	}, nil
}

// parseModelString parses "provider/model" format with shortcuts
// Examples:
//
//	"gpt-4o" -> ("openai", "gpt-4o")
//	"op/gpt-4o" -> ("openai", "gpt-4o")
//	"gemini-2.0" -> ("gemini", "gemini-2.0")
//	"g/gemini-2.0" -> ("gemini", "gemini-2.0")
//	"ol/llama3.2" -> ("ollama", "llama3.2")
//	"gq/llama-3.1-70b" -> ("groq", "llama-3.1-70b")
func parseModelString(model string) (provider, modelName string) {
	// Check for provider/model format
	if strings.Contains(model, "/") {
		parts := strings.SplitN(model, "/", 2)
		shortcut := parts[0]
		modelName = parts[1]

		// Expand shortcuts
		switch shortcut {
		case "g":
			return "gemini", modelName
		case "op":
			return "openai", modelName
		case "ol":
			return "ollama", modelName
		case "c":
			return "anthropic", modelName
		case "gq":
			return "groq", modelName
		case "or":
			return "openrouter", modelName
		default:
			return shortcut, modelName
		}
	}

	// Auto-detect from model name
	modelName = model
	if strings.HasPrefix(model, "gpt") {
		return "openai", modelName
	} else if strings.HasPrefix(model, "claude") {
		return "anthropic", modelName
	} else if strings.HasPrefix(model, "gemini") {
		return "gemini", modelName
	} else if strings.HasPrefix(model, "llama") || strings.HasPrefix(model, "qwen") || strings.HasPrefix(model, "deepseek") {
		return "ollama", modelName
	}

	// Default to Ollama for unknown models
	return "ollama", modelName
}

// Chat sends a message and gets a response
func (a *Agent) Chat(ctx context.Context, message string) (string, error) {
	// Add user message
	a.conversation = append(a.conversation, models.Message{
		Role:    "user",
		Content: message,
	})

	cfg := config.Get()

	// Get available tools
	availableTools := tools.List()
	toolDefs := make([]models.Tool, len(availableTools))
	for i, tool := range availableTools {
		toolDefs[i] = models.Tool{
			Type: "function",
			Function: models.Function{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
	}

	opts := &models.Options{
		Model:       a.model.ID, // Use the actual model ID, not the shortcut
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Tools:       toolDefs,
	}

	// Loop to handle tool calls
	for i := 0; i < a.maxIterations; i++ {
		resp, err := a.provider.Chat(ctx, a.conversation, opts)
		if err != nil {
			return "", fmt.Errorf("chat failed: %w", err)
		}

		// Add assistant message
		a.conversation = append(a.conversation, models.Message{
			Role:    resp.Role,
			Content: resp.Content,
		})

		// If no tool calls, we're done
		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		// Execute tool calls
		for _, tc := range resp.ToolCalls {
			result, err := tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Add tool result to conversation
			a.conversation = append(a.conversation, models.Message{
				Role:       "tool",
				Content:    result,
				Name:       tc.Function.Name,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", fmt.Errorf("max iterations reached")
}

// Stream sends a message and streams the response
func (a *Agent) Stream(ctx context.Context, message string) (<-chan string, error) {
	// Add user message
	a.conversation = append(a.conversation, models.Message{
		Role:    "user",
		Content: message,
	})

	cfg := config.Get()

	// Get available tools
	availableTools := tools.List()
	toolDefs := make([]models.Tool, len(availableTools))
	for i, tool := range availableTools {
		toolDefs[i] = models.Tool{
			Type: "function",
			Function: models.Function{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
	}

	opts := &models.Options{
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Tools:       toolDefs,
	}

	streamCh, err := a.provider.Stream(ctx, a.conversation, opts)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outCh := make(chan string)

	go func() {
		defer close(outCh)

		var fullContent strings.Builder
		var toolCalls []models.ToolCall

		for chunk := range streamCh {
			if chunk.Error != nil {
				outCh <- fmt.Sprintf("\nError: %v\n", chunk.Error)
				return
			}

			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
				outCh <- chunk.Content
			}

			if chunk.ToolCall != nil {
				toolCalls = append(toolCalls, *chunk.ToolCall)
			}

			if chunk.Done {
				// Add assistant message
				a.conversation = append(a.conversation, models.Message{
					Role:    "assistant",
					Content: fullContent.String(),
				})

				// Execute tool calls if any
				if len(toolCalls) > 0 {
					outCh <- "\n\n[Executing tools...]\n"

					for _, tc := range toolCalls {
						result, err := tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
						if err != nil {
							result = fmt.Sprintf("Error: %v", err)
						}

						outCh <- fmt.Sprintf("Tool %s: %s\n", tc.Function.Name, result)

						// Add tool result
						a.conversation = append(a.conversation, models.Message{
							Role:       "tool",
							Content:    result,
							Name:       tc.Function.Name,
							ToolCallID: tc.ID,
						})
					}
				}
			}
		}
	}()

	return outCh, nil
}

// Reset clears the conversation history
func (a *Agent) Reset() {
	a.conversation = make([]models.Message, 0)
}

// GetModel returns the current model
func (a *Agent) GetModel() models.Model {
	return a.model
}

// GetProvider returns the current provider
func (a *Agent) GetProvider() models.Provider {
	return a.provider
}

// SwitchModel changes the model while preserving conversation
func (a *Agent) SwitchModel(providerName, modelID string) error {
	provider, err := models.Get(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	a.provider = provider
	a.model = models.Model{
		ID:       modelID,
		Provider: providerName,
	}

	return nil
}

// GetConversation returns the current conversation
func (a *Agent) GetConversation() []models.Message {
	return a.conversation
}

// SetSystemPrompt sets the system prompt
func (a *Agent) SetSystemPrompt(prompt string) {
	// Remove existing system message if any
	for i, msg := range a.conversation {
		if msg.Role == "system" {
			a.conversation = append(a.conversation[:i], a.conversation[i+1:]...)
			break
		}
	}

	// Add new system message at the beginning
	a.conversation = append([]models.Message{{
		Role:    "system",
		Content: prompt,
	}}, a.conversation...)
}
