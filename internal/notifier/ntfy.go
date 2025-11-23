package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jikku/command-center/internal/config"
	"github.com/jikku/command-center/internal/database"
)

// Notification types
const (
	NotificationTrafficSpike = "traffic_spike"
	NotificationNewDomain    = "new_domain"
	NotificationWebhook      = "webhook_event"
	NotificationError        = "error"
)

// Send sends a notification to ntfy.sh
func Send(title, message, notificationType string) error {
	cfg := config.Get()

	// In development mode, just log instead of sending
	if cfg.IsDevelopment() {
		log.Printf("[NTFY MOCK] Type: %s, Title: %s, Message: %s", notificationType, title, message)
		return logNotification(0, notificationType, fmt.Sprintf("%s: %s", title, message))
	}

	// Check if ntfy topic is configured
	if cfg.Ntfy.Topic == "" {
		log.Println("ntfy.sh topic not configured, skipping notification")
		return nil
	}

	// Prepare notification payload
	payload := map[string]interface{}{
		"topic":   cfg.Ntfy.Topic,
		"title":   title,
		"message": message,
		"tags":    []string{"command-center", notificationType},
	}

	// Add priority based on type
	switch notificationType {
	case NotificationError:
		payload["priority"] = "high"
	case NotificationTrafficSpike:
		payload["priority"] = "default"
	default:
		payload["priority"] = "low"
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send to ntfy.sh
	url := cfg.Ntfy.URL + "/" + cfg.Ntfy.Topic
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy.sh returned status %d", resp.StatusCode)
	}

	log.Printf("Notification sent: %s - %s", title, message)
	return logNotification(0, notificationType, fmt.Sprintf("%s: %s", title, message))
}

// logNotification stores notification in database
func logNotification(eventID int64, notificationType, message string) error {
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO notifications (event_id, notification_type, message)
		VALUES (?, ?, ?)
	`, eventID, notificationType, message)
	return err
}

// CheckTrafficSpike detects unusual traffic patterns
func CheckTrafficSpike() error {
	db := database.GetDB()

	// Get average hourly events
	var avgHourly float64
	err := db.QueryRow(`
		SELECT AVG(hourly_count) FROM (
			SELECT COUNT(*) as hourly_count
			FROM events
			WHERE created_at >= DATETIME('now', '-24 hours')
			GROUP BY strftime('%Y-%m-%d %H', created_at)
		)
	`).Scan(&avgHourly)

	if err != nil {
		return err
	}

	// Get current hour count
	var currentHour int64
	err = db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE created_at >= DATETIME('now', '-1 hour')
	`).Scan(&currentHour)

	if err != nil {
		return err
	}

	// If current hour is 10x average, send alert
	threshold := avgHourly * 10
	if float64(currentHour) > threshold && currentHour > 10 {
		return Send(
			"Traffic Spike Detected",
			fmt.Sprintf("Current hour: %d events (avg: %.1f)", currentHour, avgHourly),
			NotificationTrafficSpike,
		)
	}

	return nil
}

// CheckNewDomain checks for new domains and sends notification
func CheckNewDomain(domain string) error {
	// Only notify for non-empty, non-unknown domains
	if domain == "" || domain == "unknown" {
		return nil
	}

	db := database.GetDB()

	// Check if this domain was seen before (more than 1 day ago)
	var count int64
	err := db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE domain = ? AND created_at < DATETIME('now', '-1 day')
	`, domain).Scan(&count)

	if err != nil {
		return err
	}

	// If this is the first time seeing this domain in the last day, notify
	if count == 0 {
		// Check if it's truly new (first ever)
		var totalCount int64
		db.QueryRow(`SELECT COUNT(*) FROM events WHERE domain = ?`, domain).Scan(&totalCount)

		if totalCount == 1 {
			return Send(
				"New Domain Detected",
				fmt.Sprintf("First event from domain: %s", domain),
				NotificationNewDomain,
			)
		}
	}

	return nil
}

// NotifyWebhookEvent sends notification for webhook events
func NotifyWebhookEvent(webhookName, eventType string) error {
	return Send(
		fmt.Sprintf("Webhook: %s", webhookName),
		fmt.Sprintf("Event type: %s", eventType),
		NotificationWebhook,
	)
}

// NotifyError sends error notification
func NotifyError(errorMsg string) error {
	return Send(
		"Error Detected",
		errorMsg,
		NotificationError,
	)
}
