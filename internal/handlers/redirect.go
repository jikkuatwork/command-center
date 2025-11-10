package handlers

import (
	"net/http"
)

// RedirectHandler handles redirect tracking (to be implemented in Phase 4)
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder - will be implemented in Phase 4
	http.Error(w, "Not found", http.StatusNotFound)
}
