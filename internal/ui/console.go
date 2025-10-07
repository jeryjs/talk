package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"nero/internal/agent"
	"nero/internal/config"
	"nero/internal/mood"

	"github.com/fatih/color"
)

var (
	// Colors
	userColor      = color.New(color.FgCyan, color.Bold)
	assistantColor = color.New(color.FgGreen)
	systemColor    = color.New(color.FgYellow)
	errorColor     = color.New(color.FgRed)
	dimColor       = color.New(color.FgHiBlack)

	// State
	lastCtrlC time.Time
)

// Run starts the simple console interface
func Run() error {
	cfg := config.Get()

	// Create agent
	ag, err := agent.New()
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Set initial mood
	moodPrompt := mood.GetSystemPrompt(cfg.Mood)
	ag.SetSystemPrompt(moodPrompt)

	// Welcome
	fmt.Println()
	systemColor.Println("🤖 Nero - AI Assistant")
	dimColor.Printf("Mood: %s | Provider: %s | Model: %s\n", cfg.Mood, ag.GetProvider().Name(), ag.GetModelID())
	dimColor.Println("Type /help for commands, Ctrl+C twice to quit")
	fmt.Println()

	// Handle Ctrl+C
	setupSignalHandler()

	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	for {
		// Prompt
		userColor.Print("You: ")

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if err := handleCommand(ag, input, cfg); err != nil {
				if err.Error() == "exit" {
					return nil
				}
				errorColor.Printf("Error: %v\n", err)
			}
			continue
		}

		// Send to AI with streaming
		fmt.Println()
		assistantColor.Print("Nero: ")

		err = ag.ChatStream(ctx, input, func(chunk string) {
			// Stream tokens with markdown rendering
			fmt.Print(renderMarkdown(chunk))
		})

		if err != nil {
			fmt.Println()
			errorColor.Printf("Error: %v\n", err)
		}

		fmt.Println()
		fmt.Println()
	}
}

func handleCommand(ag *agent.Agent, cmd string, cfg *config.Config) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "/help":
		showHelp()
		return nil

	case "/model":
		return handleModelCommand(ag, parts, cfg)

	case "/mood":
		return handleMoodCommand(ag, parts)

	case "/clear":
		ag.Reset()
		systemColor.Println("✓ Conversation cleared")
		return nil

	case "/quit", "/exit":
		fmt.Println()
		systemColor.Println("👋 Goodbye!")
		return fmt.Errorf("exit")

	default:
		errorColor.Printf("Unknown command: %s\n", parts[0])
		dimColor.Println("Type /help to see available commands")
		return nil
	}
}

func showHelp() {
	fmt.Println()
	systemColor.Println("📚 Available Commands:")
	fmt.Println()
	fmt.Println("  /help              - Show this help")
	fmt.Println("  /model [provider/model] - Switch model")
	fmt.Println("  /mood [name]       - Change personality")
	fmt.Println("  /clear             - Clear conversation")
	fmt.Println("  /quit              - Exit")
	fmt.Println()
	dimColor.Println("Examples:")
	dimColor.Println("  /model gq/llama-3.3-70b")
	dimColor.Println("  /mood creative")
	fmt.Println()
}

func handleModelCommand(ag *agent.Agent, parts []string, cfg *config.Config) error {
	if len(parts) < 2 {
		// Show current model
		systemColor.Printf("Current model: %s\n", ag.GetModelID())
		dimColor.Println("Usage: /model <provider/model>")
		return nil
	}

	newModel := parts[1]

	// Switch model WITHOUT clearing conversation
	if err := ag.SwitchModel(newModel); err != nil {
		return fmt.Errorf("failed to switch model: %w", err)
	}

	systemColor.Printf("✓ Switched to: %s\n", newModel)
	return nil
}

func handleMoodCommand(ag *agent.Agent, parts []string) error {
	if len(parts) < 2 {
		dimColor.Println("Available moods: helpful, creative, professional, analytical")
		return nil
	}

	newMood := parts[1]
	moodPrompt := mood.GetSystemPrompt(newMood)
	if moodPrompt == "" {
		return fmt.Errorf("unknown mood: %s", newMood)
	}

	ag.SetSystemPrompt(moodPrompt)
	moodInfo, _ := mood.Get(newMood)
	systemColor.Printf("✓ Mood set to: %s %s\n", newMood, moodInfo.Emoji)
	return nil
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range c {
			now := time.Now()
			if now.Sub(lastCtrlC) < 3*time.Second {
				// Double press - exit
				fmt.Println()
				systemColor.Println("👋 Goodbye!")
				os.Exit(0)
			}

			// First press - show warning
			lastCtrlC = now
			fmt.Println()
			dimColor.Println("Press Ctrl+C again within 3 seconds to exit")
		}
	}()
}

// Simple markdown renderer (basic support)
func renderMarkdown(text string) string {
	// TODO: Add proper markdown rendering
	// For now, just return as-is
	return text
}
