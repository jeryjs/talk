package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"nero/cli"
	"nero/kernel"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Initialize Nero kernel
	runtime := kernel.NewRuntime()

	// Start CLI interface
	cli := cli.NewInterface(runtime)
	if cli == nil {
		panic("Failed to initialize CLI - no AI provider available")
	}

	// Boot the system
	if err := runtime.Boot(ctx); err != nil {
		panic(err)
	}

	// Start interactive session
	cli.Start(ctx)
}
