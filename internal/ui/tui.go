package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nero/internal/agent"
	"nero/internal/config"
	"nero/internal/mood"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	userStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	assistantStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	systemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	statusStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	suggestionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("235"))
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("62")).Bold(true)
)

type chatMessage struct {
	role    string
	content string
}

type chatModel struct {
	viewport        viewport.Model
	textarea        textarea.Model
	agent           *agent.Agent
	ctx             context.Context
	messages        []chatMessage
	width           int
	height          int
	ready           bool
	streaming       bool
	streamBuf       string
	ctrlCCount      int
	ctrlCTime       time.Time
	showSuggestions bool
	suggestions     []string
	suggestionIdx   int
	renderer        *glamour.TermRenderer
}

type streamChunkMsg string
type streamDoneMsg struct{}
type errorMsg struct{ err error }

func Run() error {
	cfg := config.Get()
	ag, err := agent.New()
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	ag.SetSystemPrompt(mood.GetSystemPrompt(cfg.Mood))

	ta := textarea.New()
	ta.Placeholder = "Type a message... (/ for commands)"
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(80, 20)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	welcome := chatMessage{
		role:    "system",
		content: fmt.Sprintf("🤖 **Nero** - AI Assistant\n\nModel: **%s** | Mood: **%s**\n\nType `/help` for commands", cfg.Model, cfg.Mood),
	}

	m := chatModel{
		viewport:      vp,
		textarea:      ta,
		agent:         ag,
		ctx:           context.Background(),
		messages:      []chatMessage{welcome},
		renderer:      renderer,
		suggestionIdx: -1,
	}

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}

func (m chatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.viewport.Height = msg.Height - 8
			m.textarea.SetWidth(msg.Width - 4)
			m.renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(msg.Width-4),
			)
			m.updateViewport()
			m.ready = true
		}

	case tea.KeyMsg:
		// Ctrl+C handling
		if msg.Type == tea.KeyCtrlC {
			if m.ctrlCCount > 0 && time.Since(m.ctrlCTime) < 3*time.Second {
				return m, tea.Quit
			}
			m.ctrlCCount = 1
			m.ctrlCTime = time.Now()
			return m, tea.Batch(tiCmd, vpCmd)
		} else {
			m.ctrlCCount = 0
		}

		input := m.textarea.Value()

		// Command suggestions
		if strings.HasPrefix(input, "/") {
			m.showSuggestions = true
			m.updateSuggestions(input)

			switch msg.Type {
			case tea.KeyUp:
				if m.suggestionIdx > 0 {
					m.suggestionIdx--
				}
				return m, tea.Batch(tiCmd, vpCmd)
			case tea.KeyDown:
				if m.suggestionIdx < len(m.suggestions)-1 {
					m.suggestionIdx++
				}
				return m, tea.Batch(tiCmd, vpCmd)
			case tea.KeyTab:
				if m.suggestionIdx >= 0 && m.suggestionIdx < len(m.suggestions) {
					m.textarea.SetValue(m.suggestions[m.suggestionIdx] + " ")
					m.showSuggestions = false
					m.textarea.CursorEnd()
					return m, tea.Batch(tiCmd, vpCmd)
				}
			}
		} else {
			m.showSuggestions = false
		}

		// Enter to send
		if msg.Type == tea.KeyEnter && !msg.Alt && !m.streaming {
			input = strings.TrimSpace(input)
			if input == "" {
				return m, tea.Batch(tiCmd, vpCmd)
			}

			m.textarea.Reset()
			m.showSuggestions = false

			// Handle commands
			if strings.HasPrefix(input, "/") {
				return m.handleCommand(input), nil
			}

			// Add user message
			m.messages = append(m.messages, chatMessage{role: "user", content: input})
			m.updateViewport()

			// Start streaming
			m.streaming = true
			m.streamBuf = ""
			return m, m.streamResponse()
		}

	case streamChunkMsg:
		m.streamBuf += string(msg)
		m.updateViewport()
		return m, m.readNextChunk()

	case streamDoneMsg:
		m.streaming = false
		if m.streamBuf != "" {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: m.streamBuf,
			})
			m.streamBuf = ""
		}
		m.updateViewport()
		return m, nil

	case errorMsg:
		m.streaming = false
		m.messages = append(m.messages, chatMessage{
			role:    "system",
			content: fmt.Sprintf("❌ **Error:** %v", msg.err),
		})
		m.updateViewport()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *chatModel) updateSuggestions(input string) {
	commands := []string{"/help", "/model", "/mood helpful", "/mood creative", "/mood professional", "/clear", "/quit"}
	m.suggestions = []string{}

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, strings.ToLower(input)) {
			m.suggestions = append(m.suggestions, cmd)
		}
	}

	if m.suggestionIdx >= len(m.suggestions) {
		m.suggestionIdx = len(m.suggestions) - 1
	}
	if m.suggestionIdx < 0 && len(m.suggestions) > 0 {
		m.suggestionIdx = 0
	}
}

