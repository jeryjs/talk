package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"nero/behavioral"
	"nero/kernel"

	"github.com/fatih/color"
)

// Interface provides the terminal-based interaction system
type Interface struct {
	runtime   Runtime
	behavior  *behavioral.Engine
	commands  map[string]Command
	completer *Completer
}

// Runtime interface for CLI
type Runtime interface {
	SendEvent(event kernel.Event)
	GetCapability(name string) (kernel.Capability, bool)
}

// Command represents a CLI command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(args []string, ctx *CommandContext) error
}

// CommandContext provides context for command execution
type CommandContext struct {
	Interface *Interface
	Session   string
	User      string
	Args      []string
}

// NewInterface creates a new CLI interface
func NewInterface(runtime Runtime) *Interface {
	cli := &Interface{
		runtime:   runtime,
		behavior:  behavioral.NewEngine(),
		commands:  make(map[string]Command),
		completer: NewCompleter(),
	}

	// Register default commands
	cli.registerDefaultCommands()

	return cli
}

// Start begins the interactive CLI session
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

// processInput handles user input
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

// handleCommand processes CLI commands
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

// handleResourceAccess processes resource access requests
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

// handleChat processes regular chat input
func (cli *Interface) handleChat(input string) {
	// TODO: Route to AI model
	// For now, generate a simple response
	baseResponse := "I understand what you're saying..."

	// Process through behavioral engine
	response := cli.behavior.ProcessResponse(input, baseResponse)

	// Display response with formatting
	cli.displayResponse(response)
}

// displayResponse shows Nero's response with proper formatting
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

// printWelcome displays the welcome message
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

// printGoodbye displays the goodbye message
func (cli *Interface) printGoodbye() {
	goodbye := `
💜 *ruffles feathers* Leaving already? 
   Well... it's not like I'll miss you or anything! 
   
   *quietly* ...come back soon, okay?

✨ Nero signing off...
`
	color.New(color.FgMagenta).Print(goodbye)
}

// printPrompt displays the input prompt
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

// printError displays error messages
func (cli *Interface) printError(message string) {
	color.New(color.FgRed).Printf("❌ %s\n", message)
}

// extractResources finds resource tags in input
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

// removeResourceTags removes resource tags from input
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

// processResource handles specific resource access
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

// handleTerminalResource processes terminal context
func (cli *Interface) handleTerminalResource() {
	// TODO: Capture terminal state, command history, etc.
	color.New(color.FgGreen).Println("📟 Terminal context captured")
}

// handleScreenResource processes screen context
func (cli *Interface) handleScreenResource() {
	// TODO: Capture screen content, active windows, etc.
	color.New(color.FgGreen).Println("🖥️  Screen context captured")
}

// handleCodeResource processes code context
func (cli *Interface) handleCodeResource() {
	// TODO: Capture current code, git status, etc.
	color.New(color.FgGreen).Println("💻 Code context captured")
}

// getMoodEmoji returns an emoji for the current mood
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
