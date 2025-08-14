package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
		commands: []string{"help", "status", "mood", "run", "open", "exit", "quit", "provider", "stream", "thoughts"},
	}
}

// Provide the terminal-based interaction system with streaming
type Interface struct {
	core            *kernel.Core
	behavior        *behavioral.Engine
	systemProvider  *providers.SystemProvider
	commands        map[string]Command
	completer       *Completer
	renderer        *StreamingRenderer
	thoughtRenderer *ThoughtRenderer
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

// Create a new CLI interface with advanced streaming capabilities
func NewInterface(runtime interface{}) *Interface {
	// Initialize the new kernel core
	core := kernel.NewCore()
	if err := core.Initialize(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize AI core: %v\n", err)
		return nil
	}

	// Show which providers are available
	availableProviders := core.GetAvailableProviders()
	if len(availableProviders) == 0 {
		color.New(color.FgRed).Println("❌ No AI providers available!")
		color.New(color.FgYellow).Println("💡 Install Ollama or set API keys (OPENAI_API_KEY, GEMINI_API_KEY, GROQ_API_KEY)")
		return nil
	}

	// Show active provider
	activeProvider := core.GetActiveProvider()
	switch activeProvider {
	case "ollama":
		color.New(color.FgGreen).Println("🦙 Using Ollama (local)")
	case "openai":
		color.New(color.FgBlue).Println("🌐 Using OpenAI")
	case "gemini":
		color.New(color.FgMagenta).Println("💎 Using Google Gemini")
	case "groq":
		color.New(color.FgYellow).Println("⚡ Using Groq")
	}

	cli := &Interface{
		core:            core,
		behavior:        behavioral.NewEngine(),
		systemProvider:  providers.NewSystemProvider(),
		commands:        make(map[string]Command),
		completer:       NewCompleter(),
		renderer:        NewStreamingRenderer(),
		thoughtRenderer: NewThoughtRenderer(),
	}

	// Register default commands
	cli.registerDefaultCommands()

	return cli
}

// Start the interactive CLI session with streaming support
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

// Process user input with streaming support
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

	// Regular chat input - now with streaming!
	cli.handleChatStreaming(input)
}

// Handle commands
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

// Handle resource access
func (cli *Interface) handleResourceAccess(input string) {
	// Extract resources (e.g., #terminal, #screen, #code)
	resources := cli.extractResources(input)

	// Process each resource
	for _, resource := range resources {
		cli.processResource(resource, input)
	}

	// Continue with normal chat processing
	cleanInput := cli.removeResourceTags(input)
	cli.handleChatStreaming(cleanInput)
}

// Handle chat with real-time streaming and thoughts
func (cli *Interface) handleChatStreaming(input string) {
	ctx := context.Background()

	// Process through behavioral engine first for personality
	if cli.behavior != nil {
		response, err := cli.behavior.ProcessResponse(ctx, input)
		if err != nil {
			cli.printError(fmt.Sprintf("Behavioral processing error: %v", err))
		} else {
			// Display the behavioral response with streaming simulation
			cli.simulateStreaming(response.Text, response.Tone)
			return
		}
	}

	// Fallback to direct core processing
	req := &kernel.AIRequest{
		Messages: []providers.Message{
			{Role: "user", Content: input},
		},
		EnableStream:   true,
		EnableThoughts: true,
		Temperature:    0.8,
		MaxTokens:      500,
	}

	// Process request through kernel core
	response, err := cli.core.ProcessRequest(ctx, req)
	if err != nil {
		cli.printError(fmt.Sprintf("AI Error: %v", err))
		return
	}

	// Handle streaming response
	if response.StreamID != "" {
		cli.handleStreamingResponse(ctx, response.StreamID, input)
	} else {
		// Fallback to non-streaming display
		cli.displayResponse(&behavioral.Response{
			Text: response.Content,
			Tone: "neutral",
		})
	}
}

// Simulate streaming for behavioral responses
func (cli *Interface) simulateStreaming(text string, tone string) {
	// Start streaming visualization
	cli.renderer.StartStreaming()

	// Simulate word-by-word streaming
	words := strings.Fields(text)
	for _, word := range words {
		cli.renderer.AddContent(word+" ", false)
		// Small delay to simulate typing
		time.Sleep(time.Millisecond * 50)
	}

	cli.renderer.StopStreaming()
}

