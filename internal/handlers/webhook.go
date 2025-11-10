package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jikku/command-center/internal/database"
)

// WebhookHandler handles incoming webhooks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract endpoint from URL path (/webhook/{endpoint})
	path := strings.TrimPrefix(r.URL.Path, "/webhook/")
	endpoint := strings.TrimSpace(path)

	if endpoint == "" {
		http.Error(w, "Invalid webhook endpoint", http.StatusBadRequest)
		return
	}

	// Lookup webhook configuration
	db := database.GetDB()
	var webhookID int64
	var secret string
	var isActive bool
	var name string

	err := db.QueryRow(`
		SELECT id, name, secret, is_active FROM webhooks WHERE endpoint = ?
	`, endpoint).Scan(&webhookID, &name, &secret, &isActive)

	if err == sql.ErrNoRows {
		http.Error(w, "Webhook endpoint not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error looking up webhook: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if webhook is active
	if !isActive {
		http.Error(w, "Webhook is disabled", http.StatusForbidden)
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Verify signature if secret is configured
	if secret != "" {
		signature := r.Header.Get("X-Webhook-Signature")
		if signature == "" {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}

		if !verifySignature(body, secret, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse JSON payload (flexible structure)
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// If not JSON, store raw body
		payload = map[string]interface{}{
			"raw": string(body),
		}
	}

	// Extract useful fields if present
	eventType := "webhook"
	if et, ok := payload["event"].(string); ok {
		eventType = et
	} else if et, ok := payload["type"].(string); ok {
		eventType = et
	}

	// Convert payload back to JSON string for storage
	payloadJSON, _ := json.Marshal(payload)

	// Extract client info
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()

	// Log event to database
	_, err = db.Exec(`
		INSERT INTO events (domain, tags, source_type, event_type, path, referrer, user_agent, ip_address, query_params)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, endpoint, "", "webhook", eventType, "/webhook/"+endpoint, "", userAgent, ipAddress, string(payloadJSON))

	if err != nil {
		log.Printf("Error logging webhook event: %v", err)
		http.Error(w, "Failed to log event", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Webhook received",
		"webhook": name,
	})
}

// verifySignature verifies HMAC SHA256 signature
func verifySignature(body []byte, secret, signature string) bool {
	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
