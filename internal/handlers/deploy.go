package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jikku/command-center/internal/database"
	"github.com/jikku/command-center/internal/hosting"
)

// DeployHandler handles site deployments via ZIP upload
// POST /api/deploy
// - Multipart form with "file" (ZIP) and "site_name" field
// - Authorization: Bearer <token> header required
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate API key
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		jsonError(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		jsonError(w, "Invalid Authorization format, use: Bearer <token>", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	keyID, keyName, err := hosting.ValidateAPIKey(db, token)
	if err != nil {
		jsonError(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get site name
	siteName := r.FormValue("site_name")
	if siteName == "" {
		jsonError(w, "Missing site_name field", http.StatusBadRequest)
		return
	}

	// Validate site name
	if err := hosting.ValidateSubdomain(siteName); err != nil {
		jsonError(w, "Invalid site_name: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "Missing or invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify it's a ZIP file
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		jsonError(w, "File must be a ZIP archive", http.StatusBadRequest)
		return
	}

	// Read file into memory (we need to seek for zip.Reader)
	var buf bytes.Buffer
	size, err := io.Copy(&buf, file)
	if err != nil {
		jsonError(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), size)
	if err != nil {
		jsonError(w, "Invalid ZIP file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Deploy the site
	result, err := hosting.DeploySite(zipReader, siteName)
	if err != nil {
		jsonError(w, "Deployment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record deployment
	deployedBy := keyName
	if err := hosting.RecordDeployment(db, result.SiteID, result.SizeBytes, result.FileCount, deployedBy); err != nil {
		log.Printf("Failed to record deployment: %v", err)
	}

	log.Printf("Site deployed: %s by %s (key_id=%d), %d files, %d bytes",
		siteName, keyName, keyID, result.FileCount, result.SizeBytes)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"site":       siteName,
		"file_count": result.FileCount,
		"size_bytes": result.SizeBytes,
		"message":    "Deployment successful",
	})
}

// jsonError sends a JSON error response
func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
