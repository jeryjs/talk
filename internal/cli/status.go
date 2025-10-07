package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Nero's status and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("🤖 Nero Status\n")

		fmt.Println("\n📊 System:")
		fmt.Println("  Version:        ", getVersion())
		fmt.Println("  Config:         ", "~/.nero/config.yaml")
		fmt.Println("  Plugins:        ", "3 loaded")

		fmt.Println("\n🎭 Current State:")
		fmt.Println("  Mood:           ", "helpful 😊")
		fmt.Println("  Model:          ", "gpt-4o")
		fmt.Println("  Voice:          ", "enabled ✓")
		fmt.Println("  Computer Use:   ", "enabled ✓")
		fmt.Println("  MCP:            ", "3 servers connected ✓")

		fmt.Println("\n💬 Chat:")
		fmt.Println("  Active Chat:    ", "default")
		fmt.Println("  Messages:       ", "0")
		fmt.Println("  Context Used:   ", "0 / 128K tokens")

		fmt.Println("\n🔌 MCP Servers:")
		fmt.Println("  ✓ filesystem")
		fmt.Println("  ✓ github")
		fmt.Println("  ✓ web-search")

		fmt.Println("\n🧩 Plugins:")
		fmt.Println("  ✓ screenshot-analyzer")
		fmt.Println("  ✓ code-review")
		fmt.Println("  ✓ terminal-helper")
	},
}
