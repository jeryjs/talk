package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nero/internal/agent"
	"nero/internal/config"
	"nero/internal/models"
	"nero/internal/mood"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("235"))

	selectedSuggestionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("62")).
				Bold(true)
)

type model struct {
	viewport           viewport.Model
	textarea           textarea.Model
	agent              *agent.Agent
	ctx                context.Context
	messages           []message
	width              int
	height             int
	ready              bool
	err                error
	streaming          bool
	currentResponse    string
	ctrlCPressed       bool
	ctrlCTime          time.Time
	showSuggestions    bool
	suggestions        []string
	selectedSuggestion int
	renderer           *glamour.TermRenderer
}

type message struct {
	role    string // user, assistant, system
	content string
	raw     bool // if true, don't apply markdown
}

type responseChunkMsg string
type responseDoneMsg struct{}
type errMsg struct{ error }

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (/ for commands)"
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(80, 20)

	// Create markdown renderer
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	cfg := config.Get()
	ag, err := agent.New()
	if err != nil {
		return model{
			viewport: vp,
			textarea: ta,
			err:      err,
			renderer: renderer,
		}
	}

	// Set initial mood
	moodPrompt := mood.GetSystemPrompt(cfg.Mood)
	ag.SetSystemPrompt(moodPrompt)

	welcome := message{
		role:    "system",
		content: fmt.Sprintf("🤖 **Nero** - AI Assistant\n\nModel: **%s** | Mood: **%s**\n\nType `/help` for commands", cfg.Model, cfg.Mood),
		raw:     false,
	}

	return model{
		viewport:           vp,
		textarea:           ta,
		agent:              ag,
		ctx:                context.Background(),
		messages:           []message{welcome},
		ready:              false,
		renderer:           renderer,
		selectedSuggestion: -1,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8 // Leave space for input and status
			m.textarea.SetWidth(msg.Width - 4)
			m.renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(msg.Width-4),
			)
			m.updateViewport()
			m.ready = true
		}

	case tea.KeyMsg:
		// Handle Ctrl+C
		if msg.Type == tea.KeyCtrlC {
			if m.ctrlCPressed && time.Since(m.ctrlCTime) < 3*time.Second {
				return m, tea.Quit
			}
			m.ctrlCPressed = true
			m.ctrlCTime = time.Now()
			return m, tea.Batch(tiCmd, vpCmd)
		} else {
			m.ctrlCPressed = false
		}

		input := m.textarea.Value()

		// Handle command suggestions
		if strings.HasPrefix(input, "/") {
			m.showSuggestions = true
			m.updateSuggestions(input)

			// Navigate suggestions
			if m.showSuggestions && len(m.suggestions) > 0 {
				switch msg.Type {
				case tea.KeyUp:
					if m.selectedSuggestion > 0 {
						m.selectedSuggestion--
						return m, tea.Batch(tiCmd, vpCmd)
					}
				case tea.KeyDown:
					if m.selectedSuggestion < len(m.suggestions)-1 {
						m.selectedSuggestion++
						return m, tea.Batch(tiCmd, vpCmd)
					}
				case tea.KeyTab:
					if m.selectedSuggestion >= 0 && m.selectedSuggestion < len(m.suggestions) {
						m.textarea.SetValue(m.suggestions[m.selectedSuggestion] + " ")
						m.showSuggestions = false
						m.textarea.CursorEnd()
						return m, tea.Batch(tiCmd, vpCmd)
					}
				}
			}
		} else {
			m.showSuggestions = false
		}

		// Handle Enter
		if msg.Type == tea.KeyEnter && !msg.Alt {
			if m.streaming {
				return m, tea.Batch(tiCmd, vpCmd)
			}

			input = strings.TrimSpace(input)
			if input == "" {
				return m, tea.Batch(tiCmd, vpCmd)
			}

			// Handle commands
			if strings.HasPrefix(input, "/") {
				return m.handleCommand(input)
			}

			// Add user message
			m.messages = append(m.messages, message{
				role:    "user",
				content: input,
				raw:     false,
			})
			m.textarea.Reset()
			m.updateViewport()

			// Start streaming response
			m.streaming = true
			m.currentResponse = ""
			return m, m.streamResponse(input)
		}

	case responseChunkMsg:
		m.currentResponse += string(msg)
		m.updateViewport()
		// Continue listening for more chunks
		return m, m.waitForNextChunk()

	case responseDoneMsg:
		m.streaming = false
		if m.currentResponse != "" {
			m.messages = append(m.messages, message{
				role:    "assistant",
				content: m.currentResponse,
				raw:     false,
			})
			m.currentResponse = ""
		}
		m.updateViewport()
		return m, nil

	case errMsg:
		m.streaming = false
		m.messages = append(m.messages, message{
			role:    "system",
			content: fmt.Sprintf("❌ Error: %v", msg.error),
			raw:     true,
		})
		m.updateViewport()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) updateSuggestions(input string) {
	allCommands := []string{
		"/help",
		"/model",
		"/mood helpful",
		"/mood creative",
		"/mood professional",
		"/mood analytical",
		"/mood concise",
		"/clear",
		"/quit",
	}

	m.suggestions = []string{}
	lowerInput := strings.ToLower(input)

	for _, cmd := range allCommands {
		if strings.HasPrefix(strings.ToLower(cmd), lowerInput) {
			m.suggestions = append(m.suggestions, cmd)
		}
	}

	if m.selectedSuggestion >= len(m.suggestions) {
		m.selectedSuggestion = len(m.suggestions) - 1
	}
	if m.selectedSuggestion < 0 && len(m.suggestions) > 0 {
		m.selectedSuggestion = 0
	}
}

