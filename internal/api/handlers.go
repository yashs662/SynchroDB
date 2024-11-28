package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/internal/stores"
)

type Handlers struct {
	Store     *stores.KVStore
	JwtSecret string
}

func NewHandlers(store *stores.KVStore, jwtSecret string) *Handlers {
	return &Handlers{
		Store:     store,
		JwtSecret: jwtSecret,
	}
}

func (h *Handlers) SetupRoutes() {
	http.Handle("/get", loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Get))))
	http.Handle("/set", loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Set))))
	http.Handle("/delete", loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Delete))))
	http.Handle("/login", loggingMiddleware(http.HandlerFunc(h.Login)))
}

func (h *Handlers) getParam(r *http.Request, key string) (string, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", fmt.Errorf("missing parameter: %s", key)
	}
	return value, nil
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// hash the password
	// compare with stored hash
	// if match, generate token
	// return token

	hashedPassword, exists := h.Store.Get(creds.Username)
	if !exists {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if hashedPassword != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(h.JwtSecret))
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
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
