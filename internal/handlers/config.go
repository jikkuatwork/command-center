package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jikku/command-center/internal/config"
)

// ConfigHandler returns the current configuration (sanitized)
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.Get()

	// Create sanitized config (no password hash)
	sanitized := map[string]interface{}{
		"server": map[string]interface{}{
			"port":   cfg.Server.Port,
			"domain": cfg.Server.Domain,
			"env":    cfg.Server.Env,
		},
		"database": map[string]interface{}{
			"path": cfg.Database.Path,
		},
		"auth": map[string]interface{}{
			"username": cfg.Auth.Username,
			// Never expose password hash
			// enabled field removed in v0.4.0 - auth is always required
		},
		"ntfy": map[string]interface{}{
			"topic": cfg.Ntfy.Topic,
			"url":   cfg.Ntfy.URL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitized)
}
