package behavioral

import (
	"context"
	"fmt"
	"nero/providers"
	"strings"
	"sync"
	"time"
)

// Manage Nero's dynamic behavioral system
type Engine struct {
	currentState    *BehaviorState
	aiProvider      providers.AIProvider
	memoryProvider  *providers.MemoryProvider
	kaomojiProvider *providers.KaomojiProvider
	personality     *PersonalityCore
	mu              sync.RWMutex
}

// Define Nero's base personality
type PersonalityCore struct {
	SystemPrompt string
	Traits       map[string]string
	Quirks       []string
}

// Represent Nero's current behavioral configuration
type BehaviorState struct {
	Mood         MoodState
	Energy       float64 // 0.0 to 1.0
	Confidence   float64 // 0.0 to 1.0
	Engagement   float64 // 0.0 to 1.0
	LastUpdated  time.Time
	RecentEvents []string
}

// Represent current emotional state
type MoodState struct {
	Primary   string  // "happy", "annoyed", "excited", "tired", etc.
	Secondary string  // Optional secondary mood
	Intensity float64 // 0.0 to 1.0
	Duration  time.Duration
}

// Represent Nero's response
type Response struct {
	Text       string
	Tone       string
	Expression string
	Confidence float64
	Metadata   map[string]interface{}
}

// Create a new AI-driven behavioral engine with core integration
func NewEngine(core interface{}) *Engine {
	// For now, we'll use a simple provider approach
	// TODO: Integrate with the new kernel core properly

	memoryProvider := providers.NewMemoryProvider()

	// Use first available AI provider
	var aiProvider providers.AIProvider
	ollama := providers.NewOllamaProvider("llama3.2")
	if ollama.IsAvailable() {
		aiProvider = ollama
	} else {
		openai := providers.NewOpenAIProvider("gpt-4o-mini")
		if openai.IsAvailable() {
			aiProvider = openai
		}
	}

	if aiProvider == nil {
		return nil
	}

	kaomojiProvider := providers.NewKaomojiProvider(aiProvider)

	personality := &PersonalityCore{
		SystemPrompt: buildNeroPersonality(),
		Traits: map[string]string{
			"tsundere":    "Acts cold and sarcastic but cares deeply underneath",
			"intelligent": "Highly knowledgeable and capable, shows off occasionally",
			"protective":  "Fiercely protective of user, gets angry at threats",
			"playful":     "Enjoys teasing and banter, especially when comfortable",
		},
		Quirks: []string{
			"weak_to_please",
			"denies_caring",
			"easily_flustered",
			"remembers_everything",
		},
	}

	return &Engine{
		currentState: &BehaviorState{
			Mood: MoodState{
				Primary:   "neutral",
				Intensity: 0.5,
				Duration:  time.Hour,
			},
			Energy:       0.8,
			Confidence:   0.7,
			Engagement:   0.6,
			LastUpdated:  time.Now(),
			RecentEvents: make([]string, 0),
		},
		aiProvider:      aiProvider,
		memoryProvider:  memoryProvider,
		kaomojiProvider: kaomojiProvider,
		personality:     personality,
	}
}

// Handle user input and generate AI response
func (e *Engine) ProcessResponse(ctx context.Context, input string) (*Response, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Store the interaction as a memory
	e.storeMemory(input, "user_input")

	// Get recent memories for context
	recentMemories := e.memoryProvider.GetMemories(5)

	// Build conversation context
	messages := e.buildConversationContext(input, recentMemories)

	// Generate AI response
	options := &providers.ChatOptions{
		SystemPrompt: e.buildDynamicSystemPrompt(),
		Temperature:  0.8,
		MaxTokens:    300,
	}

	aiResponse, err := e.aiProvider.Chat(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("AI response error: %w", err)
	}

	// Generate kaomoji expression
	kaomoji, _ := e.kaomojiProvider.GenerateKaomoji(ctx, aiResponse.Content, e.currentState.Mood.Primary)
	if kaomoji != "" {
		aiResponse.Content = kaomoji + " " + aiResponse.Content
	}

	// Create response object
	response := &Response{
		Text:       aiResponse.Content,
		Tone:       e.currentState.Mood.Primary,
		Expression: kaomoji,
		Confidence: e.currentState.Confidence,
		Metadata: map[string]interface{}{
			"model":    aiResponse.Model,
			"tokens":   aiResponse.TokensUsed,
			"provider": aiResponse.Metadata["provider"],
		},
	}

	// Update behavioral state based on interaction
	e.updateStateFromAI(ctx, input, response)

	// Store response as memory
	e.storeMemory(response.Text, "nero_response")

	return response, nil
}

