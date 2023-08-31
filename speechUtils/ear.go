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
	fmt.Print(len(chatHistory)-InitialHistoryLength,">\t")

	sc := bufio.NewScanner(os.Stdin)
	color.Set(color.FgHiYellow)		// set console fg to Yellow
	
	var input string
	for sc.Scan() {
        line := sc.Text()
        if line == "" {break}
        input += line + "\n"
    }

	color.Set(color.FgHiRed)
	return input
}