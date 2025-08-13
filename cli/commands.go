package cli

import (
	"fmt"
	"strings"
)

// Provide autocompletion for commands and resources.
type Completer struct {
	commands  []string
	resources []string
}

// Create a new autocompletion system.
func NewCompleter() *Completer {
	return &Completer{
		commands: []string{
			"help", "banter", "force", "transform", "voice", "mood",
			"status", "capabilities", "config", "debug", "quit",
		},
		resources: []string{
			"terminal", "screen", "code", "git", "system", "network",
			"files", "process", "memory", "clipboard", "audio", "video",
		},
	}
}

// Provide autocompletion suggestions.
func (c *Completer) Complete(input string) []string {
	if strings.HasPrefix(input, "/") {
		return c.completeCommand(input[1:])
	}

	if strings.HasPrefix(input, "#") {
		return c.completeResource(input[1:])
	}

	return []string{}
}

// Provide command completions for partial input
func (c *Completer) completeCommand(partial string) []string {
	var matches []string

	for _, cmd := range c.commands {
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, "/"+cmd)
		}
	}

	return matches
}

// Provide resource completions for partial input
func (c *Completer) completeResource(partial string) []string {
	var matches []string

	for _, resource := range c.resources {
		if strings.HasPrefix(resource, partial) {
			matches = append(matches, "#"+resource)
		}
	}

	return matches
}

// Add a new command for completion
func (c *Completer) AddCommand(command string) {
	c.commands = append(c.commands, command)
}

// Add a new resource for completion
func (c *Completer) AddResource(resource string) {
	c.resources = append(c.resources, resource)
}

// Set up built-in commands for the CLI
func (cli *Interface) registerDefaultCommands() {
	cli.commands["help"] = &HelpCommand{}
	cli.commands["banter"] = &BanterCommand{}
	cli.commands["force"] = &ForceCommand{}
	cli.commands["transform"] = &TransformCommand{}
	cli.commands["voice"] = &VoiceCommand{}
	cli.commands["mood"] = &MoodCommand{}
	cli.commands["status"] = &StatusCommand{}
	cli.commands["quit"] = &QuitCommand{}
}

// Show available commands and help info
type HelpCommand struct{}

func (h *HelpCommand) Name() string        { return "help" }
func (h *HelpCommand) Description() string { return "Show available commands" }
func (h *HelpCommand) Usage() string       { return "/help [command]" }

func (h *HelpCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) > 0 {
		// Show help for a specific command
		cmd, exists := ctx.Interface.commands[args[0]]
		if !exists {
			return fmt.Errorf("unknown command: %s", args[0])
		}

		fmt.Printf("Command: %s\n", cmd.Name())
		fmt.Printf("Description: %s\n", cmd.Description())
		fmt.Printf("Usage: %s\n", cmd.Usage())
		return nil
	}

	// Show all commands
	fmt.Println("Available commands:")
	for name, cmd := range ctx.Interface.commands {
		fmt.Printf("  /%s - %s\n", name, cmd.Description())
	}

	fmt.Println("\nAvailable resources:")
	for _, resource := range ctx.Interface.completer.resources {
		fmt.Printf("  #%s\n", resource)
	}

	return nil
}

// Enable casual conversation mode
type BanterCommand struct{}

func (b *BanterCommand) Name() string        { return "banter" }
func (b *BanterCommand) Description() string { return "Switch to casual conversation mode" }
func (b *BanterCommand) Usage() string       { return "/banter" }

func (b *BanterCommand) Execute(args []string, ctx *CommandContext) error {
	ctx.Interface.behavior.UpdateMood("happy", 0.8)
	fmt.Println("😊 *perks up* Oh, you want to chat? Well... I suppose I have time...")
	return nil
}

// Bypass Nero's usual personality filters for direct responses
type ForceCommand struct{}

func (f *ForceCommand) Name() string        { return "force" }
func (f *ForceCommand) Description() string { return "Force direct responses (bypasses personality)" }
func (f *ForceCommand) Usage() string       { return "/force <message>" }

func (f *ForceCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		return fmt.Errorf("force command requires a message")
	}

	message := strings.Join(args, " ")
	fmt.Printf("💪 [DIRECT MODE] Processing: %s\n", message)

	// Process without personality filters
	return nil
}

// Manipulate text according to user instructions
type TransformCommand struct{}

func (t *TransformCommand) Name() string        { return "transform" }
func (t *TransformCommand) Description() string { return "Transform text according to instructions" }
func (t *TransformCommand) Usage() string       { return "/transform <text> <instruction>" }

func (t *TransformCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /transform <text> <instruction>")
	}

	text := args[0]
	instruction := strings.Join(args[1:], " ")

	// Pass text and instruction to the model for transformation
	fmt.Printf("[transform] Would apply instruction '%s' to text: %s\n", instruction, text)
	return nil
}

// Change voice settings
type VoiceCommand struct{}

func (v *VoiceCommand) Name() string        { return "voice" }
func (v *VoiceCommand) Description() string { return "Change voice settings" }
func (v *VoiceCommand) Usage() string       { return "/voice <type>" }

func (v *VoiceCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		fmt.Println("Available voices: soft, energetic, sultry, robotic")
		return nil
	}

	voice := args[0]
	fmt.Printf("🔊 *voice changes* Switching to %s voice...\n", voice)

	// Implement voice switching
	return nil
}

// Set Nero's mood manually
type MoodCommand struct{}

func (m *MoodCommand) Name() string        { return "mood" }
func (m *MoodCommand) Description() string { return "Set Nero's mood" }
func (m *MoodCommand) Usage() string       { return "/mood <mood>" }

func (m *MoodCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		state := ctx.Interface.behavior.GetState()
		fmt.Printf("Current mood: %s (intensity: %.1f)\n",
			state.Mood.Primary, state.Mood.Intensity)
		return nil
	}

	mood := args[0]
	ctx.Interface.behavior.UpdateMood(mood, 0.7)
	fmt.Printf("💭 *mood shifts* Now feeling %s...\n", mood)

	return nil
}

// Show system status
type StatusCommand struct{}

func (s *StatusCommand) Name() string        { return "status" }
func (s *StatusCommand) Description() string { return "Show system status" }
func (s *StatusCommand) Usage() string       { return "/status" }

func (s *StatusCommand) Execute(args []string, ctx *CommandContext) error {
	state := ctx.Interface.behavior.GetState()

	fmt.Println("🔍 Nero System Status:")
	fmt.Printf("  Personality: %s\n", state.Personality.Name)
	fmt.Printf("  Mood: %s (%.1f)\n", state.Mood.Primary, state.Mood.Intensity)
	fmt.Printf("  Energy: %.1f%%\n", state.Energy*100)
	fmt.Printf("  Confidence: %.1f%%\n", state.Confidence*100)
	fmt.Printf("  Engagement: %.1f%%\n", state.Engagement*100)

	// Show capability status, memory usage, etc.

	return nil
}

// Exit the application
type QuitCommand struct{}

func (q *QuitCommand) Name() string        { return "quit" }
func (q *QuitCommand) Description() string { return "Exit Nero" }
func (q *QuitCommand) Usage() string       { return "/quit" }

func (q *QuitCommand) Execute(args []string, ctx *CommandContext) error {
	fmt.Println("👋 *waves wing* Goodbye!")
	// Trigger graceful shutdown
	return nil
}
