package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"nero/behavioral"
	"nero/kernel"
	"nero/providers"

	"github.com/fatih/color"
)

// Provide command autocompletion
type Completer struct {
	commands []string
}

// Create a new autocompletion handler
func NewCompleter() *Completer {
	return &Completer{
		commands: []string{"help", "status", "mood", "run", "open", "exit", "quit"},
	}
}

// Provide the terminal-based interaction system
type Interface struct {
	runtime        Runtime
	behavior       *behavioral.Engine
	aiProvider     providers.AIProvider
	systemProvider *providers.SystemProvider
	commands       map[string]Command
	completer      *Completer
}

// Represent the CLI runtime interface
type Runtime interface {
	SendEvent(event kernel.Event)
	GetCapability(name string) (kernel.Capability, bool)
}

// Represent a CLI command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(args []string, ctx *CommandContext) error
}

// Provide context for command execution
type CommandContext struct {
	Interface *Interface
	Session   string
	User      string
	Args      []string
}

// Create a new CLI interface
func NewInterface(runtime Runtime) *Interface {
	// Initialize AI provider (try Ollama first, fallback to OpenAI)
	var aiProvider providers.AIProvider

	ollama := providers.NewOllamaProvider("llama3.2")
	if ollama.IsAvailable() {
		aiProvider = ollama
		color.New(color.FgGreen).Println("🦙 Using Ollama (local)")
	} else {
		openai := providers.NewOpenAIProvider("gpt-4o-mini")
		if openai.IsAvailable() {
			aiProvider = openai
			color.New(color.FgBlue).Println("🌐 Using OpenAI")
		} else {
			color.New(color.FgRed).Println("❌ No AI provider available!")
			color.New(color.FgYellow).Println("💡 Install Ollama or set OpenAI API key")
			return nil
		}
	}

	cli := &Interface{
		runtime:        runtime,
		behavior:       behavioral.NewEngine(aiProvider),
		aiProvider:     aiProvider,
		systemProvider: providers.NewSystemProvider(),
		commands:       make(map[string]Command),
		completer:      NewCompleter(),
	}

	// Register default commands
	cli.registerDefaultCommands()

	return cli
}

// Begin the interactive CLI session
func (cli *Interface) Start(ctx context.Context) {
	cli.printWelcome()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			cli.printGoodbye()
			return
		default:
			cli.printPrompt()

			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			cli.processInput(input)
		}
	}
}

// Handle user input
func (cli *Interface) processInput(input string) {
	// Check for commands
	if strings.HasPrefix(input, "/") {
		cli.handleCommand(input)
		return
	}

	// Check for resource access
	if strings.Contains(input, "#") {
		cli.handleResourceAccess(input)
		return
	}

	// Regular chat input
	cli.handleChat(input)
}

// Process CLI commands
func (cli *Interface) handleCommand(input string) {
	parts := strings.Fields(input[1:]) // Remove leading /
	if len(parts) == 0 {
		return
	}

	commandName := parts[0]
	args := parts[1:]

	command, exists := cli.commands[commandName]
	if !exists {
		cli.printError(fmt.Sprintf("Unknown command: %s", commandName))
		return
	}

	ctx := &CommandContext{
		Interface: cli,
		Args:      args,
	}

	if err := command.Execute(args, ctx); err != nil {
		cli.printError(fmt.Sprintf("Command error: %v", err))
	}
}

// Process resource access requests
func (cli *Interface) handleResourceAccess(input string) {
	// Extract resources (e.g., #terminal, #screen, #code)
	resources := cli.extractResources(input)

	// Process each resource
	for _, resource := range resources {
		cli.processResource(resource, input)
	}

	// Continue with normal chat processing
	cleanInput := cli.removeResourceTags(input)
	cli.handleChat(cleanInput)
}

// Process regular chat input with real AI
func (cli *Interface) handleChat(input string) {
	ctx := context.Background()

	// Process through behavioral engine (now with real AI)
	response, err := cli.behavior.ProcessResponse(ctx, input)
	if err != nil {
		cli.printError(fmt.Sprintf("AI Error: %v", err))
		// Fallback response
		response = &behavioral.Response{
			Text: "Hmph! Something went wrong with my thinking... *looks annoyed*",
			Tone: "annoyed",
		}
	}

	// Display response with formatting
	cli.displayResponse(response)
}

// Show Nero's response with proper formatting
func (cli *Interface) displayResponse(response *behavioral.Response) {
	// Set color based on tone
	var colorFunc func(string, ...interface{}) string

	switch response.Tone {
	case "tsundere":
		colorFunc = color.New(color.FgMagenta, color.Bold).Sprintf
	case "sarcastic":
		colorFunc = color.New(color.FgYellow).Sprintf
	case "caring":
		colorFunc = color.New(color.FgCyan).Sprintf
	case "flustered":
		colorFunc = color.New(color.FgRed).Sprintf
	default:
		colorFunc = color.New(color.FgHiCyan).Sprintf
	}

	// Print response with formatting
	fmt.Printf("%s\n\n", colorFunc(response.Text))
}

// Display the welcome message
func (cli *Interface) printWelcome() {
	welcome := `
╔═══════════════════════════════════════╗
║              🐦 NERO v2.0              ║
║        Personal AI Agent System       ║
╚═══════════════════════════════════════╝

💜 *stretches wings* Well, well... you're back after all this time?
   I suppose you expect me to be impressed by this "upgrade"...
   
   Type /help for commands, or just talk to me normally.
   Use #resources to access system capabilities.
   
*settles on virtual perch* Let's see what you've got...

`
	color.New(color.FgMagenta, color.Bold).Print(welcome)
}

