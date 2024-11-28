package api

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/internal/stores"
	"golang.org/x/time/rate"
)

type contextKey string

const userContextKey contextKey = "user"

var limiter = rate.NewLimiter(1, 5) // 1 request per second with a burst of 5

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

func rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
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

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.JwtSecret), nil // Use the secret from the struct
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate claims
		if claims.Issuer != "your-issuer" || !containsAudience(claims.Audience, "your-audience") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract the username from the token claims
		username := claims.Subject
		user, err := h.Store.Credentials.FindUserByUsername(username)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add the user to the request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
		// Consider adding rate limiting here
	})
}

func containsAudience(audience jwt.ClaimStrings, target string) bool {
	for _, aud := range audience {
		if aud == target {
			return true
		}
	}
	return false
}

func GetUserFromContext(ctx context.Context) (*stores.User, bool) {
	user, ok := ctx.Value(userContextKey).(*stores.User)
	return user, ok
}
