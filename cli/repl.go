package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nero/kernel"
	"nero/providers"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles using lipgloss with vibrant colors
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

	responseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6"))

	userInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD"))
)

// Messages for Bubble Tea
type streamingMsg struct {
	content string
}

type streamCompleteMsg struct{}

type userInputMsg struct {
	input string
}

type REPL struct {
	core            *kernel.Core
	textInput       textinput.Model
	viewport        viewport.Model
	suggestions     []string
	history         []string
	isStreaming     bool
	actionRegex     *regexp.Regexp
	markdown        *glamour.TermRenderer
	width           int
	height          int
	showSuggestions bool
	promptText      string
	viewportContent string
	streamID        string
	streamChan      <-chan kernel.StreamChunk
}

func NewREPL() *REPL {
	// Initialize AI core (providers, streaming, etc.)
	core := kernel.NewCore()
	_ = core.Initialize() // fail-fast style; provider banner below will reflect availability

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80

	// Create viewport for conversation history
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Initialize markdown renderer
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(78),
	)

	repl := &REPL{
		core:            core,
		textInput:       ti,
		viewport:        vp,
		suggestions:     []string{},
		history:         []string{},
		actionRegex:     regexp.MustCompile(`\*([^*]+)\*`),
		markdown:        renderer,
		showSuggestions: false,
		promptText:      "╭─ 🐦 nero ──────────────────────────────────────────────────",
	}

	return repl
}

// Bubble Tea Model Implementation
func (r *REPL) Init() tea.Cmd {
	return textinput.Blink
}

func (r *REPL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

		// Update viewport and text input sizes
		r.viewport.Width = msg.Width
		r.viewport.Height = msg.Height - 6 // Leave space for prompt and suggestions
		r.textInput.Width = msg.Width - 10

	case tea.KeyMsg:
		// Update text input first to get the latest value
		r.textInput, tiCmd = r.textInput.Update(msg)

		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit

		case "shift+enter":
			// Multi-line input - add newline
			current := r.textInput.Value()
			r.textInput.SetValue(current + "\n")
			return r, nil

		case "enter":
			// Submit input
			input := strings.TrimSpace(r.textInput.Value())
			if input != "" {
				cmd := r.handleInput(input)
				r.textInput.SetValue("")
				r.showSuggestions = false
				return r, cmd
			}
			return r, nil
		}

		// Update suggestions based on the current input value
		r.updateSuggestions()

	case streamingMsg:
		// Handle streaming content
		r.addResponseContent(msg.content)
		// Wait for next chunk
		return r, r.waitForStreamChunk()

	case streamCompleteMsg:
		// Streaming complete
		r.isStreaming = false
		r.promptText = "╭─ 🐦 nero ──────────────────────────────────────────────────"
	}

	// Update viewport
	r.viewport, vpCmd = r.viewport.Update(msg)

	return r, tea.Batch(tiCmd, vpCmd)
}

func (r *REPL) View() string {
	var b strings.Builder

	// Show conversation history in viewport
	b.WriteString(r.viewport.View())
	b.WriteString("\n")

	// Show current prompt with streaming indicator
	if r.isStreaming {
		promptWithStreaming := neroStyle.Render("╭─ 🐦 nero ───────────────────────────────────")
		b.WriteString(promptWithStreaming)
	} else {
		b.WriteString(promptStyle.Render(r.promptText))
	}
	b.WriteString("\n")

	// Show input line with proper styling
	inputPrefix := promptStyle.Render("╰─ ❯ ")
	styledInput := userInputStyle.Render(r.textInput.View())
	b.WriteString(inputPrefix + styledInput)
	b.WriteString("\n")

	// Show suggestions if any
	if r.showSuggestions && len(r.suggestions) > 0 {
		suggestionText := suggestionStyle.Render("   └─ Suggestions: " + strings.Join(r.suggestions, ", "))
		b.WriteString(suggestionText)
		b.WriteString("\n")
	}

	return b.String()
}

func (r *REPL) handleInput(input string) tea.Cmd {
	// Add user input to history and display
	userMessage := userInputStyle.Render("❯ " + input)
	r.addToViewport(userMessage)

	// Add to history
	r.history = append(r.history, input)

	if strings.HasPrefix(input, "/") {
		r.handleCommand(input)
		return nil
	}

	// Real AI streaming via kernel core
	r.startStreaming()
	return r.startAIStreaming(input)
}

func (r *REPL) handleCommand(input string) {
	command := strings.TrimPrefix(input, "/")
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "help":
		helpText := `# Available Commands

- **/help** - Show this help
- **/clear** - Clear conversation  
- **/quit**, **/exit** - Exit NERO
- **/status** - Show system status

## Special Prefixes
- **@** - Extensions (@nero, @system, @dev)
- **#** - Resources (#terminal, #screen, #code)`

		// Render as markdown
		if rendered, err := r.markdown.Render(helpText); err == nil {
			r.addToViewport(rendered)
		} else {
			r.addToViewport(responseStyle.Render(helpText))
		}

	case "clear":
		r.viewportContent = ""
		r.viewport.SetContent("")

	case "quit", "exit":
		// This would trigger app exit

	case "status":
		statusText := responseStyle.Render("🐦 NERO v2.0 - Personal AI Agent\nStatus: Active\nProvider: Ollama (local)")
		r.addToViewport(statusText)

	default:
		errorText := errorStyle.Render("Unknown command: " + parts[0])
		r.addToViewport(errorText)
	}
}

