package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Handle real-time thought visualization in the CLI
type ThoughtRenderer struct {
	maxWidth      int
	thoughtBuffer []string
	isThinking    bool
	startTime     time.Time
	fadeWords     int // Number of words to keep visible
}

// Create a new thought renderer
func NewThoughtRenderer() *ThoughtRenderer {
	return &ThoughtRenderer{
		maxWidth:  80,
		fadeWords: 10,
	}
}

// Start thinking animation
func (tr *ThoughtRenderer) StartThinking() {
	tr.isThinking = true
	tr.startTime = time.Now()
	tr.thoughtBuffer = make([]string, 0)
}

// Add thought text with live rendering
func (tr *ThoughtRenderer) AddThought(text string) {
	if !tr.isThinking {
		return
	}

	words := strings.Fields(text)
	tr.thoughtBuffer = append(tr.thoughtBuffer, words...)

	// Render live thoughts with fade effect
	tr.renderLiveThoughts()
}

// Render live thoughts with fade animation
func (tr *ThoughtRenderer) renderLiveThoughts() {
	// Move cursor to beginning of thought line
	fmt.Print("\r")

	// Clear the line
	fmt.Print("\033[K")

	// Render the prompt prefix
	color.New(color.FgHiBlack).Printf("╰─ ")
	color.New(color.FgMagenta, color.Faint).Printf("💭 ")

	// Calculate visible words with fade effect
	totalWords := len(tr.thoughtBuffer)
	startIdx := 0
	if totalWords > tr.fadeWords {
		startIdx = totalWords - tr.fadeWords
	}

	// Render faded thoughts
	for i := startIdx; i < totalWords; i++ {
		word := tr.thoughtBuffer[i]

		// Calculate fade intensity
		pos := float64(i - startIdx)
		intensity := (pos/float64(tr.fadeWords))*0.7 + 0.3 // 30-100% intensity

		if intensity > 1.0 {
			intensity = 1.0
		}

		// Apply fade color
		if intensity < 0.5 {
			color.New(color.FgHiBlack, color.Faint).Printf("%s ", word)
		} else if intensity < 0.8 {
			color.New(color.FgBlue, color.Faint).Printf("%s ", word)
		} else {
			color.New(color.FgCyan).Printf("%s ", word)
		}
	}

	// Add typing indicator
	color.New(color.FgMagenta, color.Faint).Printf("▋")
}

// Stop thinking and show summary
func (tr *ThoughtRenderer) StopThinking() {
	if !tr.isThinking {
		return
	}

	duration := time.Since(tr.startTime)
	tr.isThinking = false

	// Clear the thinking line
	fmt.Print("\r\033[K")

	// Show thought summary
	color.New(color.FgHiBlack).Printf("╰─ ")
	color.New(color.FgMagenta, color.Faint).Printf("💭 ")
	color.New(color.FgHiBlack, color.Faint).Printf("thought for %v ", duration.Truncate(time.Millisecond*100))
	color.New(color.FgHiBlack, color.Faint, color.Underline).Printf("(view)")
	fmt.Println()

	// Store thoughts for later viewing
	tr.storeThoughts()
}

// Store thoughts for later retrieval
func (tr *ThoughtRenderer) storeThoughts() {
	// TODO: Store in memory provider for /thoughts command
}

// View stored thoughts
func (tr *ThoughtRenderer) ViewLastThoughts() string {
	if len(tr.thoughtBuffer) == 0 {
		return "No recent thoughts available"
	}

	var result strings.Builder
	result.WriteString("💭 Recent thoughts:\n\n")

	// Group words into lines for better readability
	words := tr.thoughtBuffer
	line := ""
	for _, word := range words {
		if len(line)+len(word)+1 > 70 {
			result.WriteString("   " + line + "\n")
			line = word
		} else {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		}
	}
	if line != "" {
		result.WriteString("   " + line + "\n")
	}

	return result.String()
}

// Handle streaming response with real-time updates
type StreamingRenderer struct {
	thoughtRenderer *ThoughtRenderer
	responseBuffer  strings.Builder
	isStreaming     bool
}

// Create a new streaming renderer
func NewStreamingRenderer() *StreamingRenderer {
	return &StreamingRenderer{
		thoughtRenderer: NewThoughtRenderer(),
	}
}

// Start streaming response
func (sr *StreamingRenderer) StartStreaming() {
	sr.isStreaming = true
	sr.responseBuffer.Reset()

	// Show streaming indicator
	color.New(color.FgHiBlack).Printf("╰─ ")
	color.New(color.FgMagenta).Printf("🐦 nero ")
	color.New(color.FgHiBlack, color.Faint).Printf("(streaming)")
	fmt.Println()
	fmt.Print("   ")
}

// Add streaming content
func (sr *StreamingRenderer) AddContent(content string, isThought bool) {
	if !sr.isStreaming {
		return
	}

	if isThought {
		sr.thoughtRenderer.AddThought(content)
	} else {
		// Add to response and display
		sr.responseBuffer.WriteString(content)
		color.New(color.FgMagenta).Print(content)
	}
}

// Add streaming thought
func (sr *StreamingRenderer) AddThought(thought string) {
	if sr.thoughtRenderer.isThinking {
		sr.thoughtRenderer.AddThought(thought)
	} else {
		sr.thoughtRenderer.StartThinking()
		sr.thoughtRenderer.AddThought(thought)
	}
}

// Stop streaming and finalize
func (sr *StreamingRenderer) StopStreaming() {
	if !sr.isStreaming {
		return
	}

	sr.isStreaming = false

	// Stop any active thinking
	if sr.thoughtRenderer.isThinking {
		sr.thoughtRenderer.StopThinking()
	}

	// Finalize response
	fmt.Println()
}

// Get the complete response
func (sr *StreamingRenderer) GetResponse() string {
	return sr.responseBuffer.String()
}

// Show typing animation for regular prompt
func ShowTypingIndicator() {
	indicators := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	for i := 0; i < 10; i++ {
		fmt.Printf("\r%s Thinking...", indicators[i%len(indicators)])
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Print("\r\033[K") // Clear the line
}

// Animate text reveal
func AnimateTextReveal(text string, delay time.Duration) {
	words := strings.Fields(text)
	for _, word := range words {
		fmt.Print(word + " ")
		time.Sleep(delay)
	}
	fmt.Println()
}
