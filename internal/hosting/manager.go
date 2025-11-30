package hosting

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// fs is the active file system
	fs FileSystem
	
	// db is the database connection
	database *sql.DB

	// validSubdomainRegex matches valid subdomain names
	validSubdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// Init initializes the hosting system
func Init(db *sql.DB) error {
	database = db

	// Initialize VFS
	fs = NewSQLFileSystem(db)

	return nil
}

// GetFileSystem returns the active file system
func GetFileSystem() FileSystem {
	return fs
}

// SiteExists checks if a site directory exists
func SiteExists(subdomain string) bool {
	// Check VFS first
	exists, err := fs.Exists(subdomain, "index.html")
	if err == nil && exists {
		return true
	}
	// Check for main.js (serverless)
	exists, err = fs.Exists(subdomain, "main.js")
	if err == nil && exists {
		return true
	}
	
	return false
}

// ValidateSubdomain checks if a subdomain name is valid
func ValidateSubdomain(subdomain string) error {
	subdomain = strings.ToLower(subdomain)

	if len(subdomain) < 1 || len(subdomain) > 63 {
		return fmt.Errorf("subdomain must be 1-63 characters")
	}

	if !validSubdomainRegex.MatchString(subdomain) {
		return fmt.Errorf("subdomain must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}

	// Reserved subdomains
	reserved := []string{"www", "api", "admin", "mail", "ftp", "smtp", "pop", "imap", "ns1", "ns2", "localhost"}
	for _, r := range reserved {
		if subdomain == r {
			return fmt.Errorf("'%s' is a reserved subdomain", subdomain)
		}
	}

	return nil
}

// ListSites returns all hosted sites
func ListSites() ([]SiteInfo, error) {
	if database == nil {
		return nil, fmt.Errorf("hosting not initialized")
	}

	query := `
		SELECT site_id, COUNT(*) as file_count, SUM(size_bytes) as total_size, MAX(updated_at) as last_mod
		FROM files
		GROUP BY site_id
		ORDER BY last_mod DESC
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []SiteInfo
	for rows.Next() {
		var site SiteInfo
		var lastMod time.Time
		if err := rows.Scan(&site.Name, &site.FileCount, &site.SizeBytes, &lastMod); err != nil {
			continue
		}
		site.ModTime = lastMod
		site.Path = "vfs://" + site.Name
		sites = append(sites, site)
	}

	return sites, nil
}

// SiteInfo contains information about a hosted site
type SiteInfo struct {
	Name      string
	Path      string
	FileCount int
	SizeBytes int64
	ModTime   interface{} // time.Time
}

// CreateSite creates a new site (placeholder for VFS)
func CreateSite(subdomain string) error {
	return ValidateSubdomain(subdomain)
}

// DeleteSite removes a site and all its contents
func DeleteSite(subdomain string) error {
	// Clean up WebSocket hub
	RemoveHub(subdomain)

	// Delete from VFS
	return fs.DeleteSite(subdomain)
}
