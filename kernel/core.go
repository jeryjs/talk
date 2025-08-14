package kernel

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"nero/providers"
)

// Orchestrate all AI operations and provider management
type Core struct {
	providers     map[string]providers.AIProvider
	activeModel   string
	memory        *Memory
	config        *CoreConfig
	streamManager *StreamManager
	mu            sync.RWMutex
}

// Configure core behavior and capabilities
type CoreConfig struct {
	DefaultProvider    string
	FallbackProviders  []string
	StreamingEnabled   bool
	VisionEnabled      bool
	ReasoningEnabled   bool
	MaxConcurrentCalls int
	RequestTimeout     time.Duration
}

// Handle real-time response streaming
type StreamManager struct {
	activeStreams map[string]*StreamContext
	mu            sync.RWMutex
}

// Track individual streaming context
type StreamContext struct {
	ID        string
	Provider  string
	StartTime time.Time
	Channel   chan StreamChunk
	Cancel    context.CancelFunc
}

// Represent a chunk of streamed content
type StreamChunk struct {
	Content   string
	Type      string // "text", "reasoning", "vision"
	IsThought bool
	Delta     bool
	Done      bool
	Error     error
}

// Request for AI processing with full capabilities
type AIRequest struct {
	Messages       []providers.Message
	Provider       string // Optional: specify provider
	EnableStream   bool
	EnableVision   bool
	EnableThoughts bool
	Temperature    float64
	MaxTokens      int
	SystemPrompt   string
	Context        map[string]interface{}
}

// Response with enhanced capabilities
type AIResponse struct {
	Content     string
	Provider    string
	Model       string
	TokensUsed  int
	ProcessTime time.Duration
	HasVision   bool
	HasThoughts bool
	StreamID    string
	Metadata    map[string]interface{}
}

// Initialize the AI processing core
func NewCore() *Core {
	config := &CoreConfig{
		DefaultProvider:    "ollama",
		FallbackProviders:  []string{"openai", "gemini", "groq"},
		StreamingEnabled:   true,
		VisionEnabled:      true,
		ReasoningEnabled:   true,
		MaxConcurrentCalls: 10,
		RequestTimeout:     time.Minute * 2,
	}

	return &Core{
		providers:     make(map[string]providers.AIProvider),
		memory:        NewMemory(),
		config:        config,
		streamManager: NewStreamManager(),
	}
}

// Initialize all available AI providers
func (c *Core) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Register Ollama provider
	ollama := providers.NewOllamaProvider("llama3.2")
	if ollama.IsAvailable() {
		c.providers["ollama"] = ollama
		c.activeModel = "ollama"
	}

	// Register OpenAI provider
	openai := providers.NewOpenAIProvider("gpt-4o")
	if openai.IsAvailable() {
		c.providers["openai"] = openai
		if c.activeModel == "" {
			c.activeModel = "openai"
		}
	}

	// Register Gemini provider
	gemini := providers.NewGeminiProvider("gemini-2.0-flash-exp")
	if gemini.IsAvailable() {
		c.providers["gemini"] = gemini
		if c.activeModel == "" {
			c.activeModel = "gemini"
		}
	}

	// Register Groq provider
	groq := providers.NewGroqProvider("llama-3.3-70b-versatile")
	if groq.IsAvailable() {
		c.providers["groq"] = groq
		if c.activeModel == "" {
			c.activeModel = "groq"
		}
	}

	if len(c.providers) == 0 {
		return fmt.Errorf("no AI providers available - install Ollama or set API keys")
	}

	return nil
}

// Process AI request with intelligent provider selection
func (c *Core) ProcessRequest(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	provider := c.selectProvider(req)
	if provider == nil {
		return nil, fmt.Errorf("no suitable provider available")
	}

	startTime := time.Now()

	// Handle streaming vs non-streaming
	if req.EnableStream && c.config.StreamingEnabled {
		return c.processStreamingRequest(ctx, req, provider)
	}

	return c.processStandardRequest(ctx, req, provider, startTime)
}

// Process standard (non-streaming) request
func (c *Core) processStandardRequest(ctx context.Context, req *AIRequest, provider providers.AIProvider, startTime time.Time) (*AIResponse, error) {
	options := &providers.ChatOptions{
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		SystemPrompt: req.SystemPrompt,
		Stream:       false,
	}

	response, err := provider.Chat(ctx, req.Messages, options)
	if err != nil {
		return c.tryFallbackProvider(ctx, req, err)
	}

	return &AIResponse{
		Content:     response.Content,
		Provider:    provider.Name(),
		Model:       response.Model,
		TokensUsed:  response.TokensUsed,
		ProcessTime: time.Since(startTime),
		Metadata:    response.Metadata,
	}, nil
}

