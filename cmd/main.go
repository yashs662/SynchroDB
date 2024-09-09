package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/yashs662/SynchroDB/internal/kv_store"
	"github.com/yashs662/SynchroDB/internal/logger"
)

func main() {
	// Parse command-line flags
	debug := flag.Bool("debug", false, "enable debug mode with detailed logging")
	flag.BoolVar(debug, "d", false, "enable debug mode with detailed logging (shorthand)")
	flag.Parse()

	// Initialize logger with the debug flag
	logger.Init(*debug)

	if *debug {
		logger.Debug("Debug mode enabled")
	}

	logger.Info("Starting SynchroDB...")

	store := kv_store.NewStore()

	logger.Info("Store initialized")

	// Define HTTP routes
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value, exists := store.Get(key)
		if !exists {
			logger.Warnf("Key not found: %s", key)
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		logger.Infof("Retrieved key %s with value %s", key, value)
		fmt.Fprintf(w, "Value: %s\n", value)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		store.Set(key, value)
		logger.Infof("Set key %s to value %s", key, value)
		fmt.Fprintf(w, "Set key %s to value %s\n", key, value)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		store.Delete(key)
		logger.Infof("Deleted key %s", key)
		fmt.Fprintf(w, "Deleted key %s\n", key)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
