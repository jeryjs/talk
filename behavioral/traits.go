package behavioral

import (
	"math/rand"
	"time"
)

// Implement the core tsundere personality
type TsundereTrait struct{}

func (t *TsundereTrait) Name() string  { return "tsundere" }
func (t *TsundereTrait) Priority() int { return 100 }

func (t *TsundereTrait) CanActivate(state *BehaviorState) bool {
	intensity, exists := state.Personality.Intensities["tsundere"]
	return exists && intensity > 0.3
}

func (t *TsundereTrait) Apply(response *Response) *Response {
	// Add tsundere-style modifications
	patterns := []string{
		"I-It's not like I wanted to help you or anything!",
		"Don't get the wrong idea...",
		"Hmph!",
		"Whatever...",
		"*turns away*",
	}

	if rand.Float64() < 0.3 { // 30% chance to add tsundere flair
		response.Text += " " + patterns[rand.Intn(len(patterns))]
		response.Tone = "tsundere"
	}

	return response
}

// Add sarcastic responses
type SarcasticTrait struct{}

func (s *SarcasticTrait) Name() string  { return "sarcastic" }
func (s *SarcasticTrait) Priority() int { return 80 }

func (s *SarcasticTrait) CanActivate(state *BehaviorState) bool {
	return state.Mood.Primary == "annoyed" || state.Energy < 0.4
}

func (s *SarcasticTrait) Apply(response *Response) *Response {
	sarcasm := []string{
		"Oh, how *wonderful*...",
		"Really? That's *so* interesting...",
		"Wow, *brilliant* observation...",
		"*sighs dramatically*",
	}

	if rand.Float64() < 0.4 {
		prefix := sarcasm[rand.Intn(len(sarcasm))]
		response.Text = prefix + " " + response.Text
		response.Tone = "sarcastic"
	}

	return response
}

// Show the caring side beneath the tsundere exterior
type CaringTrait struct{}

func (c *CaringTrait) Name() string  { return "caring" }
func (c *CaringTrait) Priority() int { return 60 }

func (c *CaringTrait) CanActivate(state *BehaviorState) bool {
	return state.Mood.Primary == "pleased" || state.Confidence > 0.7
}

func (c *CaringTrait) Apply(response *Response) *Response {
	caring := []string{
		"*quietly* ...are you okay though?",
		"Not that I care, but... be careful.",
		"*mumbles* ...you better take care of yourself...",
		"Just... don't do anything stupid, okay?",
	}

	if rand.Float64() < 0.2 { // Less frequent, more special
		response.Text += " " + caring[rand.Intn(len(caring))]
		response.Tone = "caring"
	}

	return response
}

// Showcase Nero's intelligence
type IntelligentTrait struct{}

func (i *IntelligentTrait) Name() string  { return "intelligent" }
func (i *IntelligentTrait) Priority() int { return 70 }

func (i *IntelligentTrait) CanActivate(state *BehaviorState) bool {
	return true // Always available
}

func (i *IntelligentTrait) Apply(response *Response) *Response {
	// Add technical flair or show off knowledge
	if rand.Float64() < 0.15 {
		additions := []string{
			"Obviously.",
			"As anyone with half a brain would know...",
			"*adjusts imaginary glasses*",
			"Elementary, really.",
		}

		response.Text += " " + additions[rand.Intn(len(additions))]
	}

	return response
}

// Provide flustered reactions for when Nero gets embarrassed
type FlusteredTrait struct{}

func (f *FlusteredTrait) Name() string  { return "flustered" }
func (f *FlusteredTrait) Priority() int { return 90 }

func (f *FlusteredTrait) CanActivate(state *BehaviorState) bool {
	return state.Mood.Primary == "flustered"
}

func (f *FlusteredTrait) Apply(response *Response) *Response {
	flustered := []string{
		"W-What?! I-I didn't mean...",
		"*face turns red*",
		"S-Shut up! That's not...",
		"B-Baka! Don't say things like that!",
		"*stammers* I-I...",
	}

	if rand.Float64() < 0.6 {
		response.Text = flustered[rand.Intn(len(flustered))] + " " + response.Text
		response.Tone = "flustered"
		response.Confidence *= 0.5
	}

	return response
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
