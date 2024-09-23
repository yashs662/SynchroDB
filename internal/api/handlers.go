package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"

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
	http.Handle("/get", loggingMiddleware(http.HandlerFunc(h.Get)))
	http.Handle("/set", loggingMiddleware(http.HandlerFunc(h.Set)))
	http.Handle("/delete", loggingMiddleware(http.HandlerFunc(h.Delete)))
}

func (h *Handlers) getParam(r *http.Request, key string) (string, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", fmt.Errorf("missing parameter: %s", key)
	}
	return value, nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Check for X-Forwarded-For header for proxies
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ips := strings.Split(forwarded, ",")
			ip = strings.TrimSpace(ips[0])
		}

		logger.Infof("Request: %s %s from IP: %s", r.Method, r.URL.Path, ip)
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	key, err := h.getParam(r, "key")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	key, err := h.getParam(r, "key")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	value, err := h.getParam(r, "value")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Store.Set(key, value)
	logger.Infof("Set key %s to value %s", key, value)
	fmt.Fprintf(w, "Set key %s to value %s\n", key, value)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	key, err := h.getParam(r, "key")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Store.Delete(key)
	logger.Infof("Deleted key %s", key)
	fmt.Fprintf(w, "Deleted key %s\n", key)
}
