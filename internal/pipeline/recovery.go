package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler creates a cancellation context that cancels on SIGINT or SIGTERM.
func SetupSignalHandler(onInterrupt func()) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		if onInterrupt != nil {
			onInterrupt()
		}
		cancel()
	}()

	return ctx
}

// SafeGo runs a function in a goroutine with panic recovery.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[RECOVERED PANIC] %v\n", r)
			}
		}()
		fn()
	}()
}
