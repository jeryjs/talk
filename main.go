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
	ai := flag.String("ai", "gpt", "Chat AI to use (gpt/gemini/liberty)")
	speech := flag.String("se", "tts", "Speech engine to use (espeak/tts/htgotts)")

	// Parse command line arguments
	flag.Parse()

	msg = *text
	currentAI := *ai
	currentSpeech := *speech

	for msg != "exit" {
		if msg == "" {
			msg = su.Listen()
		}
		var text string

		// Helper function for setting AI and printing messages
		setAIPrefix := func(aiType, prefix string) {
			if strings.HasPrefix(msg, prefix) {
				currentAI = aiType
				msg = strings.TrimPrefix(msg, prefix)
				color.New(color.FgMagenta).Printf("Switching to %s AI\n", aiType)
			}
		}
		// Determine chat engine
		setAIPrefix("gemini", "<b")
		setAIPrefix("gpt", "<g")
		setAIPrefix("liberty", "<l")

		// Helper function for setting speech engine and printing messages
		setSpeechPrefix := func(speechType, prefix string) {
			if strings.HasPrefix(msg, prefix) {
				currentSpeech = speechType
				msg = strings.TrimPrefix(msg, prefix)
				color.New(color.FgMagenta).Printf("Switching to %s speech\n", speechType)
			}
		}
		// Determine speech engine
		setSpeechPrefix("espeak", "<1")
		setSpeechPrefix("tts", "<2")
		setSpeechPrefix("htgotts", "<3")

		// Select chat engine based on flag
		switch currentAI {
			case "gpt": text = su.ChatWithGPT(msg)
			case "gemini": text = su.ChatWithGemini(msg)
			case "liberty": text = su.ChatWithLiberty(msg)
		}

		// Select speech engine based on flag
		switch currentSpeech {
			case "espeak": go su.SayWithEspeak(text) // Robotic Voice
			case "tts": go su.SayWithTTS(text) // Male Voice
			case "htgotts": go su.SayWithHtgoTts(text) // Female Voice
		}

		// Reset the msg variable
		msg = ""
	}

	color.HiRed("\nPress Enter to exit...")
	fmt.Scanln()
}