func (m *model) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return *m, nil
	}

	m.textarea.Reset()
	m.showSuggestions = false

	switch parts[0] {
	case "/help":
		help := `**Available Commands:**

• **/help** - Show this help
• **/model [provider/model]** - Switch AI model (or open selector)
• **/mood [name]** - Change personality
• **/clear** - Clear conversation
• **/quit** - Exit

**Keyboard Shortcuts:**
• **Enter** - Send message
• **Ctrl+C** twice - Quit (within 3 seconds)
• **Tab** - Auto-complete command
• **↑/↓** - Navigate command suggestions`

		m.messages = append(m.messages, message{
			role:    "system",
			content: help,
			raw:     false,
		})
		m.updateViewport()

	case "/model":
		if len(parts) > 1 {
			// Model specified
			return m.switchModel(parts[1])
		} else {
			// TODO: Open model selector in alternate screen
			m.messages = append(m.messages, message{
				role:    "system",
				content: "**Model selector coming soon!** For now, use: `/model provider/model`\n\nExamples:\n• `/model gq/llama-3.3-70b`\n• `/model g/gemini-2.5-pro`\n• `/model o/gpt-4o`",
				raw:     false,
			})
			m.updateViewport()
		}

	case "/mood":
		if len(parts) < 2 {
			cfg := config.Get()
			m.messages = append(m.messages, message{
				role:    "system",
				content: fmt.Sprintf("**Current mood:** %s", cfg.Mood),
				raw:     false,
			})
		} else {
			newMood := parts[1]
			if moodPrompt := mood.GetSystemPrompt(newMood); moodPrompt != "" {
				m.agent.SetSystemPrompt(moodPrompt)
				moodInfo, _ := mood.Get(newMood)
				m.messages = append(m.messages, message{
					role:    "system",
					content: fmt.Sprintf("✓ **Mood set to:** %s %s", newMood, moodInfo.Emoji),
					raw:     false,
				})
			} else {
				m.messages = append(m.messages, message{
					role:    "system",
					content: fmt.Sprintf("❌ Unknown mood: **%s**", newMood),
					raw:     false,
				})
			}
		}
		m.updateViewport()

	case "/clear":
		m.agent.Reset()
		m.messages = []message{{
			role:    "system",
			content: "✓ **Conversation cleared**",
			raw:     false,
		}}
		m.updateViewport()

	case "/quit", "/exit":
		return *m, tea.Quit

	default:
		m.messages = append(m.messages, message{
			role:    "system",
			content: fmt.Sprintf("❌ Unknown command: **%s**", parts[0]),
			raw:     false,
		})
		m.updateViewport()
	}

	return *m, nil
}

