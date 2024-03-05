package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	su "talk/speechUtils"

	"github.com/fatih/color"
)

func scan() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return ""
}

func main() {
	var msg string

	// Define flags
	ai := flag.String("ai", "gpt", "Chat AI to use (gpt/bard)")
	speech := flag.String("s", "espeak", "Speech engine to use (espeak/tts/htgotts)")

	// Parse command line arguments
	flag.Parse()

	for msg != "exit" {
		msg = su.Listen()
		var text string

		// Select chat engine based on flag
		switch *ai {
		case "gpt":
			text = su.ChatWithGPT(msg)
		case "bard":
			text = su.ChatWithBard(msg)
		}

		// Select speech engine based on flag
		switch *speech {
		case "espeak":
			go su.SayWithEspeak(text) // Robotic Voice
		case "tts":
			go su.SayWithTTS(text) // Male Voice
		case "htgotts":
			go su.SayWithHtgoTts(text) // Female Voice
		}
	}

	color.HiRed("\nPress Enter to exit...")
	fmt.Scanln()
}
