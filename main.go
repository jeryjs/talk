package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
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
	text := flag.String("t", "", "Text to say")
	ai := flag.String("ai", "gemini", "Chat AI to use (gpt/gemini/liberty)")
	speech := flag.String("se", "tts", "Speech engine to use (espeak/tts/htgotts)")

	// Parse command line arguments
	flag.Parse()

	msg = *text
	for msg != "exit" {
		if msg == "" {
			msg = su.Listen()
		}
		var text string

		var ce string
		if strings.HasPrefix(msg, "<b") {
			ce = "gemini"
			msg = strings.TrimPrefix(msg, "<b")
		} else if strings.HasPrefix(msg, "<g") {
			ce = "gpt"
			msg = strings.TrimPrefix(msg, "<g")
		} else if strings.HasPrefix(msg, "<l") {
			ce = "liberty"
			msg = strings.TrimPrefix(msg, "<l")
		} else {
			ce = *ai
		}
		var se string
		if strings.HasPrefix(msg, "<1") {
			se = "espeak"
			msg = strings.TrimPrefix(msg, "<1")
		} else if strings.HasPrefix(msg, "<2") {
			se = "tts"
			msg = strings.TrimPrefix(msg, "<2")
		} else if strings.HasPrefix(msg, "<3") {
			se = "htgotts"
			msg = strings.TrimPrefix(msg, "<3")
		} else {
			se = *speech
		}

		// Select chat engine based on flag
		switch ce {
		case "gpt":
			text = su.ChatWithGPT(msg)
		case "gemini":
			text = su.ChatWithGemini(msg)
		case "liberty":
			text = su.ChatWithLiberty(msg)
		}

		// Select speech engine based on flag
		switch se {
		case "espeak":
			go su.SayWithEspeak(text) // Robotic Voice
		case "tts":
			go su.SayWithTTS(text) // Male Voice
		case "htgotts":
			go su.SayWithHtgoTts(text) // Female Voice
		}

		// Reset the msg variable
		msg = ""
	}

	color.HiRed("\nPress Enter to exit...")
	fmt.Scanln()
}
