package main

import (
	"bufio"
	"fmt"
	"os"
	su "talk/speechUtils"
)

func scan() string {
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {return scanner.Text()}
    if err := scanner.Err(); err != nil {fmt.Fprintln(os.Stderr, err)}
    return ""
}

func main() {
	var msg string
	for(msg != "exit") {
		msg = su.Listen()
		text := su.Chat(msg)
		su.SayWithEspeak(text)
	}
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}