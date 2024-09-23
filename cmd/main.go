package main

import (
	"log"
	"net/http"

	"github.com/yashs662/SynchroDB/internal/api"
	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/kv_store"
	"github.com/yashs662/SynchroDB/internal/logger"
)

func main() {
	// Initialize logger
	logger.Init(false)

	// Parse command-line flags
	config := config.ParseFlags()

	logger.Info("Starting SynchroDB...")

	store := kv_store.NewStore()

	logger.Info("Store initialized")

	// Initialize API handlers
	handlers := api.NewHandlers(store)

	// Set up the HTTP routes
	handlers.SetupRoutes()

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
