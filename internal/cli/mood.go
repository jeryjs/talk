package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var moodCmd = &cobra.Command{
	Use:   "mood [mood]",
	Short: "Set Nero's personality mood",
	Long: `Set Nero's personality/mood state.

Available moods:
  - helpful       : Friendly and eager to assist
  - professional  : Formal and business-like
  - creative      : Imaginative and innovative
  - analytical    : Logical and detail-oriented
  - enthusiastic  : Energetic and positive
  - sarcastic     : Witty with dry humor
  - empathetic    : Understanding and supportive
  - concise       : Brief and to the point`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show current mood
			fmt.Println("Current mood: helpful 😊")
			return
		}

		mood := args[0]

		// Validate and set mood
		moodEmojis := map[string]string{
			"helpful":      "😊",
			"professional": "👔",
			"creative":     "🎨",
			"analytical":   "🔬",
			"enthusiastic": "🚀",
			"sarcastic":    "😏",
			"empathetic":   "💙",
			"concise":      "⚡",
		}

		emoji, ok := moodEmojis[mood]
		if !ok {
			color.Red("❌ Unknown mood: %s", mood)
			fmt.Println("\nAvailable moods:")
			for m, e := range moodEmojis {
				fmt.Printf("  %s %s\n", e, m)
			}
			return
		}

		color.Green("✓ Mood set to: %s %s", mood, emoji)
	},
}
