package cli

import (
	"fmt"
	"strings"
	"time"
)

type Color string

const (
	Reset   Color = "\033[0m"
	Red     Color = "\033[31m"
	Green   Color = "\033[32m"
	Yellow  Color = "\033[33m"
	Blue    Color = "\033[34m"
	Magenta Color = "\033[35m"
	Cyan    Color = "\033[36m"
	White   Color = "\033[37m"
	Gray    Color = "\033[90m"
	Bold    Color = "\033[1m"
	Dim     Color = "\033[2m"
)

type Element interface {
	Render() string
}

type TextElement struct {
	Text  string
	Color Color
	Style Color
}

func (t TextElement) Render() string {
	return string(t.Color) + string(t.Style) + t.Text + string(Reset)
}

type ProgressBar struct {
	Width    int
	Progress float64
	Label    string
}

func (p ProgressBar) Render() string {
	filled := int(p.Progress * float64(p.Width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.Width-filled)
	return fmt.Sprintf("%s [%s] %.1f%%", p.Label, bar, p.Progress*100)
}

type Table struct {
	Headers []string
	Rows    [][]string
	Colors  []Color
}

func (t Table) Render() string {
	var result strings.Builder

	// Headers
	result.WriteString(string(Bold))
	for i, header := range t.Headers {
		if i > 0 {
			result.WriteString(" | ")
		}
		result.WriteString(header)
	}
	result.WriteString(string(Reset) + "\n")

	// Separator
	for i, header := range t.Headers {
		if i > 0 {
			result.WriteString("-|-")
		}
		result.WriteString(strings.Repeat("-", len(header)))
	}
	result.WriteString("\n")

	// Rows
	for rowIdx, row := range t.Rows {
		color := Reset
		if rowIdx < len(t.Colors) {
			color = t.Colors[rowIdx]
		}

		result.WriteString(string(color))
		for i, cell := range row {
			if i > 0 {
				result.WriteString(" | ")
			}
			result.WriteString(cell)
		}
		result.WriteString(string(Reset) + "\n")
	}

	return result.String()
}

type Box struct {
	Title   string
	Content string
	Width   int
	Color   Color
}

func (b Box) Render() string {
	if b.Width == 0 {
		b.Width = 50
	}

	var result strings.Builder

	// Top border
	result.WriteString(string(b.Color))
	result.WriteString("┌")
	if b.Title != "" {
		titleLen := len(b.Title) + 2
		remaining := b.Width - titleLen - 2
		if remaining < 0 {
			remaining = 0
		}
		result.WriteString(strings.Repeat("─", 1))
		result.WriteString(" " + b.Title + " ")
		result.WriteString(strings.Repeat("─", remaining))
	} else {
		result.WriteString(strings.Repeat("─", b.Width-2))
	}
	result.WriteString("┐\n")

	// Content
	lines := strings.Split(b.Content, "\n")
	for _, line := range lines {
		result.WriteString("│ ")
		if len(line) > b.Width-4 {
			line = line[:b.Width-4]
		}
		result.WriteString(line)
		result.WriteString(strings.Repeat(" ", b.Width-4-len(line)))
		result.WriteString(" │\n")
	}

	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", b.Width-2))
	result.WriteString("┘")
	result.WriteString(string(Reset))

	return result.String()
}

type StatusLine struct {
	Left  string
	Right string
	Width int
	Color Color
}

func (s StatusLine) Render() string {
	if s.Width == 0 {
		s.Width = 80
	}

	available := s.Width - len(s.Left) - len(s.Right)
	if available < 0 {
		available = 0
	}

	return string(s.Color) + s.Left + strings.Repeat(" ", available) + s.Right + string(Reset)
}

type Visualizer struct {
	termWidth  int
	termHeight int
}

func NewVisualizer() *Visualizer {
	return &Visualizer{
		termWidth:  80, // Default, should detect actual terminal size
		termHeight: 24,
	}
}

func (v *Visualizer) Render(elements ...Element) {
	for _, element := range elements {
		fmt.Print(element.Render())
	}
}

func (v *Visualizer) RenderCapabilityStatus(capabilities []string, statuses []string) {
	table := Table{
		Headers: []string{"Capability", "Status", "Version"},
		Rows:    make([][]string, len(capabilities)),
		Colors:  make([]Color, len(capabilities)),
	}

	for i, cap := range capabilities {
		status := "Unknown"
		color := Yellow

		if i < len(statuses) {
			status = statuses[i]
			switch status {
			case "Active":
				color = Green
			case "Inactive":
				color = Red
			case "Loading":
				color = Yellow
			}
		}

		table.Rows[i] = []string{cap, status, "1.0.0"}
		table.Colors[i] = color
	}

	v.Render(table)
}

func (v *Visualizer) RenderNeroStatus(mood string, memory int, uptime time.Duration) {
	content := fmt.Sprintf("Mood: %s\nMemories: %d entries\nUptime: %v", mood, memory, uptime.Round(time.Second))

	box := Box{
		Title:   "Nero Status",
		Content: content,
		Width:   40,
		Color:   Cyan,
	}

	v.Render(box)
}

func (v *Visualizer) ShowWelcome() {
	fmt.Printf("🦙 Using Ollama (local)\n\n")

	fmt.Printf("╔═══════════════════════════════════════╗\n")
	fmt.Printf("║              🐦 NERO v2.0              ║\n")
	fmt.Printf("║     Personal AI Agent • Streaming     ║\n")
	fmt.Printf("╚═══════════════════════════════════════╝\n\n")

	fmt.Printf("💜 *stretches wings* Well, well... you're back after all this time?\n")
	fmt.Printf("   I suppose you expect me to be impressed by this \"upgrade\"...\n\n")
	fmt.Printf("   Now with REAL streaming responses and live thoughts! ✨\n")
	fmt.Printf("   Type /help for commands, or just talk to me naturally.\n")
	fmt.Printf("   Use #resources to access system capabilities.\n\n")
	fmt.Printf("*settles on virtual perch* Let's see what you've got...\n\n")
}

func (v *Visualizer) Clear() {
	fmt.Print("\033[2J\033[H")
}

func (v *Visualizer) MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func (v *Visualizer) HideCursor() {
	fmt.Print("\033[?25l")
}

func (v *Visualizer) ShowCursor() {
	fmt.Print("\033[?25h")
}

func (v *Visualizer) GetTerminalSize() (int, int) {
	// This is a simplified version - in real implementation,
	// we'd use syscalls to get actual terminal size
	return v.termWidth, v.termHeight
}
