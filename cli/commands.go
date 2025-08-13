package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Show available commands
type HelpCommand struct{}

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Description() string { return "Show available commands" }
func (c *HelpCommand) Usage() string       { return "/help" }

func (c *HelpCommand) Execute(args []string, ctx *CommandContext) error {
	help := `
🎭 Available Commands:

🗣️  Chat Commands:
  /mood <state>     Set emotional state (happy, sad, excited, grumpy)
  /status           Show system and mood status

💻 System Commands:
  /run <command>    Execute system command or script
  /open <app>       Open application (notepad, calculator, etc.)

🏃 Control:
  /exit, /quit      Exit gracefully
  /help             Show this help

💡 Pro Tips:
  • Talk naturally - Nero understands context!
  • Use #terminal, #screen, #code for resource access
  • Nero remembers your conversations across sessions

Example: "Nero, please help me open VS Code" or "/run git status"
`
	color.New(color.FgCyan).Print(help)
	return nil
}

// Show system status
type StatusCommand struct{}

func (c *StatusCommand) Name() string        { return "status" }
func (c *StatusCommand) Description() string { return "Show system status" }
func (c *StatusCommand) Usage() string       { return "/status" }

func (c *StatusCommand) Execute(args []string, ctx *CommandContext) error {
	status := `
📊 System Status:

🧠 Memory: Active and persistent
🎭 Behavioral Engine: Operational
🤖 AI Provider: Connected
💾 Session Storage: Available

Nero Status: Ready to assist (with attitude!) 💜
`
	color.New(color.FgGreen).Print(status)
	return nil
}

// Change Nero's mood
type MoodCommand struct{}

func (c *MoodCommand) Name() string        { return "mood" }
func (c *MoodCommand) Description() string { return "Set Nero's emotional state" }
func (c *MoodCommand) Usage() string       { return "/mood <happy|sad|excited|grumpy|confident>" }

func (c *MoodCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify a mood: happy, sad, excited, grumpy, confident")
	}

	mood := strings.ToLower(args[0])

	// Update behavioral engine mood
	ctx.Interface.behavior.UpdateMood(mood, 0.8)

	emoji := getMoodEmoji(mood)
	color.New(color.FgMagenta).Printf("💫 Mood updated to: %s %s\n", mood, emoji)
	color.New(color.FgMagenta).Printf("Nero: *adjusts mood* Fine... if that's what you want... 😤\n")

	return nil
}

// Execute system commands
type RunCommand struct{}

func (c *RunCommand) Name() string        { return "run" }
func (c *RunCommand) Description() string { return "Execute system command" }
func (c *RunCommand) Usage() string       { return "/run <command> [args...]" }

func (c *RunCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify a command to run")
	}

	color.New(color.FgYellow).Printf("🔧 Executing: %s\n", strings.Join(args, " "))
	color.New(color.FgMagenta).Println("Nero: *reluctantly* Fine, I'll run your command... but don't blame me if something breaks! 😤")

	// Run command through system provider
	output, err := ctx.Interface.systemProvider.RunCommand(args[0], args[1:]...)
	if err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		color.New(color.FgCyan).Printf("Output:\n%s\n", output)
	} else {
		color.New(color.FgGreen).Println("✅ Command executed successfully")
	}

	return nil
}

// Open applications
type OpenCommand struct{}

func (c *OpenCommand) Name() string        { return "open" }
func (c *OpenCommand) Description() string { return "Open application" }
func (c *OpenCommand) Usage() string       { return "/open <app>" }

func (c *OpenCommand) Execute(args []string, ctx *CommandContext) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify an application to open")
	}

	app := strings.Join(args, " ")

	color.New(color.FgYellow).Printf("🚀 Opening: %s\n", app)
	color.New(color.FgMagenta).Println("Nero: *sighs* There... your precious application is starting. You're welcome! 💜")

	err := ctx.Interface.systemProvider.OpenApp(app)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", app, err)
	}

	color.New(color.FgGreen).Printf("✅ %s opened successfully\n", app)
	return nil
}

// Exit the application
type ExitCommand struct{}

func (c *ExitCommand) Name() string        { return "exit" }
func (c *ExitCommand) Description() string { return "Exit Nero" }
func (c *ExitCommand) Usage() string       { return "/exit or /quit" }

func (c *ExitCommand) Execute(args []string, ctx *CommandContext) error {
	color.New(color.FgMagenta).Println("\n💜 Nero: *pouts* F-Fine! But... you better come back soon, okay?")
	color.New(color.FgCyan).Println("Shutting down gracefully...")
	os.Exit(0)
	return nil
}

// Register all default commands
func (cli *Interface) registerDefaultCommands() {
	commands := []Command{
		&HelpCommand{},
		&StatusCommand{},
		&MoodCommand{},
		&RunCommand{},
		&OpenCommand{},
		&ExitCommand{},
	}

	for _, cmd := range commands {
		cli.commands[cmd.Name()] = cmd

		// Register aliases for exit command
		if cmd.Name() == "exit" {
			cli.commands["quit"] = cmd
		}
	}
}
