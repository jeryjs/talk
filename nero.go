package main

import (
	"log"

	"nero/cli"
)

func main() {
	// Initialize CLI with Bubble Tea
	repl := cli.NewREPL()

	// Start the Bubble Tea REPL - this will handle everything
	if err := repl.Start(); err != nil {
		log.Fatal("Failed to start REPL:", err)
	}
}
