package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jikku/command-center/internal/database"
	"github.com/jikku/command-center/internal/models"
)

// StatsHandler returns dashboard statistics
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := database.GetDB()
	stats := models.Stats{
		EventsBySourceType: make(map[string]int64),
	}

	// Total events today
	db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE DATE(created_at) = DATE('now')
	`).Scan(&stats.TotalEventsToday)

	// Total events this week
	db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE created_at >= DATE('now', '-7 days')
	`).Scan(&stats.TotalEventsWeek)

	// Total events this month
	db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE created_at >= DATE('now', '-30 days')
	`).Scan(&stats.TotalEventsMonth)

	// Total events all time
	db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&stats.TotalEventsAllTime)

	// Events by source type
	rows, _ := db.Query(`
		SELECT source_type, COUNT(*) as count
		FROM events
		GROUP BY source_type
	`)
	defer rows.Close()
	for rows.Next() {
		var sourceType string
		var count int64
		rows.Scan(&sourceType, &count)
		stats.EventsBySourceType[sourceType] = count
	}

	// Top 10 domains
	rows, _ = db.Query(`
		SELECT domain, COUNT(*) as count
		FROM events
		WHERE domain != ''
		GROUP BY domain
		ORDER BY count DESC
		LIMIT 10
	`)
	defer rows.Close()
	for rows.Next() {
		var ds models.DomainStat
		rows.Scan(&ds.Domain, &ds.Count)
		stats.TopDomains = append(stats.TopDomains, ds)
	}

	// Top 10 tags
	rows, _ = db.Query(`
		SELECT tags, COUNT(*) as count
		FROM events
		WHERE tags != ''
		GROUP BY tags
		ORDER BY count DESC
		LIMIT 10
	`)
	defer rows.Close()
	for rows.Next() {
		var tagsStr string
		var count int64
		rows.Scan(&tagsStr, &count)
		// Split tags and count individually
		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			stats.TopTags = append(stats.TopTags, models.TagStat{Tag: strings.TrimSpace(tag), Count: count})
		}
	}

	// Events timeline (hourly for last 24 hours)
	rows, _ = db.Query(`
		SELECT strftime('%Y-%m-%d %H:00', created_at) as hour, COUNT(*) as count
		FROM events
		WHERE created_at >= DATETIME('now', '-24 hours')
		GROUP BY hour
		ORDER BY hour
	`)
	defer rows.Close()
	for rows.Next() {
		var ts models.TimelineStat
		rows.Scan(&ts.Timestamp, &ts.Count)
		stats.EventsTimeline = append(stats.EventsTimeline, ts)
	}

	// Total unique domains
	db.QueryRow(`SELECT COUNT(DISTINCT domain) FROM events`).Scan(&stats.TotalUniqueDomains)

	// Total redirect clicks
	db.QueryRow(`SELECT COALESCE(SUM(click_count), 0) FROM redirects`).Scan(&stats.TotalRedirectClicks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// EventsHandler returns paginated events with filtering
func EventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	domain := query.Get("domain")
	tags := query.Get("tags")
	sourceType := query.Get("source_type")
	limit := parseInt(query.Get("limit"), 50)
	offset := parseInt(query.Get("offset"), 0)

	// Build query
	where := []string{"1=1"}
	args := []interface{}{}

	if domain != "" {
		where = append(where, "domain = ?")
		args = append(args, domain)
	}
	if tags != "" {
		where = append(where, "tags LIKE ?")
		args = append(args, "%"+tags+"%")
	}
	if sourceType != "" {
		where = append(where, "source_type = ?")
		args = append(args, sourceType)
	}

	whereClause := strings.Join(where, " AND ")
	sql := "SELECT id, domain, tags, source_type, event_type, path, referrer, user_agent, ip_address, created_at FROM events WHERE " + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	db := database.GetDB()
	rows, err := db.Query(sql, args...)
	if err != nil {
		log.Printf("Error querying events: %v", err)
		http.Error(w, "Failed to query events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var domain, tags, sourceType, eventType, path, referrer, userAgent, ipAddress string
		var createdAt time.Time

		rows.Scan(&id, &domain, &tags, &sourceType, &eventType, &path, &referrer, &userAgent, &ipAddress, &createdAt)

		events = append(events, map[string]interface{}{
			"id":          id,
			"domain":      domain,
			"tags":        strings.Split(tags, ","),
			"source_type": sourceType,
			"event_type":  eventType,
			"path":        path,
			"referrer":    referrer,
			"user_agent":  userAgent,
			"ip_address":  ipAddress,
			"created_at":  createdAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// DomainsHandler returns list of domains with event counts
func DomainsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT domain, COUNT(*) as count
		FROM events
		WHERE domain != ''
		GROUP BY domain
		ORDER BY count DESC
	`)
	if err != nil {
		log.Printf("Error querying domains: %v", err)
		http.Error(w, "Failed to query domains", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	domains := []models.DomainStat{}
	for rows.Next() {
		var ds models.DomainStat
		rows.Scan(&ds.Domain, &ds.Count)
		domains = append(domains, ds)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}

// TagsHandler returns list of tags with usage counts
func TagsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT tags, COUNT(*) as count
		FROM events
		WHERE tags != ''
		GROUP BY tags
		ORDER BY count DESC
	`)
	if err != nil {
		log.Printf("Error querying tags: %v", err)
		http.Error(w, "Failed to query tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Aggregate tags (they may be comma-separated)
	tagCounts := make(map[string]int64)
	for rows.Next() {
		var tagsStr string
		var count int64
		rows.Scan(&tagsStr, &count)

		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagCounts[tag] += count
			}
		}
	}

	// Convert to array
	tags := []models.TagStat{}
	for tag, count := range tagCounts {
		tags = append(tags, models.TagStat{Tag: tag, Count: count})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// RedirectsHandler handles redirects CRUD
func RedirectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		// List all redirects
		db := database.GetDB()
		rows, err := db.Query(`
			SELECT id, slug, destination, tags, click_count, created_at
			FROM redirects
			ORDER BY click_count DESC
		`)
		if err != nil {
			log.Printf("Error querying redirects: %v", err)
			http.Error(w, "Failed to query redirects", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		redirects := []map[string]interface{}{}
		for rows.Next() {
			var id, clickCount int64
			var slug, destination, tags string
			var createdAt time.Time

			rows.Scan(&id, &slug, &destination, &tags, &clickCount, &createdAt)

			redirects = append(redirects, map[string]interface{}{
				"id":          id,
				"slug":        slug,
				"destination": destination,
				"tags":        strings.Split(tags, ","),
				"click_count": clickCount,
				"created_at":  createdAt.Format(time.RFC3339),
			})
		}

		json.NewEncoder(w).Encode(redirects)

	} else if r.Method == http.MethodPost {
		// Create new redirect
		var req struct {
			Slug        string   `json:"slug"`
			Destination string   `json:"destination"`
			Tags        []string `json:"tags"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate
		if req.Slug == "" || req.Destination == "" {
			http.Error(w, "Slug and destination are required", http.StatusBadRequest)
			return
		}

		// Check if slug exists
		db := database.GetDB()
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM redirects WHERE slug = ?", req.Slug).Scan(&exists)
		if exists > 0 {
			http.Error(w, "Slug already exists", http.StatusConflict)
			return
		}

		// Insert
		tagsStr := strings.Join(req.Tags, ",")
		result, err := db.Exec(`
			INSERT INTO redirects (slug, destination, tags)
			VALUES (?, ?, ?)
		`, req.Slug, req.Destination, tagsStr)

		if err != nil {
			log.Printf("Error creating redirect: %v", err)
			http.Error(w, "Failed to create redirect", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          id,
			"slug":        req.Slug,
			"destination": req.Destination,
			"tags":        req.Tags,
			"click_count": 0,
		})

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// WebhooksHandler handles webhooks CRUD
func WebhooksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		// List all webhooks
		db := database.GetDB()
		rows, err := db.Query(`
			SELECT id, name, endpoint, secret, is_active, created_at
			FROM webhooks
			ORDER BY created_at DESC
		`)
		if err != nil {
			log.Printf("Error querying webhooks: %v", err)
			http.Error(w, "Failed to query webhooks", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		webhooks := []map[string]interface{}{}
		for rows.Next() {
			var id int64
			var name, endpoint, secret string
			var isActive bool
			var createdAt time.Time

			rows.Scan(&id, &name, &endpoint, &secret, &isActive, &createdAt)

			webhooks = append(webhooks, map[string]interface{}{
				"id":         id,
				"name":       name,
				"endpoint":   endpoint,
				"has_secret": secret != "",
				"is_active":  isActive,
				"created_at": createdAt.Format(time.RFC3339),
			})
		}

		json.NewEncoder(w).Encode(webhooks)

	} else if r.Method == http.MethodPost {
		// Create new webhook
		var req struct {
			Name     string `json:"name"`
			Endpoint string `json:"endpoint"`
			Secret   string `json:"secret"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate
		if req.Name == "" || req.Endpoint == "" {
			http.Error(w, "Name and endpoint are required", http.StatusBadRequest)
			return
		}

		// Check if endpoint exists
		db := database.GetDB()
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM webhooks WHERE endpoint = ?", req.Endpoint).Scan(&exists)
		if exists > 0 {
			http.Error(w, "Endpoint already exists", http.StatusConflict)
			return
		}

		// Insert
		result, err := db.Exec(`
			INSERT INTO webhooks (name, endpoint, secret, is_active)
			VALUES (?, ?, ?, 1)
		`, req.Name, req.Endpoint, req.Secret)

		if err != nil {
			log.Printf("Error creating webhook: %v", err)
			http.Error(w, "Failed to create webhook", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"name":       req.Name,
			"endpoint":   req.Endpoint,
			"has_secret": req.Secret != "",
			"is_active":  true,
		})

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DashboardHandler serves the main dashboard page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/index.html")
}

// parseInt parses string to int with default value
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return i
}
