package handlers

import (
	"net/http"
)

// WebhookHandler handles incoming webhooks (to be implemented in Phase 5)
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Placeholder - will be implemented in Phase 5
	w.WriteHeader(http.StatusOK)
}
