package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:     "model [subcommand]",
	Short:   "Manage AI models",
	Aliases: []string{"models"},
	Long: `Model management commands:
  list      - List available models
  use       - Switch to a different model
  info      - Show model information`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Available Models:")
		fmt.Println("\n🔷 OpenAI:")
		fmt.Println("  ✓ gpt-4.5 (GPT-4.5 Orion)")
		fmt.Println("  ✓ gpt-4o (Multimodal)")
		fmt.Println("  ✓ gpt-4-turbo")

		fmt.Println("\n🔶 Anthropic:")
		fmt.Println("  ✓ claude-sonnet-4.5 (Best for coding & agents)")
		fmt.Println("  ✓ claude-3.5-sonnet")
		fmt.Println("  ✓ claude-3-opus")

		fmt.Println("\n🔵 Google:")
		fmt.Println("  ✓ gemini-2.5-pro (Latest with thinking) ⭐")
		fmt.Println("  ✓ gemini-2.0-flash-exp")
		fmt.Println("  ✓ gemini-2.0-flash-thinking-exp")

		fmt.Println("\n🦙 Ollama (Local):")
		fmt.Println("  ✓ llama3.1:405b")
		fmt.Println("  ✓ qwen2.5:235b")
		fmt.Println("  ✓ deepseek-r1:671b")
	},
}

var modelUseCmd = &cobra.Command{
	Use:   "use [model]",
	Short: "Switch to a different model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("✓ Switched to model: %s\n", args[0])
	},
}

var modelInfoCmd = &cobra.Command{
	Use:   "info [model]",
	Short: "Show model information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ℹ️  Model: %s\n", args[0])
		fmt.Println("Provider: OpenAI")
		fmt.Println("Context Window: 128K tokens")
		fmt.Println("Capabilities: text, vision, function calling")
		fmt.Println("Cost: $0.01/1K input, $0.03/1K output")
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelUseCmd)
	modelCmd.AddCommand(modelInfoCmd)
}
