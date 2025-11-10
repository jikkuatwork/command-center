package handlers

import (
	"net/http"
)

// TrackHandler handles tracking requests (to be implemented in Phase 3)
func TrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Placeholder - will be implemented in Phase 3
	w.WriteHeader(http.StatusNoContent)
}
