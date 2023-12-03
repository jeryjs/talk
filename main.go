package main

import (
	"bufio"
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
	for msg != "exit" {
		msg = su.Listen()
		// text := su.ChatWithGPT(msg)
		text := su.ChatWithBard(msg)
		go su.SayWithEspeak(text) // Robotic Voice
		// go su.SayWithTTS(text)			// Male Voice
		// go su.SayWithHtgoTts(text)		// Female Voice
	}
	color.HiRed("\nPress Enter to exit...")
	fmt.Scanln()
}
