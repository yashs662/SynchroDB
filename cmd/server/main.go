package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/pkg/protocol"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println("FATAL: " + err.Error())
		os.Exit(1)
	}

	// Initialize Logger
	logger.Init(config)

	// Channel to listen for OS interrupt signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start server in a goroutine
	go func() {
		// Start the protocol server
		if err := protocol.StartServer(config); err != nil {
			logger.Fatal("Failed to start the server: " + err.Error())
			os.Exit(1)
		}
	}()

	// Block until we receive a signal in stop channel
	<-stop
	logger.Info("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := protocol.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}
	logger.Info("Server exiting gracefully")
}