// Display the goodbye message
func (cli *Interface) printGoodbye() {
	goodbye := `
💜 *ruffles feathers* Leaving already? 
   Well... it's not like I'll miss you or anything! 
   
   *quietly* ...come back soon, okay?

✨ Nero signing off...
`
	color.New(color.FgMagenta).Print(goodbye)
}

// Display the input prompt
func (cli *Interface) printPrompt() {
	state := cli.behavior.GetState()
	mood := getMoodEmoji(state.Mood.Primary)

	color.New(color.FgHiBlack).Printf("╭─")
	color.New(color.FgMagenta).Printf(" %s nero", mood)
	color.New(color.FgHiBlack).Printf(" ──────────────────────────────────────────────────")
	fmt.Println()
	color.New(color.FgHiBlack).Printf("╰─ ")
	color.New(color.FgHiYellow).Printf("❯ ")
}

// Display error messages
func (cli *Interface) printError(message string) {
	color.New(color.FgRed).Printf("❌ %s\n", message)
}

// Find resource tags in input
func (cli *Interface) extractResources(input string) []string {
	var resources []string
	words := strings.Fields(input)

	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			resources = append(resources, strings.TrimPrefix(word, "#"))
		}
	}

	return resources
}

// Remove resource tags from input
func (cli *Interface) removeResourceTags(input string) string {
	words := strings.Fields(input)
	var cleaned []string

	for _, word := range words {
		if !strings.HasPrefix(word, "#") {
			cleaned = append(cleaned, word)
		}
	}

	return strings.Join(cleaned, " ")
}

// Handle specific resource access
func (cli *Interface) processResource(resource string, input string) {
	color.New(color.FgCyan).Printf("🔍 Accessing %s resource...\n", resource)

	// TODO: Implement actual resource handlers
	switch resource {
	case "terminal":
		cli.handleTerminalResource()
	case "screen":
		cli.handleScreenResource()
	case "code":
		cli.handleCodeResource()
	default:
		cli.printError(fmt.Sprintf("Unknown resource: %s", resource))
	}
}

// Process terminal context
func (cli *Interface) handleTerminalResource() {
	// Get current working directory
	cwd, _ := cli.systemProvider.GetWorkingDirectory()
	color.New(color.FgGreen).Printf("📟 Terminal context captured - CWD: %s\n", cwd)
}

// Process screen context
func (cli *Interface) handleScreenResource() {
	// TODO: Implement screen capture when needed
	color.New(color.FgGreen).Println("🖥️  Screen context captured")
}

// Process code context
func (cli *Interface) handleCodeResource() {
	// Get git status if in a git repo
	if output, err := cli.systemProvider.RunCommand("git", "status", "--porcelain"); err == nil {
		if strings.TrimSpace(output) != "" {
			color.New(color.FgGreen).Printf("💻 Code context captured - Git changes detected\n")
		} else {
			color.New(color.FgGreen).Printf("💻 Code context captured - Git clean\n")
		}
	} else {
		color.New(color.FgGreen).Println("💻 Code context captured")
	}
}

// Handle system commands within chat
func (cli *Interface) executeSystemCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "open":
		if len(args) > 0 {
			if err := cli.systemProvider.OpenApp(args[0]); err != nil {
				cli.printError(fmt.Sprintf("Failed to open %s: %v", args[0], err))
			} else {
				color.New(color.FgGreen).Printf("🚀 Opened %s\n", args[0])
			}
		}
	case "run":
		if len(args) > 0 {
			output, err := cli.systemProvider.RunCommand(args[0], args[1:]...)
			if err != nil {
				cli.printError(fmt.Sprintf("Command failed: %v", err))
			} else {
				color.New(color.FgCyan).Printf("Command output:\n%s\n", output)
			}
		}
	case "cd":
		if len(args) > 0 {
			if err := cli.systemProvider.ChangeDirectory(args[0]); err != nil {
				cli.printError(fmt.Sprintf("Failed to change directory: %v", err))
			} else {
				color.New(color.FgGreen).Printf("📁 Changed to %s\n", args[0])
			}
		}
	case "ls", "dir":
		cwd, _ := cli.systemProvider.GetWorkingDirectory()
		files, err := cli.systemProvider.ListDirectory(cwd)
		if err != nil {
			cli.printError(fmt.Sprintf("Failed to list directory: %v", err))
		} else {
			color.New(color.FgCyan).Println("📂 Directory contents:")
			for _, file := range files {
				fmt.Printf("  %s\n", file)
			}
		}
	default:
		// Try to run as system command
		output, err := cli.systemProvider.RunCommand(cmd, args...)
		if err != nil {
			cli.printError(fmt.Sprintf("Unknown command or error: %v", err))
		} else {
			color.New(color.FgCyan).Printf("Output:\n%s\n", output)
		}
	}
}

// Return an emoji for the current mood
func getMoodEmoji(mood string) string {
	switch mood {
	case "happy":
		return "😊"
	case "annoyed":
		return "😤"
	case "flustered":
		return "😳"
	case "pleased":
		return "😌"
	case "tired":
		return "😴"
	default:
		return "🐦"
	}
}
