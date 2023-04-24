package speechUtils

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Implement Speech-To-Text using OpenAI's Whisper API.
// TODO: implement STT using Whisper API
func Listen() string {
	color.Set(color.FgHiGreen)
	fmt.Print(">\t")
	// fmt.Print("Listening...\t")
	sc := bufio.NewScanner(os.Stdin)
	color.Set(color.FgHiYellow)		// set console fg to Yellow
	sc.Scan()
	color.Unset()
	return sc.Text()
}