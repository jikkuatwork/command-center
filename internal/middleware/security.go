package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/jikku/command-center/internal/config"
)

// MaxBodySize is the default maximum request body size (1MB)
const MaxBodySize = 1 << 20 // 1MB

// BodySizeLimit limits the size of request bodies to prevent memory exhaustion
func BodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for paths that have their own limits (deploy has 100MB)
			if r.URL.Path == "/api/deploy" {
				next.ServeHTTP(w, r)
				return
			}

			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestTracing adds a unique request ID header for tracing
func RequestTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from load balancer)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set on response for client visibility
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context (can be used in handlers)
		r.Header.Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a short random hex string
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()

		// Basic security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data: https://cdn.jsdelivr.net; " +
			"connect-src 'self'"

		w.Header().Set("Content-Security-Policy", csp)

		// HSTS in production
		if cfg.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
