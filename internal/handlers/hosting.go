package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jikku/command-center/internal/database"
	"github.com/jikku/command-center/internal/hosting"
)

// HostingPageHandler serves the hosting management page
func HostingPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/hosting.html")
}

// SitesHandler returns the list of hosted sites
func SitesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sites, err := hosting.ListSites()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sites":   sites,
	})
}

// APIKeysHandler handles API key CRUD operations
func APIKeysHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	switch r.Method {
	case http.MethodGet:
		// List API keys
		keys, err := hosting.ListAPIKeys(db)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"keys":    keys,
		})

	case http.MethodPost:
		// Create new API key
		var req struct {
			Name   string `json:"name"`
			Scopes string `json:"scopes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			jsonError(w, "Name is required", http.StatusBadRequest)
			return
		}

		token, err := hosting.CreateAPIKey(db, req.Name, req.Scopes)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"token":   token,
			"message": "API key created. Save this token - it won't be shown again!",
		})

	case http.MethodDelete:
		// Delete API key
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			jsonError(w, "ID parameter required", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			jsonError(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if err := hosting.DeleteAPIKey(db, id); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "API key revoked",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// EnvVarsHandler handles environment variables CRUD
func EnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	siteID := r.URL.Query().Get("site_id")

	switch r.Method {
	case http.MethodGet:
		if siteID == "" {
			jsonError(w, "site_id required", http.StatusBadRequest)
			return
		}
		rows, err := db.Query("SELECT id, name FROM env_vars WHERE site_id = ?", siteID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var vars []map[string]interface{}
		for rows.Next() {
			var id int64
			var name string
			if rows.Scan(&id, &name) == nil {
				vars = append(vars, map[string]interface{}{"id": id, "name": name})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "vars": vars})

	case http.MethodPost:
		var req struct {
			SiteID string `json:"site_id"`
			Name   string `json:"name"`
			Value  string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		_, err := db.Exec(`
			INSERT INTO env_vars (site_id, name, value) VALUES (?, ?, ?)
			ON CONFLICT(site_id, name) DO UPDATE SET value = ?
		`, req.SiteID, req.Name, req.Value, req.Value)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			jsonError(w, "id required", http.StatusBadRequest)
			return
		}
		id, _ := strconv.ParseInt(idStr, 10, 64)
		db.Exec("DELETE FROM env_vars WHERE id = ?", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeploymentsHandler returns recent deployments
func DeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT id, site_id, size_bytes, file_count, deployed_by, created_at
		FROM deployments
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var deployments []map[string]interface{}
	for rows.Next() {
		var id int64
		var siteID, deployedBy, createdAt string
		var sizeBytes, fileCount int64
		if err := rows.Scan(&id, &siteID, &sizeBytes, &fileCount, &deployedBy, &createdAt); err != nil {
			continue
		}
		deployments = append(deployments, map[string]interface{}{
			"id":          id,
			"site_id":     siteID,
			"size_bytes":  sizeBytes,
			"file_count":  fileCount,
			"deployed_by": deployedBy,
			"created_at":  createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"deployments": deployments,
	})
}
