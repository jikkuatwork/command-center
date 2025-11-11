package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/jikku/command-center/internal/auth"
	"github.com/jikku/command-center/internal/config"
)

// AuthMiddleware checks if a user is authenticated before allowing access to protected routes
func AuthMiddleware(sessionStore *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := config.Get()

			// If auth is disabled, allow all requests
			if !cfg.Auth.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check if the path requires authentication
			if !requiresAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get session cookie
			sessionID, err := auth.GetSessionCookie(r)
			if err != nil {
				// No session cookie, redirect to login
				log.Printf("No session cookie for %s %s", r.Method, r.URL.Path)
				redirectToLogin(w, r)
				return
			}

			// Validate session
			valid, err := sessionStore.ValidateSession(sessionID)
			if err != nil {
				log.Printf("Session validation error: %v", err)
				redirectToLogin(w, r)
				return
			}

			if !valid {
				log.Printf("Invalid or expired session for %s %s", r.Method, r.URL.Path)
				redirectToLogin(w, r)
				return
			}

			// Session is valid, allow request
			next.ServeHTTP(w, r)
		})
	}
}

// requiresAuth returns true if the path requires authentication
func requiresAuth(path string) bool {
	// Public endpoints (no auth required)
	publicPaths := []string{
		"/track",
		"/pixel.gif",
		"/r/",
		"/webhook/",
		"/static/",
		"/login",
		"/api/login",
		"/health",
	}

	// Check if path matches any public path
	for _, public := range publicPaths {
		if path == public || strings.HasPrefix(path, public) {
			return false
		}
	}

	// All other paths require authentication
	return true
}

// redirectToLogin redirects the user to the login page
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// For API requests, return 401 Unauthorized
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// For HTML requests, redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
