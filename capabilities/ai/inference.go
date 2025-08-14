package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type ModelSize string

const (
	ModelHelper ModelSize = "helper" // <1B parameters - for parsing/classification
	ModelSmall  ModelSize = "small"  // 1-7B parameters - for simple tasks
	ModelLarge  ModelSize = "large"  // 7B+ parameters - for complex reasoning
)

type Provider interface {
	Chat(ctx context.Context, messages []Message, stream chan<- string) error
	GetModelSize() ModelSize
	IsLocal() bool
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Router struct {
	providers   map[string]Provider
	helperModel Provider
	mainModel   Provider
	mutex       sync.RWMutex
}

func NewRouter() *Router {
	return &Router{
		providers: make(map[string]Provider),
	}
}

func (r *Router) RegisterProvider(name string, provider Provider) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.providers[name] = provider

	// Auto-assign based on model size
	switch provider.GetModelSize() {
	case ModelHelper:
		if r.helperModel == nil {
			r.helperModel = provider
		}
	case ModelLarge:
		if r.mainModel == nil {
			r.mainModel = provider
		}
	}
}

func (r *Router) GetHelperModel() Provider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.helperModel
}

func (r *Router) GetMainModel() Provider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.mainModel
}

func (r *Router) SetMainModel(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	provider, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	r.mainModel = provider
	return nil
}

// OllamaProvider for local models
type OllamaProvider struct {
	baseURL   string
	modelName string
	modelSize ModelSize
}

func NewOllamaProvider(baseURL, modelName string, size ModelSize) *OllamaProvider {
	return &OllamaProvider{
		baseURL:   baseURL,
		modelName: modelName,
		modelSize: size,
	}
}

func (o *OllamaProvider) Chat(ctx context.Context, messages []Message, stream chan<- string) error {
	defer close(stream)

	reqBody := map[string]interface{}{
		"model":    o.modelName,
		"messages": messages,
		"stream":   true,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			continue
		}

		if msg, ok := response["message"].(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				select {
				case stream <- content:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		if done, ok := response["done"].(bool); ok && done {
			break
		}
	}

	return scanner.Err()
}

func (o *OllamaProvider) GetModelSize() ModelSize {
	return o.modelSize
}

func (o *OllamaProvider) IsLocal() bool {
	return true
}

// CloudProvider for OpenAI/Groq/etc
type CloudProvider struct {
	name      string
	apiKey    string
	baseURL   string
	modelName string
	modelSize ModelSize
}

func NewCloudProvider(name, apiKey, baseURL, modelName string, size ModelSize) *CloudProvider {
	return &CloudProvider{
		name:      name,
		apiKey:    apiKey,
		baseURL:   baseURL,
		modelName: modelName,
		modelSize: size,
	}
}

func (c *CloudProvider) Chat(ctx context.Context, messages []Message, stream chan<- string) error {
	defer close(stream)

	reqBody := map[string]interface{}{
		"model":    c.modelName,
		"messages": messages,
		"stream":   true,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						select {
						case stream <- content:
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}
			}
		}
	}

	return nil
}

func (c *CloudProvider) GetModelSize() ModelSize {
	return c.modelSize
}

func (c *CloudProvider) IsLocal() bool {
	return false
}

// LoadProviders initializes all available providers
func LoadProviders() *Router {
	router := NewRouter()

	// Try to load Ollama models
	if err := loadOllamaModels(router); err == nil {
		// Ollama available
	}

	// Load cloud providers
	loadCloudProviders(router)

	return router
}

func loadOllamaModels(router *Router) error {
	// Check if Ollama is running by listing available models
	models, err := getOllamaModels("http://localhost:11434")
	if err != nil {
		return err
	}

	// Try to find helper models for parsing
	helperModels := []string{"qwen2.5:0.5b", "smollm2:135m", "phi4:mini", "gemma2:2b"}
	for _, model := range helperModels {
		if contains(models, model) {
			router.RegisterProvider("ollama-helper", NewOllamaProvider("http://localhost:11434", model, ModelHelper))
			break
		}
	}

	// Try to find main models
	mainModels := []string{"qwen2.5:7b", "llama3.2:3b", "phi3.5:3.8b", "gemma2:9b", "llama3.1:8b"}
	for _, model := range mainModels {
		if contains(models, model) {
			router.RegisterProvider("ollama-main", NewOllamaProvider("http://localhost:11434", model, ModelLarge))
			return nil
		}
	}

	// If no main model found but we have any model, use the first one
	if len(models) > 0 {
		router.RegisterProvider("ollama-main", NewOllamaProvider("http://localhost:11434", models[0], ModelLarge))
	}

	return nil
}

func getOllamaModels(baseURL string) ([]string, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama not available (status: %d)", resp.StatusCode)
	}

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var models []string
	for _, model := range response.Models {
		models = append(models, model.Name)
	}

	return models, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func loadCloudProviders(router *Router) {
	// Load API keys
	if apiKey := loadAPIKey("openai.key"); apiKey != "" {
		router.RegisterProvider("openai", NewCloudProvider("openai", apiKey, "https://api.openai.com/v1", "gpt-4o-mini", ModelLarge))
	}

	if apiKey := loadAPIKey("groq.key"); apiKey != "" {
		router.RegisterProvider("groq", NewCloudProvider("groq", apiKey, "https://api.groq.com/openai/v1", "llama-3.1-8b-instant", ModelLarge))
	}

	if apiKey := loadAPIKey("gemini.key"); apiKey != "" {
		// Gemini uses different API format, would need separate provider
	}
}

func loadAPIKey(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func testOllamaModel(baseURL, model string) bool {
	req, err := http.NewRequest("POST", baseURL+"/api/generate", strings.NewReader(`{"model":"`+model+`","prompt":"test","stream":false}`))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
