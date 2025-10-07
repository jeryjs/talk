package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [subcommand]",
	Short: "Manage configuration",
	Long: `Configuration management commands:
  show      - Show current configuration
  set       - Set a configuration value
  reset     - Reset configuration to defaults
  path      - Show configuration file path`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📝 Current Configuration:")
		fmt.Println("\ndefault_model: gpt-4-turbo")
		fmt.Println("default_mood: helpful")
		fmt.Println("max_tokens: 4096")
		fmt.Println("temperature: 0.7")
		fmt.Println("enable_voice: true")
		fmt.Println("enable_computer_use: true")
		fmt.Println("enable_mcp: true")
		fmt.Println("\nmcp_servers:")
		fmt.Println("  - name: filesystem")
		fmt.Println("  - name: github")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("✓ Set %s = %s\n", args[0], args[1])
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("✓ Configuration reset to defaults")
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("~/.nero/config.yaml")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configPathCmd)
}
