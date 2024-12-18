package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
	http.Handle("/get", rateLimitingMiddleware(loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Get)))))
	http.Handle("/set", rateLimitingMiddleware(loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Set)))))
	http.Handle("/delete", rateLimitingMiddleware(loggingMiddleware(h.JWTAuthMiddleware(http.HandlerFunc(h.Delete)))))
	http.Handle("/login", rateLimitingMiddleware(loggingMiddleware(http.HandlerFunc(h.Login))))
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

	user, err := h.Store.Credentials.FindUserByUsername(creds.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := stores.ComparePassword(user.HashedPassword, creds.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		Subject:   user.Username,
		Issuer:    "your-issuer",
		Audience:  jwt.ClaimStrings{"your-audience"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(h.JwtSecret))
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
	// Consider optimizing the login process here
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r.Context())
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
	logger.Infof("User <%s> retrieved key <%s> with value <%s>", user.Username, key, value)
}

func (h *Handlers) Set(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r.Context())
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
	logger.Infof("User <%s> set key <%s> to value <%s", user.Username, key, value)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r.Context())
	key, err := h.getParam(r, "key")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Store.Delete(key)
	logger.Infof("User <%s> deleted key <%s>", user.Username, key)
}
