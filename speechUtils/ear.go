package speechUtils

import (
	"bufio"
	"fmt"
	"os"
)

// Implement Speech-To-Text using OpenAI's Whisper API.
// TODO: implement STT using Whisper API
func Listen() string {
	fmt.Print(">\t")
	fmt.Print("Listening...\t")
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return sc.Text()
}