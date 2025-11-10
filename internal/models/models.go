package models

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// Event represents a tracking event in the system
type Event struct {
	ID          int64     `json:"id"`
	Domain      string    `json:"domain"`
	Tags        []string  `json:"tags"`
	SourceType  string    `json:"source_type"` // web/pixel/redirect/webhook
	EventType   string    `json:"event_type"`  // pageview/click/redirect/webhook
	Path        string    `json:"path"`
	Referrer    string    `json:"referrer"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	QueryParams string    `json:"query_params"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
}

// TagsToString converts tags slice to comma-separated string for storage
func (e *Event) TagsToString() string {
	if len(e.Tags) == 0 {
		return ""
	}
	return strings.Join(e.Tags, ",")
}

// TagsFromString parses comma-separated string to tags slice
func (e *Event) TagsFromString(tagsStr string) {
	if tagsStr == "" {
		e.Tags = []string{}
		return
	}
	e.Tags = strings.Split(tagsStr, ",")
}

// Redirect represents a URL redirect with click tracking
type Redirect struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Destination string    `json:"destination"`
	Tags        []string  `json:"tags"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// TagsToString converts tags slice to comma-separated string for storage
func (r *Redirect) TagsToString() string {
	if len(r.Tags) == 0 {
		return ""
	}
	return strings.Join(r.Tags, ",")
}

// TagsFromString parses comma-separated string to tags slice
func (r *Redirect) TagsFromString(tagsStr string) {
	if tagsStr == "" {
		r.Tags = []string{}
		return
	}
	r.Tags = strings.Split(tagsStr, ",")
}

// Webhook represents a webhook endpoint configuration
type Webhook struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint"`
	Secret    string    `json:"secret,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Notification represents a sent notification
type Notification struct {
	ID               int64          `json:"id"`
	EventID          sql.NullInt64  `json:"event_id,omitempty"`
	NotificationType string         `json:"notification_type"`
	Message          string         `json:"message"`
	SentAt           time.Time      `json:"sent_at"`
}

// Stats represents dashboard statistics
type Stats struct {
	TotalEventsToday     int64            `json:"total_events_today"`
	TotalEventsWeek      int64            `json:"total_events_week"`
	TotalEventsMonth     int64            `json:"total_events_month"`
	TotalEventsAllTime   int64            `json:"total_events_all_time"`
	EventsBySourceType   map[string]int64 `json:"events_by_source_type"`
	TopDomains           []DomainStat     `json:"top_domains"`
	TopTags              []TagStat        `json:"top_tags"`
	EventsTimeline       []TimelineStat   `json:"events_timeline"`
	TotalUniqueDomains   int64            `json:"total_unique_domains"`
	TotalRedirectClicks  int64            `json:"total_redirect_clicks"`
}

// DomainStat represents statistics for a domain
type DomainStat struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

// TagStat represents statistics for a tag
type TagStat struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

// TimelineStat represents events in a time bucket
type TimelineStat struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// TrackRequest represents an incoming tracking request
type TrackRequest struct {
	Hostname    string            `json:"h"`     // hostname/domain
	Domain      string            `json:"d"`     // explicit domain override
	Path        string            `json:"p"`     // page path
	EventType   string            `json:"e"`     // event type
	Tags        []string          `json:"t"`     // tags array
	QueryParams map[string]string `json:"q"`     // query parameters
	Referrer    string            `json:"ref"`   // referrer
}

// ToQueryParamsJSON converts query params map to JSON string
func (tr *TrackRequest) ToQueryParamsJSON() string {
	if len(tr.QueryParams) == 0 {
		return ""
	}
	bytes, err := json.Marshal(tr.QueryParams)
	if err != nil {
		return ""
	}
	return string(bytes)
}