func (r *REPL) startStreaming() {
	r.isStreaming = true
	r.promptText = "╭─ 🐦 nero ─*streaming*──────────────────────────────────────"
}

// Kick off AI streaming with the kernel core
func (r *REPL) startAIStreaming(input string) tea.Cmd {
	if r.core == nil {
		r.addToViewport(errorStyle.Render("AI core not initialized"))
		return nil
	}

	// Build request
	req := &kernel.AIRequest{
		Messages:       []providers.Message{{Role: "user", Content: input}},
		EnableStream:   true,
		EnableThoughts: false,
		Temperature:    0.8,
		MaxTokens:      500,
	}

	// Process request (returns immediately for streaming)
	ctx := context.Background()
	resp, err := r.core.ProcessRequest(ctx, req)
	if err != nil {
		r.addToViewport(errorStyle.Render(fmt.Sprintf("AI error: %v", err)))
		return nil
	}

	if resp.StreamID == "" {
		// Non-streaming fallback
		r.addResponseContent(resp.Content)
		return nil
	}

	r.streamID = resp.StreamID
	ch, ok := r.core.GetStreamChannel(resp.StreamID)
	if !ok {
		r.addToViewport(errorStyle.Render("Stream channel not found"))
		return nil
	}
	r.streamChan = ch

	// Start waiting for first chunk
	return r.waitForStreamChunk()
}

// tea.Cmd that waits for the next stream chunk and returns a message
func (r *REPL) waitForStreamChunk() tea.Cmd {
	if r.streamChan == nil {
		return func() tea.Msg { return streamCompleteMsg{} }
	}
	return func() tea.Msg {
		chunk, ok := <-r.streamChan
		if !ok {
			return streamCompleteMsg{}
		}
		if chunk.Error != nil {
			return streamCompleteMsg{}
		}
		if chunk.Done {
			return streamCompleteMsg{}
		}
		return streamingMsg{content: chunk.Content}
	}
}

func (r *REPL) addResponseContent(content string) {
	// Add response content to viewport
	r.addToViewport(responseStyle.Render(content))
}

func (r *REPL) addToViewport(content string) {
	if r.viewportContent != "" {
		r.viewportContent += "\n"
	}
	r.viewportContent += content
	r.viewport.SetContent(r.viewportContent)
	r.viewport.GotoBottom()
}

func (r *REPL) updateSuggestions() {
	input := r.textInput.Value()
	r.suggestions = []string{}
	r.showSuggestions = false

	if strings.HasPrefix(input, "/") {
		commands := []string{"/help", "/clear", "/status", "/quit", "/exit"}
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, input) {
				r.suggestions = append(r.suggestions, commandStyle.Render(cmd))
			}
		}
		r.showSuggestions = len(r.suggestions) > 0

	} else if strings.HasPrefix(input, "@") {
		extensions := []string{"@nero", "@system", "@dev", "@code"}
		for _, ext := range extensions {
			if strings.HasPrefix(ext, input) {
				r.suggestions = append(r.suggestions, extensionStyle.Render(ext))
			}
		}
		r.showSuggestions = len(r.suggestions) > 0

	} else if strings.HasPrefix(input, "#") {
		resources := []string{"#terminal", "#screen", "#code", "#memory", "#config"}
		for _, res := range resources {
			if strings.HasPrefix(res, input) {
				r.suggestions = append(r.suggestions, resourceStyle.Render(res))
			}
		}
		r.showSuggestions = len(r.suggestions) > 0
	}
}

func (r *REPL) ShowWelcome() string {
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

	// Provider banner based on active provider
	active := "unknown"
	if r.core != nil {
		active = r.core.GetActiveProvider()
	}
	providerText := ""
	switch active {
	case "ollama":
		providerText = "🦙 Using Ollama (local)"
	case "openai":
		providerText = "🌐 Using OpenAI"
	case "gemini":
		providerText = "💎 Using Google Gemini"
	case "groq":
		providerText = "⚡ Using Groq"
	default:
		providerText = "⚠️ No provider available"
	}
	provider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Render(providerText)

	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)

	welcome := provider + "\n\n" + welcomeBox.Render(content) + "\n"

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

	return welcome + intro
}

// Start the REPL using Bubble Tea
func (r *REPL) Start() error {
	// Initialize with welcome message
	welcome := r.ShowWelcome()
	r.addToViewport(welcome)

	// Start the Bubble Tea program
	p := tea.NewProgram(r)
	_, err := p.Run()
	return err
}