// Process streaming request with real-time capabilities
func (c *Core) processStreamingRequest(ctx context.Context, req *AIRequest, provider providers.AIProvider) (*AIResponse, error) {
	streamID := fmt.Sprintf("stream_%d", time.Now().UnixNano())

	streamCtx := &StreamContext{
		ID:        streamID,
		Provider:  provider.Name(),
		StartTime: time.Now(),
		Channel:   make(chan StreamChunk, 100),
	}

	streamCtx.Cancel = func() {
		close(streamCtx.Channel)
		c.streamManager.removeStream(streamID)
	}

	c.streamManager.addStream(streamID, streamCtx)

	// Start streaming in goroutine
	go c.handleStreamingResponse(ctx, req, provider, streamCtx)

	return &AIResponse{
		Provider:    provider.Name(),
		StreamID:    streamID,
		HasThoughts: req.EnableThoughts,
	}, nil
}

// Handle real-time streaming response
func (c *Core) handleStreamingResponse(ctx context.Context, req *AIRequest, provider providers.AIProvider, streamCtx *StreamContext) {
	defer close(streamCtx.Channel)

	options := &providers.ChatOptions{
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		SystemPrompt: req.SystemPrompt,
		Stream:       true,
	}

	// Check if provider supports streaming
	if streamer, ok := provider.(providers.StreamingProvider); ok {
		streamer.ChatStream(ctx, req.Messages, options, func(chunk providers.StreamChunk) {
			streamChunk := StreamChunk{
				Content:   chunk.Content,
				Type:      chunk.Type,
				IsThought: chunk.IsThought,
				Delta:     chunk.Delta,
				Done:      chunk.Done,
				Error:     chunk.Error,
			}

			select {
			case streamCtx.Channel <- streamChunk:
			case <-ctx.Done():
				return
			}
		})
	} else {
		// Fallback: simulate streaming for non-streaming providers
		response, err := provider.Chat(ctx, req.Messages, options)
		if err != nil {
			streamCtx.Channel <- StreamChunk{Error: err, Done: true}
			return
		}

		// Simulate word-by-word streaming
		words := strings.Fields(response.Content)
		for i, word := range words {
			select {
			case streamCtx.Channel <- StreamChunk{
				Content: word + " ",
				Type:    "text",
				Delta:   true,
				Done:    i == len(words)-1,
			}:
				time.Sleep(time.Millisecond * 50) // Simulate typing
			case <-ctx.Done():
				return
			}
		}
	}
}

// Get streaming channel for real-time updates
func (c *Core) GetStreamChannel(streamID string) (<-chan StreamChunk, bool) {
	return c.streamManager.getStreamChannel(streamID)
}

// Try fallback providers on failure
func (c *Core) tryFallbackProvider(ctx context.Context, req *AIRequest, originalErr error) (*AIResponse, error) {
	c.mu.RLock()
	fallbacks := c.config.FallbackProviders
	c.mu.RUnlock()

	for _, providerName := range fallbacks {
		if provider, exists := c.providers[providerName]; exists && provider.Name() != req.Provider {
			if response, err := c.processStandardRequest(ctx, req, provider, time.Now()); err == nil {
				response.Metadata["fallback_from"] = req.Provider
				response.Metadata["original_error"] = originalErr.Error()
				return response, nil
			}
		}
	}

	return nil, fmt.Errorf("all providers failed: %w", originalErr)
}

// Select best provider for request
func (c *Core) selectProvider(req *AIRequest) providers.AIProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Use specified provider if available
	if req.Provider != "" {
		if provider, exists := c.providers[req.Provider]; exists {
			return provider
		}
	}

	// Use active provider
	if provider, exists := c.providers[c.activeModel]; exists {
		return provider
	}

	// Use any available provider
	for _, provider := range c.providers {
		return provider
	}

	return nil
}

// Switch active provider
func (c *Core) SwitchProvider(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.providers[name]; !exists {
		return fmt.Errorf("provider %s not available", name)
	}

	c.activeModel = name
	return nil
}

// Get available providers
func (c *Core) GetAvailableProviders() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var providers []string
	for name := range c.providers {
		providers = append(providers, name)
	}
	return providers
}

// Get current active provider
func (c *Core) GetActiveProvider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activeModel
}

// Create stream manager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		activeStreams: make(map[string]*StreamContext),
	}
}

// Add new stream
func (sm *StreamManager) addStream(id string, ctx *StreamContext) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.activeStreams[id] = ctx
}

// Remove stream
func (sm *StreamManager) removeStream(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.activeStreams, id)
}

// Get stream channel
func (sm *StreamManager) getStreamChannel(id string) (<-chan StreamChunk, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if ctx, exists := sm.activeStreams[id]; exists {
		return ctx.Channel, true
	}
	return nil, false
}
