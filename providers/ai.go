package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Define the interface for AI model providers
type AIProvider interface {
	Name() string
	Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error)
	IsAvailable() bool
}

// Represent a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Configure chat behavior
type ChatOptions struct {
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
	Stream       bool    `json:"stream,omitempty"`
}

// Represent an AI response
type Response struct {
	Content    string
	TokensUsed int
	Model      string
	Metadata   map[string]interface{}
}

// Implement local Ollama integration
type OllamaProvider struct {
	baseURL string
	model   string
}

// Create a new Ollama provider
func NewOllamaProvider(model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: "http://localhost:11434",
		model:   model,
	}
}

func (o *OllamaProvider) Name() string {
	return "ollama"
}

func (o *OllamaProvider) IsAvailable() bool {
	resp, err := http.Get(o.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (o *OllamaProvider) Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
	// Build prompt from messages
	var prompt strings.Builder

	// Add system prompt if provided
	if options != nil && options.SystemPrompt != "" {
		prompt.WriteString("System: " + options.SystemPrompt + "\n\n")
	}

	// Add conversation history
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			prompt.WriteString("Human: " + msg.Content + "\n\n")
		case "assistant":
			prompt.WriteString("Assistant: " + msg.Content + "\n\n")
		}
	}

	prompt.WriteString("Assistant: ")

	// Prepare request
	reqData := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt.String(),
		"stream": false,
	}

	if options != nil {
		if options.Temperature > 0 {
			reqData["options"] = map[string]interface{}{
				"temperature": options.Temperature,
			}
		}
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	// Make request to Ollama API
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: %s", string(body))
	}

	// Parse Ollama response
	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Response{
		Content: strings.TrimSpace(result.Response),
		Model:   o.model,
		Metadata: map[string]interface{}{
			"provider": "ollama",
		},
	}, nil
}

// Implement OpenAI/compatible API
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
}

// Create a new OpenAI provider
func NewOpenAIProvider(model string) *OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Try reading from file
		if data, err := os.ReadFile("openai.key"); err == nil {
			apiKey = strings.TrimSpace(string(data))
		}
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

func (o *OpenAIProvider) IsAvailable() bool {
	return o.apiKey != ""
}

func (o *OpenAIProvider) Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
	if !o.IsAvailable() {
		return nil, fmt.Errorf("OpenAI API key not available")
	}

	// Prepare messages for OpenAI format
	var apiMessages []map[string]string

	if options != nil && options.SystemPrompt != "" {
		apiMessages = append(apiMessages, map[string]string{
			"role":    "system",
			"content": options.SystemPrompt,
		})
	}

	for _, msg := range messages {
		apiMessages = append(apiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqData := map[string]interface{}{
		"model":    o.model,
		"messages": apiMessages,
	}

	if options != nil {
		if options.Temperature > 0 {
			reqData["temperature"] = options.Temperature
		}
		if options.MaxTokens > 0 {
			reqData["max_tokens"] = options.MaxTokens
		}
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &Response{
		Content:    result.Choices[0].Message.Content,
		TokensUsed: result.Usage.TotalTokens,
		Model:      o.model,
		Metadata: map[string]interface{}{
			"provider": "openai",
		},
	}, nil
}