func (m *chatModel) handleCommand(cmd string) tea.Model {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return *m
	}

	switch parts[0] {
	case "/help":
		m.messages = append(m.messages, chatMessage{
			role:    "system",
			content: "**Commands:**\n• `/help` - Show help\n• `/model [provider/model]` - Switch model\n• `/mood [name]` - Change mood\n• `/clear` - Clear chat\n• `/quit` - Exit\n\n**Shortcuts:** `Ctrl+C` twice to quit",
		})
		m.updateViewport()

	case "/model":
		if len(parts) > 1 {
			m.switchModel(parts[1])
		} else {
			m.messages = append(m.messages, chatMessage{
				role:    "system",
				content: fmt.Sprintf("**Current model:** %s\n\nUsage: `/model provider/model`", config.Get().Model),
			})
			m.updateViewport()
		}

	case "/mood":
		if len(parts) > 1 {
			moodName := parts[1]
			if prompt := mood.GetSystemPrompt(moodName); prompt != "" {
				m.agent.SetSystemPrompt(prompt)
				info, _ := mood.Get(moodName)
				m.messages = append(m.messages, chatMessage{
					role:    "system",
					content: fmt.Sprintf("✓ **Mood:** %s %s", moodName, info.Emoji),
				})
			} else {
				m.messages = append(m.messages, chatMessage{
					role:    "system",
					content: fmt.Sprintf("❌ Unknown mood: **%s**", moodName),
				})
			}
			m.updateViewport()
		}

	case "/clear":
		m.agent.Reset()
		m.messages = []chatMessage{{role: "system", content: "✓ **Cleared**"}}
		m.updateViewport()

	case "/quit", "/exit":
		return *m // Will quit via tea.Quit in calling code
	}

	return *m
}

func (m *chatModel) switchModel(modelStr string) {
	config.Get().Model = modelStr
	newAgent, err := agent.New()
	if err != nil {
		m.messages = append(m.messages, chatMessage{
			role:    "system",
			content: fmt.Sprintf("❌ Failed: %v", err),
		})
	} else {
		m.agent = newAgent
		m.agent.SetSystemPrompt(mood.GetSystemPrompt(config.Get().Mood))
		m.messages = append(m.messages, chatMessage{
			role:    "system",
			content: fmt.Sprintf("✓ **Switched to:** %s", modelStr),
		})
	}
	m.updateViewport()
}

var streamChan <-chan tea.Msg

func (m *chatModel) streamResponse() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement proper streaming via agent
		// For now, use regular Chat
		response, err := m.agent.Chat(m.ctx, m.messages[len(m.messages)-1].content)
		if err != nil {
			return errorMsg{err}
		}

		// Simulate streaming by sending it as one chunk
		return streamChunkMsg(response)
	}
}

func (m *chatModel) readNextChunk() tea.Cmd {
	return func() tea.Msg {
		return streamDoneMsg{}
	}
}

func (m *chatModel) updateViewport() {
	var content strings.Builder

	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			content.WriteString(userStyle.Render("You: "))
			content.WriteString(msg.content)
			content.WriteString("\n\n")

		case "assistant":
			content.WriteString(assistantStyle.Render("Nero: "))
			rendered, _ := m.renderer.Render(msg.content)
			content.WriteString(rendered)
			content.WriteString("\n\n")

		case "system":
			rendered, _ := m.renderer.Render(msg.content)
			content.WriteString(rendered)
			content.WriteString("\n\n")
		}
	}

	// Show streaming
	if m.streaming && m.streamBuf != "" {
		content.WriteString(assistantStyle.Render("Nero: "))
		rendered, _ := m.renderer.Render(m.streamBuf + " ▊")
		content.WriteString(rendered)
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func (m chatModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var view strings.Builder

	view.WriteString(m.viewport.View())
	view.WriteString("\n")
	view.WriteString(strings.Repeat("─", m.width))
	view.WriteString("\n")

	// Suggestions
	if m.showSuggestions && len(m.suggestions) > 0 {
		view.WriteString(statusStyle.Render("Commands:") + "\n")
		for i, sug := range m.suggestions {
			if i == m.suggestionIdx {
				view.WriteString(selectedStyle.Render(" → " + sug))
			} else {
				view.WriteString(suggestionStyle.Render("   " + sug))
			}
			view.WriteString("\n")
		}
		view.WriteString(strings.Repeat("─", m.width))
		view.WriteString("\n")
	}

	view.WriteString(m.textarea.View())
	view.WriteString("\n")

	// Status
	status := ""
	if m.streaming {
		status = statusStyle.Render("⏳ Generating...")
	} else if m.ctrlCCount > 0 {
		status = errorStyle.Render("⚠️  Press Ctrl+C again to quit")
	} else {
		status = statusStyle.Render(fmt.Sprintf("Model: %s | / for commands", config.Get().Model))
	}
	view.WriteString(status)

	return view.String()
}
