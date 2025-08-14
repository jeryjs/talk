package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type CursorAnimation struct {
	frames  []string
	current int
	active  bool
	mutex   sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewCursorAnimation() *CursorAnimation {
	return &CursorAnimation{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (ca *CursorAnimation) Start(ctx context.Context) {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()

	if ca.active {
		return
	}

	ca.ctx, ca.cancel = context.WithCancel(ctx)
	ca.active = true
	ca.current = 0

	go ca.animate()
}

func (ca *CursorAnimation) Stop() {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()

	if !ca.active {
		return
	}

	ca.active = false
	ca.cancel()

	// Clear spinner
	fmt.Print("\r \r")
}

func (ca *CursorAnimation) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ca.ctx.Done():
			return
		case <-ticker.C:
			ca.mutex.Lock()
			if ca.active {
				fmt.Printf("\r%s", ca.frames[ca.current])
				ca.current = (ca.current + 1) % len(ca.frames)
			}
			ca.mutex.Unlock()
		}
	}
}

type StreamRenderer struct {
	buffer      strings.Builder
	lineBuffer  strings.Builder
	animation   *CursorAnimation
	mutex       sync.Mutex
	isStreaming bool
}

func NewStreamRenderer() *StreamRenderer {
	return &StreamRenderer{
		animation: NewCursorAnimation(),
	}
}

func (sr *StreamRenderer) StartStreaming(ctx context.Context) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	sr.isStreaming = true
	sr.buffer.Reset()
	sr.lineBuffer.Reset()
	sr.animation.Start(ctx)
}

func (sr *StreamRenderer) Write(content string) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if !sr.isStreaming {
		return
	}

	// Add to buffers
	sr.buffer.WriteString(content)

	// Just print the content directly without complex line buffering
	fmt.Print(content)
}

func (sr *StreamRenderer) FinishStreaming() {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	sr.animation.Stop()

	// Ensure we end with a newline
	fmt.Println()

	sr.isStreaming = false
}

func (sr *StreamRenderer) GetComplete() string {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	return sr.buffer.String()
}

type AdvancedREPL struct {
	renderer       *StreamRenderer
	highlighter    *SyntaxHighlighter
	completer      *AutoCompleter
	history        []string
	historyPos     int
	prompt         string
	multiline      bool
	reader         *bufio.Reader
	errorAnimation *CursorAnimation
}

func NewAdvancedREPL() *AdvancedREPL {
	return &AdvancedREPL{
		renderer:       NewStreamRenderer(),
		highlighter:    NewSyntaxHighlighter(),
		completer:      NewAutoCompleter(),
		history:        make([]string, 0),
		historyPos:     -1,
		prompt:         "nero> ",
		reader:         bufio.NewReader(os.Stdin),
		errorAnimation: NewCursorAnimation(),
	}
}

func (repl *AdvancedREPL) SetPrompt(prompt string) {
	repl.prompt = prompt
}

func (repl *AdvancedREPL) ReadInput() (string, error) {
	// Beautiful fancy prompt
	fmt.Printf("╭─ 🐦 nero ──────────────────────────────────────────────────\n")
	fmt.Printf("╰─ ❯ ")

	// Read a line of input (this properly blocks)
	line, err := repl.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input := strings.TrimSpace(line)

	// Show live suggestions while typing
	if input != "" {
		suggestions := repl.completer.GetSuggestions(input, len(input))
		if len(suggestions) > 0 {
			// Show suggestions below the input
			fmt.Printf("%s   └─ Suggestions: %s%s\n", Gray, strings.Join(suggestions, ", "), Reset)
		}
	}

	// Add to history and completer
	if input != "" {
		repl.history = append(repl.history, input)
		repl.historyPos = len(repl.history)
		repl.completer.AddToHistory(input)
	}

	return input, nil
}

func (repl *AdvancedREPL) StreamResponse(stream <-chan string) {
	// Show streaming prompt
	fmt.Printf("╰─ 🐦 nero (streaming)\n   ")

	var fullResponse strings.Builder

	for content := range stream {
		fmt.Print(content)
		fullResponse.WriteString(content)
	}

	// End with newline and close the response
	fmt.Println()
}

func (repl *AdvancedREPL) PrintMessage(message string) {
	fmt.Println(message)
}

func (repl *AdvancedREPL) PrintError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func (repl *AdvancedREPL) Clear() {
	fmt.Print("\033[2J\033[H")
}

func (repl *AdvancedREPL) ShowTransientError(err error) {
	// Show error with animation, then clear after 3 seconds
	ctx := context.Background()
	repl.errorAnimation.Start(ctx)

	fmt.Printf("\r%sError: %v%s", Red, err, Reset)

	go func() {
		time.Sleep(3 * time.Second)
		repl.errorAnimation.Stop()
		// Clear the line
		fmt.Print("\r" + strings.Repeat(" ", 50) + "\r")
	}()
}