func (m *model) switchModel(modelStr string) (tea.Model, tea.Cmd) {
	oldModel := config.Get().Model
	config.Get().Model = modelStr

	newAgent, err := agent.New()
	if err != nil {
		config.Get().Model = oldModel
		m.messages = append(m.messages, message{
			role:    "system",
			content: fmt.Sprintf("❌ Failed to switch model: %v", err),
			raw:     true,
		})
	} else {
		// Keep conversation, just switch the agent
		m.agent = newAgent
		cfg := config.Get()
		m.agent.SetSystemPrompt(mood.GetSystemPrompt(cfg.Mood))

		m.messages = append(m.messages, message{
			role:    "system",
			content: fmt.Sprintf("✓ **Switched to:** %s", modelStr),
			raw:     false,
		})
	}

	m.updateViewport()
	return *m, nil
}

func (m *model) streamResponse(input string) tea.Cmd {
	return func() tea.Msg {
		// Get conversation history and add new message
		// For now, simple streaming without full conversation context
		opts := &models.Options{
			Model:       m.agent.GetModel().ID,
			MaxTokens:   config.Get().MaxTokens,
			Temperature: config.Get().Temperature,
		}

		// Build messages from conversation
		messages := make([]models.Message, 0)
		for _, msg := range m.messages {
			if msg.role == "user" || msg.role == "assistant" {
				messages = append(messages, models.Message{
					Role:    msg.role,
					Content: msg.content,
				})
			}
		}
		messages = append(messages, models.Message{
			Role:    "user",
			Content: input,
		})

		stream, err := m.agent.GetProvider().Stream(m.ctx, messages, opts)
		if err != nil {
			return errMsg{err}
		}

		// Create a command that listens to the stream
		return func() tea.Msg {
			for chunk := range stream {
				if chunk.Error != nil {
					return errMsg{chunk.Error}
				}
				if chunk.Done {
					return responseDoneMsg{}
				}
				if chunk.Content != "" {
					return responseChunkMsg(chunk.Content)
				}
			}
			return responseDoneMsg{}
		}
	}
}

func (m *model) updateViewport() {
	var content strings.Builder

	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			content.WriteString(userStyle.Render("You: "))
			content.WriteString(msg.content)
			content.WriteString("\n\n")

		case "assistant":
			content.WriteString(assistantStyle.Render("Nero: "))
			if msg.raw {
				content.WriteString(msg.content)
			} else {
				rendered, _ := m.renderer.Render(msg.content)
				content.WriteString(rendered)
			}
			content.WriteString("\n\n")

		case "system":
			if msg.raw {
				content.WriteString(msg.content)
			} else {
				rendered, _ := m.renderer.Render(msg.content)
				content.WriteString(rendered)
			}
			content.WriteString("\n\n")
		}
	}

	// Show streaming response
	if m.streaming && m.currentResponse != "" {
		content.WriteString(assistantStyle.Render("Nero: "))
		rendered, _ := m.renderer.Render(m.currentResponse + " ▊")
		content.WriteString(rendered)
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var view strings.Builder

	// Main viewport
	view.WriteString(m.viewport.View())
	view.WriteString("\n")
	view.WriteString(strings.Repeat("─", m.width))
	view.WriteString("\n")

	// Command suggestions
	if m.showSuggestions && len(m.suggestions) > 0 {
		view.WriteString(statusStyle.Render("Suggestions:") + "\n")
		for i, sug := range m.suggestions {
			if i == m.selectedSuggestion {
				view.WriteString(selectedSuggestionStyle.Render(" → " + sug))
			} else {
				view.WriteString(suggestionStyle.Render("   " + sug))
			}
			view.WriteString("\n")
		}
		view.WriteString(strings.Repeat("─", m.width))
		view.WriteString("\n")
	}

	// Input area
	view.WriteString(m.textarea.View())
	view.WriteString("\n")

	// Status bar
	status := ""
	if m.streaming {
		status = statusStyle.Render("⏳ Streaming...")
	} else if m.ctrlCPressed {
		status = errorStyle.Render("⚠️  Press Ctrl+C again to quit")
	} else {
		cfg := config.Get()
		status = statusStyle.Render(fmt.Sprintf("Model: %s | / for commands | Ctrl+C twice to quit", cfg.Model))
	}
	view.WriteString(status)

	return view.String()
}

// Run starts the TUI
func Run() error {
	p := tea.NewProgram(
		initialModel(),
		// NO tea.WithAltScreen() - render on primary console!
	)

	_, err := p.Run()
	return err
}
