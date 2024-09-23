package api

import (
	"fmt"
	"net/http"

	"github.com/yashs662/SynchroDB/internal/kv_store"
	"github.com/yashs662/SynchroDB/internal/logger"
)

type Handlers struct {
	Store *kv_store.Store
}

func NewHandlers(store *kv_store.Store) *Handlers {
	return &Handlers{Store: store}
}

func (h *Handlers) SetupRoutes() {
	http.HandleFunc("/get", h.Get)
	http.HandleFunc("/set", h.Set)
	http.HandleFunc("/delete", h.Delete)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, exists := h.Store.Get(key)
	if !exists {
		logger.Warnf("Key not found: %s", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	logger.Infof("Retrieved key %s with value %s", key, value)
	fmt.Fprintf(w, "Value: %s\n", value)
}

func (h *Handlers) Set(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	h.Store.Set(key, value)
	logger.Infof("Set key %s to value %s", key, value)
	fmt.Fprintf(w, "Set key %s to value %s\n", key, value)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	h.Store.Delete(key)
	logger.Infof("Deleted key %s", key)
	fmt.Fprintf(w, "Deleted key %s\n", key)
}
