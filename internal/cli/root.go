package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"nero/internal/agent"
	"nero/internal/config"
	"nero/internal/mood"
	"nero/internal/ui"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "nero [message]",
	Short: "Nero - Your AI-Powered CLI Assistant",
	Long: `Nero is a cutting-edge CLI assistant with:
  🤖 Agentic AI with heavy tool use
  🎭 Personality system with mood states
  📸 Computer use capabilities
  🗣️ Voice synthesis
  🔌 Full MCP server support
  🧩 Extendable plugin system
  💬 Multi-model support (OpenAI, Anthropic, Gemini)`,
	Version: getVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// Direct message mode
			runDirectMessage(args)
		} else {
			// Interactive mode
			runInteractive()
		}
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nero/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Local flags
	rootCmd.Flags().StringP("model", "m", "", "LLM model to use")
	rootCmd.Flags().StringP("mood", "", "helpful", "mood/personality mode")
	rootCmd.Flags().BoolP("no-voice", "", false, "disable voice synthesis")

	// Keep useful diagnostic commands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
	}
}

func getVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, CommitSHA, BuildDate)
}

func mustCreateAgent() *agent.Agent {
	cfg := config.Get()

	ag, err := agent.New()
	if err != nil {
		color.Red("Failed to create agent: %v", err)
		os.Exit(1)
	}

	// Set mood
	moodPrompt := mood.GetSystemPrompt(cfg.Mood)
	ag.SetSystemPrompt(moodPrompt)

	return ag
}

func runDirectMessage(args []string) {
	// Import packages
	agent := mustCreateAgent()
	message := strings.Join(args, " ")

	fmt.Printf("You: %s\n\n", message)

	response, err := agent.Chat(context.Background(), message)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	color.Cyan("Nero: ")
	fmt.Println(response)
}

func runInteractive() {
	// Import the ui package
	if err := ui.Run(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}
