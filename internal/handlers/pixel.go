package handlers

import (
	"net/http"
)

// PixelHandler serves a 1x1 transparent GIF for tracking (to be implemented in Phase 4)
func PixelHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder - will be implemented in Phase 4
	w.WriteHeader(http.StatusOK)
}
