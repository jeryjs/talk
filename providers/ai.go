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
	"time"
)

// Define the interface for AI model providers
type AIProvider interface {
	Name() string
	Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error)
	IsAvailable() bool
}

// Define streaming capability interface
type StreamingProvider interface {
	AIProvider
	ChatStream(ctx context.Context, messages []Message, options *ChatOptions, callback StreamCallback) error
}

// Define vision capability interface
type VisionProvider interface {
	AIProvider
	SupportsVision() bool
	ChatWithVision(ctx context.Context, messages []VisionMessage, options *ChatOptions) (*Response, error)
}

// Define reasoning capability interface
type ReasoningProvider interface {
	AIProvider
	SupportsReasoning() bool
	ChatWithReasoning(ctx context.Context, messages []Message, options *ChatOptions) (*ReasoningResponse, error)
}

// Callback for streaming responses
type StreamCallback func(chunk StreamChunk)

// Represent a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Represent a vision-enabled message
type VisionMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// Represent content parts for multimodal messages
type ContentPart struct {
	Type     string    `json:"type"` // "text" or "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// Represent image URL data
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

// Configure chat behavior
type ChatOptions struct {
	Temperature    float64 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"max_tokens,omitempty"`
	SystemPrompt   string  `json:"system_prompt,omitempty"`
	Stream         bool    `json:"stream,omitempty"`
	EnableVision   bool    `json:"enable_vision,omitempty"`
	EnableThoughts bool    `json:"enable_thoughts,omitempty"`
	TopP           float64 `json:"top_p,omitempty"`
}

// Represent an AI response
type Response struct {
	Content    string
	TokensUsed int
	Model      string
	Metadata   map[string]interface{}
}

// Represent a reasoning response with thoughts
type ReasoningResponse struct {
	Response
	Thoughts     string
	ThoughtsTime time.Duration
}

// Represent a streaming chunk
type StreamChunk struct {
	Content   string
	Type      string // "text", "reasoning", "vision"
	IsThought bool
	Delta     bool
	Done      bool
	Error     error
	Metadata  map[string]interface{}
}

// Implement local Ollama integration with streaming
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
	if options != nil && options.Stream {
		// Use streaming but collect full response
		var fullResponse strings.Builder
		err := o.ChatStream(ctx, messages, options, func(chunk StreamChunk) {
			if !chunk.Done && chunk.Error == nil {
				fullResponse.WriteString(chunk.Content)
			}
		})
		if err != nil {
			return nil, err
		}

		return &Response{
			Content: fullResponse.String(),
			Model:   o.model,
			Metadata: map[string]interface{}{
				"provider": "ollama",
			},
		}, nil
	}

	return o.chatNonStreaming(ctx, messages, options)
}

func (o *OllamaProvider) chatNonStreaming(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
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

// Implement streaming for Ollama
func (o *OllamaProvider) ChatStream(ctx context.Context, messages []Message, options *ChatOptions, callback StreamCallback) error {
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

	// Prepare streaming request
	reqData := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt.String(),
		"stream": true,
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
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama error: %s", string(body))
	}

	// Process streaming response
	decoder := json.NewDecoder(resp.Body)
	for {
		var result struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}

		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			callback(StreamChunk{Error: err, Done: true})
			return err
		}

		// Send chunk to callback
		callback(StreamChunk{
			Content: result.Response,
			Type:    "text",
			Delta:   true,
			Done:    result.Done,
		})

		if result.Done {
			break
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// Implement OpenAI/compatible API with vision and streaming
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

func (o *OpenAIProvider) SupportsVision() bool {
	return strings.Contains(o.model, "vision") || strings.Contains(o.model, "4o")
}

func (o *OpenAIProvider) Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
	if !o.IsAvailable() {
		return nil, fmt.Errorf("OpenAI API key not available")
	}

	if options != nil && options.Stream {
		// Use streaming but collect full response
		var fullResponse strings.Builder
		err := o.ChatStream(ctx, messages, options, func(chunk StreamChunk) {
			if !chunk.Done && chunk.Error == nil {
				fullResponse.WriteString(chunk.Content)
			}
		})
		if err != nil {
			return nil, err
		}

		return &Response{
			Content: fullResponse.String(),
			Model:   o.model,
			Metadata: map[string]interface{}{
				"provider": "openai",
			},
		}, nil
	}

	return o.chatNonStreaming(ctx, messages, options)
}

func (o *OpenAIProvider) chatNonStreaming(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
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

// Implement streaming for OpenAI
func (o *OpenAIProvider) ChatStream(ctx context.Context, messages []Message, options *ChatOptions, callback StreamCallback) error {
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
		"stream":   true,
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
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai error: %s", string(body))
	}

	// Process streaming response
	reader := resp.Body
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			callback(StreamChunk{Error: err, Done: true})
			return err
		}

		// Parse SSE data
		data := string(buffer[:n])
		lines := strings.Split(data, "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")
				if jsonData == "[DONE]" {
					callback(StreamChunk{Done: true})
					return nil
				}

				var result struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
					continue // Skip malformed JSON
				}

				if len(result.Choices) > 0 {
					content := result.Choices[0].Delta.Content
					if content != "" {
						callback(StreamChunk{
							Content: content,
							Type:    "text",
							Delta:   true,
							Done:    false,
						})
					}
				}
			}
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// Implement Google Gemini provider
type GeminiProvider struct {
	apiKey string
	model  string
}

// Create a new Gemini provider
func NewGeminiProvider(model string) *GeminiProvider {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Try reading from file
		if data, err := os.ReadFile("gemini.key"); err == nil {
			apiKey = strings.TrimSpace(string(data))
		}
	}

	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

func (g *GeminiProvider) IsAvailable() bool {
	return g.apiKey != ""
}

func (g *GeminiProvider) SupportsVision() bool {
	return strings.Contains(g.model, "vision") || strings.Contains(g.model, "2.0")
}

func (g *GeminiProvider) Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("Gemini API key not available")
	}

	// Convert messages to Gemini format
	var contents []map[string]interface{}

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}

	reqData := map[string]interface{}{
		"contents": contents,
	}

	if options != nil {
		config := map[string]interface{}{}
		if options.Temperature > 0 {
			config["temperature"] = options.Temperature
		}
		if options.MaxTokens > 0 {
			config["maxOutputTokens"] = options.MaxTokens
		}
		if len(config) > 0 {
			reqData["generationConfig"] = config
		}

		if options.SystemPrompt != "" {
			reqData["systemInstruction"] = map[string]interface{}{
				"parts": []map[string]string{
					{"text": options.SystemPrompt},
				},
			}
		}
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("gemini error: %s", string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			TotalTokenCount int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	return &Response{
		Content:    result.Candidates[0].Content.Parts[0].Text,
		TokensUsed: result.UsageMetadata.TotalTokenCount,
		Model:      g.model,
		Metadata: map[string]interface{}{
			"provider": "gemini",
		},
	}, nil
}

// Implement Groq provider (uses OpenAI-compatible API)
type GroqProvider struct {
	apiKey string
	model  string
}

// Create a new Groq provider
func NewGroqProvider(model string) *GroqProvider {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		// Try reading from file
		if data, err := os.ReadFile("groq.key"); err == nil {
			apiKey = strings.TrimSpace(string(data))
		}
	}

	return &GroqProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (g *GroqProvider) Name() string {
	return "groq"
}

func (g *GroqProvider) IsAvailable() bool {
	return g.apiKey != ""
}

func (g *GroqProvider) Chat(ctx context.Context, messages []Message, options *ChatOptions) (*Response, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("Groq API key not available")
	}

	// Groq uses OpenAI-compatible API
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
		"model":    g.model,
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq error: %s", string(body))
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
		return nil, fmt.Errorf("no response from Groq")
	}

	return &Response{
		Content:    result.Choices[0].Message.Content,
		TokensUsed: result.Usage.TotalTokens,
		Model:      g.model,
		Metadata: map[string]interface{}{
			"provider": "groq",
		},
	}, nil
}
