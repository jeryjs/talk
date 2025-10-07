package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaProvider implements Provider for Ollama (local models)
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		baseURL: "http://localhost:11434",
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  ollamaOptions   `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

type ollamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

// Chat sends a message and gets a response
func (p *OllamaProvider) Chat(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
	// Convert messages
	ollamaMessages := make([]ollamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request
	reqBody := ollamaChatRequest{
		Model:    opts.Model,
		Messages: ollamaMessages,
		Stream:   false,
		Options: ollamaOptions{
			Temperature: opts.Temperature,
			NumPredict:  opts.MaxTokens,
			TopP:        opts.TopP,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: %s", string(body))
	}

	// Parse response
	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Response{
		Content: chatResp.Message.Content,
		Role:    chatResp.Message.Role,
		Usage: Usage{
			// Ollama doesn't provide token counts in non-streaming mode
			TotalTokens: 0,
		},
	}, nil
}

// Stream sends a message and streams the response
func (p *OllamaProvider) Stream(ctx context.Context, messages []Message, opts *Options) (<-chan StreamChunk, error) {
	// Convert messages
	ollamaMessages := make([]ollamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request
	reqBody := ollamaChatRequest{
		Model:    opts.Model,
		Messages: ollamaMessages,
		Stream:   true,
		Options: ollamaOptions{
			Temperature: opts.Temperature,
			NumPredict:  opts.MaxTokens,
			TopP:        opts.TopP,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: %s", string(body))
	}

	// Create channel
	ch := make(chan StreamChunk)

	// Start goroutine to read stream
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chatResp ollamaChatResponse
			if err := decoder.Decode(&chatResp); err != nil {
				if err == io.EOF {
					ch <- StreamChunk{Done: true}
					return
				}
				ch <- StreamChunk{Error: err}
				return
			}

			if chatResp.Done {
				ch <- StreamChunk{Done: true}
				return
			}

			ch <- StreamChunk{
				Content: chatResp.Message.Content,
				Role:    chatResp.Message.Role,
			}
		}
	}()

	return ch, nil
}

type ollamaModelsResponse struct {
	Models []struct {
		Name       string `json:"name"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modified_at"`
	} `json:"models"`
}

// ListModels returns available models
func (p *OllamaProvider) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	var modelsResp ollamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]Model, len(modelsResp.Models))
	for i, m := range modelsResp.Models {
		models[i] = Model{
			ID:           m.Name,
			Provider:     "ollama",
			ContextSize:  8192, // Default, varies by model
			Capabilities: []string{"text"},
		}
	}

	return models, nil
}

func init() {
	Register(NewOllamaProvider())
}
