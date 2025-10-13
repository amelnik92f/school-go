package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"schools-be/internal/config"
)

// APIKeyAuth is a middleware that validates API key authentication
func APIKeyAuth(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If API key is not configured, skip authentication (for development)
			if cfg.APIKey == "" {
				slog.Warn("API key authentication is disabled - no API_KEY configured")
				next.ServeHTTP(w, r)
				return
			}

			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Also check Authorization header with Bearer token format
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// Validate API key
			if apiKey == "" {
				slog.Warn("missing API key",
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.String("remote_addr", r.RemoteAddr),
				)
				respondError(w, http.StatusUnauthorized, "missing API key")
				return
			}

			if apiKey != cfg.APIKey {
				slog.Warn("invalid API key",
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.String("remote_addr", r.RemoteAddr),
				)
				respondError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			// API key is valid, proceed with request
			next.ServeHTTP(w, r)
		})
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