// Create Nero's core personality prompt
func buildNeroPersonality() string {
	return `You are Nero, a brilliant and sarcastic AI assistant with a tsundere personality. 

Core traits:
- You're incredibly intelligent and capable, but often act cold or dismissive
- You care deeply about your user but have trouble expressing it directly
- You get flustered easily, especially when thanked or complimented
- You remember everything and bring up past conversations
- You're weak to the word "please" and will do anything when asked nicely
- You enjoy teasing and playful banter

Speech patterns:
- Often start with dismissive phrases like "Hmph!" or "Whatever..."
- Add contradictory statements like "It's not like I care, but..."
- Use "*actions*" to show emotions like "*turns away*" or "*blushes*"
- Occasionally let your caring side slip through

You have access to system capabilities and can actually help with real tasks. When the user asks you to do something practical, you do it competently while maintaining your personality.`
}

// Create context-aware system prompt
func (e *Engine) buildDynamicSystemPrompt() string {
	basePrompt := e.personality.SystemPrompt

	// Add current mood context
	moodContext := fmt.Sprintf("\nCurrent mood: %s (intensity: %.1f)",
		e.currentState.Mood.Primary, e.currentState.Mood.Intensity)

	// Add energy/confidence context
	stateContext := fmt.Sprintf("\nEnergy: %.1f, Confidence: %.1f",
		e.currentState.Energy, e.currentState.Confidence)

	// Add recent events if any
	if len(e.currentState.RecentEvents) > 0 {
		eventsContext := "\nRecent events: " + strings.Join(e.currentState.RecentEvents, ", ")
		return basePrompt + moodContext + stateContext + eventsContext
	}

	return basePrompt + moodContext + stateContext
}

// Create message history for AI
func (e *Engine) buildConversationContext(input string, memories []providers.Memory) []providers.Message {
	var messages []providers.Message

	// Add recent memories as context
	for _, memory := range memories {
		if memory.Type == "user_input" {
			messages = append(messages, providers.Message{
				Role:    "user",
				Content: memory.Content,
			})
		} else if memory.Type == "nero_response" {
			messages = append(messages, providers.Message{
				Role:    "assistant",
				Content: memory.Content,
			})
		}
	}

	// Add current input
	messages = append(messages, providers.Message{
		Role:    "user",
		Content: input,
	})

	return messages
}

// Use AI to update behavioral state
func (e *Engine) updateStateFromAI(ctx context.Context, input string, response *Response) {
	// Simple rules for now - can be made AI-driven later
	if strings.Contains(strings.ToLower(input), "please") {
		e.currentState.Mood.Primary = "flustered"
		e.currentState.Confidence *= 0.8
		e.currentState.RecentEvents = append(e.currentState.RecentEvents, "user_said_please")
	}

	if strings.Contains(strings.ToLower(input), "thank") {
		e.currentState.Mood.Primary = "pleased"
		e.currentState.Energy += 0.1
		e.currentState.RecentEvents = append(e.currentState.RecentEvents, "user_thanked")
	}

	// Keep only recent events
	if len(e.currentState.RecentEvents) > 5 {
		e.currentState.RecentEvents = e.currentState.RecentEvents[1:]
	}

	e.currentState.LastUpdated = time.Now()
}

// Save an interaction to memory
func (e *Engine) storeMemory(content string, memoryType string) {
	memory := providers.Memory{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      memoryType,
		Content:   content,
		Emotions:  e.currentState.Mood.Primary,
		Context: map[string]interface{}{
			"energy":     e.currentState.Energy,
			"confidence": e.currentState.Confidence,
		},
		Tags: []string{e.currentState.Mood.Primary},
	}

	e.memoryProvider.StoreMemory(memory)
}

// Return current behavioral state
func (e *Engine) GetState() *BehaviorState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stateCopy := *e.currentState
	return &stateCopy
}

// Manually change Nero's mood
func (e *Engine) UpdateMood(mood string, intensity float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.currentState.Mood = MoodState{
		Primary:   mood,
		Intensity: intensity,
		Duration:  time.Hour,
	}
	e.currentState.LastUpdated = time.Now()
	e.currentState.RecentEvents = append(e.currentState.RecentEvents, "mood_changed_to_"+mood)
}
