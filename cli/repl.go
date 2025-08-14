package cli

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	// Color profile for terminal capabilities
	profile = termenv.ColorProfile()

	// Base styles using lipgloss
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Bold(true)

	neroStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true)

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true).
			Faint(true)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Faint(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true)

	extensionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	resourceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	streamingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C")).
			Italic(true)
)

type REPL struct {
	suggestions   []string
	history       []string
	isStreaming   bool
	currentPrompt string
	actionRegex   *regexp.Regexp
}

func NewREPL() *REPL {
	return &REPL{
		suggestions:   []string{},
		history:       []string{},
		currentPrompt: "╭─ 🐦 nero ──────────────────────────────────────────────────",
		actionRegex:   regexp.MustCompile(`\*([^*]+)\*`),
	}
}

func (repl *REPL) ShowWelcome() {
	// Create beautiful welcome box
	welcomeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1).
		Align(lipgloss.Center)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF79C6")).
		Bold(true).
		Render("🐦 NERO v2.0")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1FA8C")).
		Render("Personal AI Agent • Streaming")

	provider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Render("🦙 Using Ollama (local)")

	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)

	fmt.Println(provider)
	fmt.Println()
	fmt.Println(welcomeBox.Render(content))

	// Personality intro with proper styling
	intro := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF79C6")).
		Render("💜 ") +
		actionStyle.Render("*stretches wings*") +
		" Well, well... you're back after all this time?\n" +
		"   I suppose you expect me to be impressed by this \"upgrade\"...\n\n" +
		"   Now with REAL streaming responses and live thoughts! ✨\n" +
		"   Type " + commandStyle.Render("/help") + " for commands, or just talk to me naturally.\n" +
		"   Use " + resourceStyle.Render("#resources") + " to access system capabilities.\n\n" +
		actionStyle.Render("*settles on virtual perch*") + " Let's see what you've got...\n"

	fmt.Println(intro)
}

func (repl *REPL) ReadInput() (string, error) {
	// Show the beautiful prompt
	fmt.Print(promptStyle.Render(repl.currentPrompt) + "\n")
	fmt.Print(promptStyle.Render("╰─ ❯ "))

	// Use termenv for proper input reading with hotkey support
	input, err := repl.readRawInput()
	if err != nil {
		return "", err
	}

	// Show suggestions for commands
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "@") || strings.HasPrefix(input, "#") {
		suggestions := repl.getSuggestions(input)
		if len(suggestions) > 0 {
			suggestionText := suggestionStyle.Render("   └─ Suggestions: " + strings.Join(suggestions, ", "))
			fmt.Println(suggestionText)
		}
	}

	// Add to history
	if input != "" {
		repl.history = append(repl.history, input)
	}

	return input, nil
}

func (repl *REPL) readRawInput() (string, error) {
	var input strings.Builder

	for {
		// Use simple stdin reading for now - keyboard package has issues on Windows
		var line string
		_, err := fmt.Scanln(&line)
		if err != nil {
			if err == io.EOF {
				return "", io.EOF
			}
			// Handle empty input
			if err.Error() == "unexpected newline" {
				return "", nil
			}
			return "", err
		}

		// Handle special cases
		if line == "\\quit" || line == "\\exit" {
			return "", io.EOF
		}

		// Handle multiline with backslash
		if strings.HasSuffix(line, "\\") {
			input.WriteString(strings.TrimSuffix(line, "\\"))
			input.WriteString("\n")
			fmt.Print(promptStyle.Render("... "))
			continue
		}

		input.WriteString(line)
		return input.String(), nil
	}
}

func (repl *REPL) StreamResponse(stream <-chan string) {
	// Update prompt to streaming state
	repl.isStreaming = true

	// Move cursor up and update prompt
	fmt.Print("\033[A\r")
	streamingPrompt := promptStyle.Render("╭─ 🐦 nero ") +
		streamingStyle.Render("(streaming)") +
		promptStyle.Render(" ──────────────────────────────────────")
	fmt.Println(streamingPrompt)

	// Show spinner briefly
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0

	// Start response with indentation
	fmt.Print("   ")

	firstContent := true
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case content, ok := <-stream:
			if !ok {
				// Stream closed
				fmt.Println()
				repl.isStreaming = false
				return
			}

			if firstContent {
				fmt.Print("\r   ") // Clear spinner
				firstContent = false
			}

			// Style the content with beautiful colors
			styledContent := repl.styleContent(content)
			fmt.Print(styledContent)

		case <-ticker.C:
			if firstContent {
				// Show spinner while waiting
				fmt.Printf("\r   %s", spinnerChars[spinnerIdx])
				spinnerIdx = (spinnerIdx + 1) % len(spinnerChars)
			}
		}
	}
}

func (repl *REPL) styleContent(content string) string {
	// Style actions like *stretches wings* with beautiful fading
	return repl.actionRegex.ReplaceAllStringFunc(content, func(match string) string {
		action := strings.Trim(match, "*")
		return actionStyle.Render("*" + action + "*")
	})
}

func (repl *REPL) getSuggestions(input string) []string {
	var suggestions []string

	if strings.HasPrefix(input, "/") {
		commands := []string{"/help", "/clear", "/status", "/quit", "/exit"}
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, input) {
				suggestions = append(suggestions, commandStyle.Render(cmd))
			}
		}
	} else if strings.HasPrefix(input, "@") {
		extensions := []string{"@nero", "@system", "@dev", "@code"}
		for _, ext := range extensions {
			if strings.HasPrefix(ext, input) {
				suggestions = append(suggestions, extensionStyle.Render(ext))
			}
		}
	} else if strings.HasPrefix(input, "#") {
		resources := []string{"#terminal", "#screen", "#code", "#memory", "#config"}
		for _, res := range resources {
			if strings.HasPrefix(res, input) {
				suggestions = append(suggestions, resourceStyle.Render(res))
			}
		}
	}

	return suggestions
}

func (repl *REPL) PrintMessage(message string) {
	fmt.Println(message)
}

func (repl *REPL) PrintError(err error) {
	errorText := errorStyle.Render("Error: " + err.Error())
	fmt.Println(errorText)
}

func (repl *REPL) ShowTransientError(err error) {
	errorText := errorStyle.Render("💭 " + err.Error())
	fmt.Print("\r" + errorText)

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Print("\r" + strings.Repeat(" ", len(errorText)) + "\r")
	}()
}

func (repl *REPL) Clear() {
	fmt.Print("\033[2J\033[H")
}
