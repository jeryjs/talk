package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"nero/internal/agent"
	"nero/internal/config"
	"nero/internal/mood"
)

// Run starts the interactive console interface
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

	// Welcome message
	fmt.Printf("\n🤖 Nero - AI Assistant\n")
	fmt.Printf("Mood: %s | Using: %s\n", cfg.Mood, cfg.Model)
	fmt.Printf("(Shortcuts: op/gpt-4o, gq/llama-3.3-70b, ol/llama3.2)\n")
	fmt.Printf("\nType /help for commands, Ctrl+C to quit\n")
	fmt.Println(strings.Repeat("─", 60))

	ctx := context.Background()
	currentMood := cfg.Mood

	// Create reader for input
	reader := bufio.NewReader(os.Stdin)

	// Input loop
	for {
		fmt.Print("\nYou: ")

		// Read full line
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
			parts := strings.Fields(input)
			switch parts[0] {
			case "/help":
				fmt.Println("\n📚 Commands:")
				fmt.Println("  /help              - Show this help")
				fmt.Println("  /mood [name]       - Change personality")
				fmt.Println("  /model [provider/model] - Switch AI model")
				fmt.Println("  /clear             - Clear conversation")
				fmt.Println("  /quit              - Exit")
				continue

			case "/mood":
				if len(parts) < 2 {
					fmt.Printf("Current mood: %s\n", currentMood)
				} else {
					newMood := parts[1]
					if moodPrompt := mood.GetSystemPrompt(newMood); moodPrompt != "" {
						currentMood = newMood
						ag.SetSystemPrompt(moodPrompt)
						moodInfo, _ := mood.Get(newMood)
						fmt.Printf("✓ Mood set to: %s %s\n", newMood, moodInfo.Emoji)
					} else {
						fmt.Printf("❌ Unknown mood: %s (type /help)\n", newMood)
					}
				}
				continue

			case "/model":
				if len(parts) < 2 {
					fmt.Printf("Current model: %s\n", cfg.Model)
					fmt.Println("Usage: /model <provider/model>")
					fmt.Println("Examples: /model op/gpt-4o, /model gq/llama-3.3-70b")
				} else {
					newModel := parts[1]
					// Switch to the new model
					cfg.Model = newModel
					ag.SetSystemPrompt(mood.GetSystemPrompt(currentMood))
					fmt.Printf("✓ Switched to model: %s\n", newModel)
				}
				continue

			case "/clear":
				ag.Reset()
				fmt.Println("✓ Conversation cleared")
				continue

			case "/quit", "/exit":
				fmt.Println("\n👋 Goodbye!")
				return nil

			default:
				fmt.Printf("❌ Unknown command: %s (type /help)\n", parts[0])
				continue
			}
		}

		// Send to AI
		fmt.Print("\nNero: ")
		response, err := ag.Chat(ctx, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println(strings.Repeat("─", 60))
	}
}
