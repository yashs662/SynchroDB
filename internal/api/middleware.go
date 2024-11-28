package api

import (
	"net"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/yashs662/SynchroDB/internal/logger"
)

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

func (h *Handlers) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Bearer <token>
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		claims := &jwt.StandardClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.JwtSecret), nil // Use the secret from the struct
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