// Handle real-time streaming response with thoughts
func (cli *Interface) handleStreamingResponse(ctx context.Context, streamID string, originalInput string) {
	// Get the streaming channel
	streamChan, exists := cli.core.GetStreamChannel(streamID)
	if !exists {
		cli.printError("Stream not found")
		return
	}

	// Start streaming visualization
	cli.renderer.StartStreaming()

	var fullResponse strings.Builder
	thoughtsStarted := false

	// Process streaming chunks in real-time
	for {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				// Stream closed
				cli.renderer.StopStreaming()
				return
			}

			if chunk.Error != nil {
				cli.printError(fmt.Sprintf("Streaming error: %v", chunk.Error))
				cli.renderer.StopStreaming()
				return
			}

			if chunk.IsThought {
				// Handle thought streaming
				if !thoughtsStarted {
					thoughtsStarted = true
					cli.renderer.thoughtRenderer.StartThinking()
				}
				cli.renderer.AddThought(chunk.Content)
			} else {
				// Handle regular response streaming
				if thoughtsStarted {
					// Stop thoughts and start response
					cli.renderer.thoughtRenderer.StopThinking()
					thoughtsStarted = false
				}

				cli.renderer.AddContent(chunk.Content, false)
				fullResponse.WriteString(chunk.Content)
			}

			if chunk.Done {
				cli.renderer.StopStreaming()
				return
			}

		case <-ctx.Done():
			cli.renderer.StopStreaming()
			return
		}
	}
}

// Display non-streaming response (fallback)
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
		colorFunc = color.New(color.FgMagenta).Sprintf
	}

	// Print response with formatting
	fmt.Printf("%s\n\n", colorFunc(response.Text))
}

// Print welcome message
func (cli *Interface) printWelcome() {
	welcome := `
╔═══════════════════════════════════════╗
║              🐦 NERO v2.0              ║
║     Personal AI Agent • Streaming     ║
╚═══════════════════════════════════════╝

💜 *stretches wings* Well, well... you're back after all this time?
   I suppose you expect me to be impressed by this "upgrade"...
   
   Now with REAL streaming responses and live thoughts! ✨
   Type /help for commands, or just talk to me naturally.
   Use #resources to access system capabilities.
   
*settles on virtual perch* Let's see what you've got...

`
	color.New(color.FgMagenta, color.Bold).Print(welcome)
}

// Print goodbye message
func (cli *Interface) printGoodbye() {
	goodbye := `
💜 *ruffles feathers* Leaving already? 
   Well... it's not like I'll miss you or anything! 
   
   *quietly* ...come back soon, okay?

✨ Nero signing off...
`
	color.New(color.FgMagenta).Print(goodbye)
}

// Print input prompt
func (cli *Interface) printPrompt() {
	var mood string
	if cli.behavior != nil {
		state := cli.behavior.GetState()
		mood = getMoodEmoji(state.Mood.Primary)
	} else {
		mood = "🐦"
	}

	color.New(color.FgHiBlack).Printf("╭─")
	color.New(color.FgMagenta).Printf(" %s nero", mood)
	color.New(color.FgHiBlack).Printf(" ──────────────────────────────────────────────────")
	fmt.Println()
	color.New(color.FgHiBlack).Printf("╰─ ")
	color.New(color.FgHiYellow).Printf("❯ ")
}

// Print error messages
func (cli *Interface) printError(message string) {
	color.New(color.FgRed).Printf("❌ %s\n", message)
}

// Extract resource tags from input
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

// Process specific resource access
func (cli *Interface) processResource(resource string, input string) {
	color.New(color.FgCyan).Printf("🔍 Accessing %s resource...\n", resource)

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

// Handle terminal resource access
func (cli *Interface) handleTerminalResource() {
	// Get current working directory
	cwd, _ := cli.systemProvider.GetWorkingDirectory()
	color.New(color.FgGreen).Printf("📟 Terminal context captured - CWD: %s\n", cwd)
}

// Handle screen resource access
func (cli *Interface) handleScreenResource() {
	color.New(color.FgGreen).Println("🖥️  Screen context captured")
}

// Handle code resource access
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

// Get mood emoji for current state
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
	case "excited":
		return "🎉"
	case "confident":
		return "😎"
	default:
		return "🐦"
	}
}
