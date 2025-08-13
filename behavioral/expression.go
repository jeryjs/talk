package behavioral

import (
	"math/rand"
)

// Generate visual expressions for Nero
type ExpressionSystem struct {
	expressions map[string][]string
}

// Learn from interactions to modify behavior
type AdaptationSystem struct {
	patterns map[string]float64
	history  []Interaction
}

// Represent a recorded interaction for learning
type Interaction struct {
	Input    string
	Response string
	Feedback string
	Success  bool
}

// Create a new expression system
func NewExpressionSystem() *ExpressionSystem {
	return &ExpressionSystem{
		expressions: map[string][]string{
			"neutral":   {"😐", "😑", "🙄"},
			"happy":     {"😊", "😄", "😸", "✨"},
			"annoyed":   {"😤", "💢", "😠", "🔥"},
			"flustered": {"😳", "😖", "💜", "😵‍💫"},
			"pleased":   {"😌", "😏", "💫", "✨"},
			"sarcastic": {"🙄", "😒", "💅", "🎭"},
			"caring":    {"💜", "😊", "🥺", "💕"},
			"tired":     {"😴", "😪", "💤", "😑"},
			"excited":   {"✨", "🔥", "💫", "⚡"},
			"confident": {"😎", "💪", "✨", "👑"},
		},
	}
}

// Create a new adaptation system
func NewAdaptationSystem() *AdaptationSystem {
	return &AdaptationSystem{
		patterns: make(map[string]float64),
		history:  make([]Interaction, 0),
	}
}

// Generate an appropriate expression based on state and response
func (e *ExpressionSystem) Generate(state *BehaviorState, response *Response) string {
	var expressionPool []string

	// Use expressions based on mood and tone
	mood := state.Mood.Primary
	if expressions, exists := e.expressions[mood]; exists {
		expressionPool = append(expressionPool, expressions...)
	}

	// Add expressions for specific tone
	if response.Tone != "neutral" {
		if expressions, exists := e.expressions[response.Tone]; exists {
			expressionPool = append(expressionPool, expressions...)
		}
	}

	// Add expressions based on confidence
	if response.Confidence > 0.8 {
		expressionPool = append(expressionPool, e.expressions["confident"]...)
	} else if response.Confidence < 0.3 {
		expressionPool = append(expressionPool, e.expressions["flustered"]...)
	}

	if len(expressionPool) == 0 {
		expressionPool = e.expressions["neutral"]
	}

	return expressionPool[rand.Intn(len(expressionPool))]
}

// Add a new expression to a category
func (e *ExpressionSystem) AddExpression(category string, expression string) {
	if e.expressions[category] == nil {
		e.expressions[category] = make([]string, 0)
	}
	e.expressions[category] = append(e.expressions[category], expression)
}

// Add an interaction to the learning history
func (a *AdaptationSystem) Record(interaction Interaction) {
	a.history = append(a.history, interaction)

	// Keep only recent history (last 1000 interactions)
	if len(a.history) > 1000 {
		a.history = a.history[len(a.history)-1000:]
	}

	// Update patterns based on success
	if interaction.Success {
		a.patterns[interaction.Input] += 0.1
	} else {
		a.patterns[interaction.Input] -= 0.05
	}
}

// Return learned adaptation for given input
func (a *AdaptationSystem) GetAdaptation(input string) float64 {
	return a.patterns[input]
}

// Suggest behavioral modifications based on learning
func (a *AdaptationSystem) SuggestBehavior(input string) map[string]float64 {
	suggestions := make(map[string]float64)

	// Analyze successful interactions with similar inputs
	for _, interaction := range a.history {
		if interaction.Success && similarity(input, interaction.Input) > 0.7 {
			// Extract successful patterns
			if len(interaction.Response) > 0 {
				suggestions["confidence"] += 0.1
			}
		}
	}

	return suggestions
}

// Calculate similarity between two strings (simple implementation)
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	// Use simple length-based similarity
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	if maxLen == 0 {
		return 1.0
	}

	return float64(minLen) / float64(maxLen)
}
