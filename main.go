package main

import (
	"fmt"
	"os"

	"nero/internal/cli"
)

// Version information (set during build)
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Set version information
	cli.Version = Version
	cli.CommitSHA = CommitSHA
	cli.BuildDate = BuildDate

	// Execute the CLI
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
