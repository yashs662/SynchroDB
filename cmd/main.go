package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/yashs662/SynchroDB/internal/api"
	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/internal/stores"
)

func main() {
	// Initialize logger
	logger.Init(false)

	// Parse command-line flags
	config := config.ParseFlags()

	logger.Info("Loading User Data...")

	loadedCredentials, err := stores.LoadCredentials(config.EncryptionKey, config.CredentialFilePath)
	if err != nil {
		logger.Errorf("Error loading credentials: %v", err)
		os.Exit(1)
	} else {
		config.Credentials = *loadedCredentials
	}

	logger.Info("Starting SynchroDB...")

	// Initialize the key-value store
	store := stores.NewStore()

	logger.Info("Store initialized")

	// Initialize API handlers
	handlers := api.NewHandlers(store, config.JwtSecret)
	handlers.Store.Credentials = config.Credentials

	// Set up the HTTP routes
	handlers.SetupRoutes()

	// Create a new HTTP server instance
	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: http.DefaultServeMux,
	}

	// Channel to listen for OS interrupt signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start server in a goroutine
	go func() {
		logger.Infof("Server is listening on port %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Could not listen on %s: %v\n", config.Port, err)
		}
	}()

	// Block until we receive a signal in stop channel
	<-stop
	logger.Info("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting gracefully")
}
