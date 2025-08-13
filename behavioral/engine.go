package behavioral

import (
	"sync"
	"time"
)

// Manage Nero's dynamic behavioral system
type Engine struct {
	currentState *BehaviorState
	traits       map[string]Trait
	expressions  *ExpressionSystem
	adaptation   *AdaptationSystem
	mu           sync.RWMutex
}

// Represent Nero's current behavioral configuration
type BehaviorState struct {
	Personality PersonalityProfile
	Mood        MoodState
	Energy      float64 // 0.0 to 1.0
	Confidence  float64 // 0.0 to 1.0
	Engagement  float64 // 0.0 to 1.0
	LastUpdated time.Time
}

// Define core personality traits
type PersonalityProfile struct {
	Name        string
	Base        string // "tsundere", "cheerful", "professional", etc.
	Traits      []string
	Intensities map[string]float64
	Quirks      []string
}

// Represent current emotional state
type MoodState struct {
	Primary   string  // "happy", "annoyed", "excited", "tired", etc.
	Secondary string  // Optional secondary mood
	Intensity float64 // 0.0 to 1.0
	Duration  time.Duration
}

// Define a behavioral component
type Trait interface {
	Name() string
	Apply(response *Response) *Response
	CanActivate(state *BehaviorState) bool
	Priority() int
}

// Represent Nero's response to be modified by traits
type Response struct {
	Text       string
	Tone       string
	Expression string
	Confidence float64
	Metadata   map[string]interface{}
}

// Create a new behavioral engine
func NewEngine() *Engine {
	engine := &Engine{
		traits:      make(map[string]Trait),
		expressions: NewExpressionSystem(),
		adaptation:  NewAdaptationSystem(),
	}

	// Initialize default personality (Tsundere Nero)
	engine.currentState = &BehaviorState{
		Personality: PersonalityProfile{
			Name:   "Nero",
			Base:   "tsundere",
			Traits: []string{"sarcastic", "caring", "intelligent", "protective"},
			Intensities: map[string]float64{
				"tsundere":    0.8,
				"sarcastic":   0.7,
				"caring":      0.6,
				"intelligent": 0.9,
			},
			Quirks: []string{"denies_caring", "easily_flustered", "weak_to_please"},
		},
		Mood: MoodState{
			Primary:   "neutral",
			Intensity: 0.5,
			Duration:  time.Hour,
		},
		Energy:      0.8,
		Confidence:  0.7,
		Engagement:  0.6,
		LastUpdated: time.Now(),
	}

	// Register default traits
	engine.registerDefaultTraits()

	return engine
}

// Apply current behavioral state to a response
func (e *Engine) ProcessResponse(input string, baseResponse string) *Response {
	e.mu.RLock()
	defer e.mu.RUnlock()

	response := &Response{
		Text:       baseResponse,
		Tone:       "neutral",
		Expression: "neutral",
		Confidence: e.currentState.Confidence,
		Metadata:   make(map[string]interface{}),
	}

	// Apply active traits in priority order
	activeTraits := e.getActiveTraits()
	for _, trait := range activeTraits {
		response = trait.Apply(response)
	}

	// Add expression
	expression := e.expressions.Generate(e.currentState, response)
	if expression != "" {
		response.Expression = expression
		response.Text = expression + " " + response.Text
	}

	// Update state based on interaction
	e.updateState(input, response)

	return response
}

// Change Nero's current mood
func (e *Engine) UpdateMood(mood string, intensity float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.currentState.Mood = MoodState{
		Primary:   mood,
		Intensity: intensity,
		Duration:  time.Hour, // Default duration
	}
	e.currentState.LastUpdated = time.Now()
}

// Return current behavioral state
func (e *Engine) GetState() *BehaviorState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return copy to prevent external modification
	stateCopy := *e.currentState
	return &stateCopy
}

// Add a new behavioral trait
func (e *Engine) RegisterTrait(trait Trait) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.traits[trait.Name()] = trait
}

// Return traits that can activate in current state
func (e *Engine) getActiveTraits() []Trait {
	var active []Trait

	for _, trait := range e.traits {
		if trait.CanActivate(e.currentState) {
			active = append(active, trait)
		}
	}

	// Sort by priority
	for i := 0; i < len(active)-1; i++ {
		for j := i + 1; j < len(active); j++ {
			if active[i].Priority() < active[j].Priority() {
				active[i], active[j] = active[j], active[i]
			}
		}
	}

	return active
}

// Modify behavioral state based on interaction
func (e *Engine) updateState(input string, response *Response) {
	// Simple state updates - can be made more sophisticated
	if contains(input, "please") {
		e.currentState.Mood.Primary = "flustered"
		e.currentState.Confidence *= 0.8
	}

	if contains(input, "thank") {
		e.currentState.Mood.Primary = "pleased"
		e.currentState.Energy += 0.1
	}

	// Decay energy over time
	timeSince := time.Since(e.currentState.LastUpdated)
	if timeSince > time.Minute {
		e.currentState.Energy *= 0.99
	}

	e.currentState.LastUpdated = time.Now()
}

// Set up Nero's base personality traits
func (e *Engine) registerDefaultTraits() {
	e.RegisterTrait(&TsundereTrait{})
	e.RegisterTrait(&SarcasticTrait{})
	e.RegisterTrait(&CaringTrait{})
	e.RegisterTrait(&IntelligentTrait{})
}

// Check if a substring exists in a string (helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
