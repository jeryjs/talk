package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"nero/behavioral"
	"nero/capabilities"
	"nero/capabilities/ai"
	"nero/cli"
	extensions "nero/extensions/nero"
	"nero/kernel"
)

func main() {
	// Initialize runtime
	runtime := kernel.NewRuntime()
	if err := runtime.Start(); err != nil {
		log.Fatal("Failed to start runtime:", err)
	}
	defer runtime.Stop()

	// Initialize AI providers
	aiRouter := ai.LoadProviders()

	// Initialize capabilities (unused for now)
	_ = capabilities.NewLoader("extensions")
	_ = capabilities.NewLifecycleManager()
	_ = capabilities.NewMesh()

	// Load extensions (future implementation)
	// if err := loader.LoadAll(); err != nil {
	// 	log.Printf("Warning: Failed to load some extensions: %v", err)
	// }

	// Initialize @nero extension
	neroExt := extensions.NewNeroExtension()
	if err := neroExt.Initialize(); err != nil {
		log.Fatal("Failed to initialize @nero extension:", err)
	}

	// Initialize behavioral engine
	engine := behavioral.NewEngine()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		log.Fatal("Failed to start behavioral engine:", err)
	}
	defer engine.Stop()

	// Initialize CLI
	repl := cli.NewAdvancedREPL()
	visualizer := cli.NewVisualizer()

	// Show welcome
	visualizer.ShowWelcome()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Printf("\n💜 *ruffles feathers* Leaving already?\n")
		fmt.Printf("   Well... it's not like I'll miss you or anything!\n\n")
		fmt.Printf("   *quietly* ...come back soon, okay?\n\n")
		fmt.Printf("✨ Nero signing off...\n")
		cancel()
	}()

	// Main REPL loop - properly blocks on input
	for {
		input, err := repl.ReadInput()
		if err != nil {
			// Handle EOF (Ctrl+C) gracefully
			if err == io.EOF {
				fmt.Println("\nShutting down Nero...")
				return
			}
			repl.ShowTransientError(err)
			continue
		}

		// Check for cancellation after getting input
		select {
		case <-ctx.Done():
			return
		default:
		}

		if input == "" {
			continue
		}

		// Handle special commands
		if handleSpecialCommand(input, repl, visualizer, neroExt) {
			continue
		}

		// Process with AI (with animated loading)
		if err := processAIRequest(input, aiRouter, engine, repl); err != nil {
			repl.ShowTransientError(err)
		}
	}
}

func handleSpecialCommand(input string, repl *cli.AdvancedREPL, visualizer *cli.Visualizer, neroExt *extensions.NeroExtension) bool {
	switch input {
	case "/help":
		repl.PrintMessage(`Nero Commands:
  /help       - Show this help
  /clear      - Clear screen  
  /status     - Show system status
  /quit, /exit - Exit Nero
  
  @nero <cmd> - Execute @nero extension commands
  #resource   - Access system resources
  Regular text - Chat with Nero`)
		return true

	case "/clear":
		visualizer.Clear()
		return true

	case "/status":
		config := neroExt.GetConfig()
		visualizer.RenderNeroStatus(config.Personality, 0, 0)
		return true

	case "/quit", "/exit":
		fmt.Printf("💜 *ruffles feathers* Leaving already?\n")
		fmt.Printf("   Well... it's not like I'll miss you or anything!\n\n")
		fmt.Printf("   *quietly* ...come back soon, okay?\n\n")
		fmt.Printf("✨ Nero signing off...\n")
		os.Exit(0)
		return true
	}

	// Handle @nero commands
	if len(input) > 5 && input[:5] == "@nero" {
		args := parseCommand(input[5:])
		if len(args) > 0 {
			result, err := neroExt.ExecuteCommand(args[0], args[1:])
			if err != nil {
				repl.PrintError(err)
			} else {
				repl.PrintMessage(result)
			}
		} else {
			// Show @nero help
			result, _ := neroExt.ExecuteCommand("help", []string{})
			repl.PrintMessage(result)
		}
		return true
	}

	return false
}

func processAIRequest(input string, aiRouter *ai.Router, engine *behavioral.Engine, repl *cli.AdvancedREPL) error {
	mainModel := aiRouter.GetMainModel()
	if mainModel == nil {
		// Show animated "searching for models" message
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		animation := cli.NewCursorAnimation()
		animation.Start(ctx)
		fmt.Printf("\r%sSearching for AI models...%s", cli.Yellow, cli.Reset)

		<-ctx.Done()
		animation.Stop()
		fmt.Print("\r" + strings.Repeat(" ", 30) + "\r")

		return fmt.Errorf("no AI models available - try: ollama pull qwen2.5:3b")
	}

	// Get personality context from behavioral engine
	mood := engine.GetCurrentMood()
	personality := engine.GetPersonalityPrompt()

	// Build messages
	messages := []ai.Message{
		{Role: "system", Content: personality},
		{Role: "user", Content: input},
	}

	// Add mood context
	if mood != "" {
		messages = append(messages, ai.Message{
			Role:    "system",
			Content: fmt.Sprintf("<mood>%s</mood>", mood),
		})
	}

	// Stream response with visual feedback
	stream := make(chan string, 100)
	ctx := context.Background()

	go func() {
		if err := mainModel.Chat(ctx, messages, stream); err != nil {
			repl.ShowTransientError(err)
		}
	}()

	// Render streaming response with fancy visuals
	repl.StreamResponse(stream)

	return nil
}

func parseCommand(cmd string) []string {
	// Simple command parsing - split by spaces
	var args []string
	var current string
	inQuotes := false

	for _, char := range cmd {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes && current != "" {
				args = append(args, current)
				current = ""
			} else if inQuotes {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}
