package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/pkg/database"
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

	store := database.NewKVStore()
	var aofWriter *database.AOFWriter
	if config.Server.PersistentAOFPath != "" {
		var err error
		aofWriter, err = database.NewAOFWriter(config.Server.PersistentAOFPath)
		if err != nil {
			logger.Fatal("Failed to create AOF writer: " + err.Error())
			os.Exit(1)
		}
	}

	server := protocol.NewServer(config, store, aofWriter)

	// Channel to listen for OS interrupt signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start server in a goroutine
	go func() {
		if err := server.Start(config); err != nil {
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
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}
	logger.Info("Server exiting gracefully")
	logger.Close()
}
