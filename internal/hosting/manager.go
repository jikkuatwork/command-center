package hosting

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// SitesDir is the base directory for all hosted sites
	sitesBaseDir string

	// validSubdomainRegex matches valid subdomain names
	validSubdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// Init initializes the hosting system with the config directory
func Init(configDir string) error {
	sitesBaseDir = filepath.Join(configDir, "sites")

	// Create sites directory if it doesn't exist
	if err := os.MkdirAll(sitesBaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create sites directory: %w", err)
	}

	return nil
}

// GetSitesDir returns the base directory for all sites
func GetSitesDir() string {
	return sitesBaseDir
}

// GetSiteDir returns the directory for a specific site
func GetSiteDir(subdomain string) string {
	return filepath.Join(sitesBaseDir, subdomain)
}

// SiteExists checks if a site directory exists
func SiteExists(subdomain string) bool {
	siteDir := GetSiteDir(subdomain)
	info, err := os.Stat(siteDir)
	if err != nil {
		return false
	}
	return info.IsDir()
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

// ListSites returns all site directories
func ListSites() ([]SiteInfo, error) {
	if sitesBaseDir == "" {
		return nil, fmt.Errorf("hosting not initialized")
	}

	entries, err := os.ReadDir(sitesBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SiteInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read sites directory: %w", err)
	}

	var sites []SiteInfo
	for _, entry := range entries {
		if entry.IsDir() {
			siteDir := filepath.Join(sitesBaseDir, entry.Name())
			info, _ := entry.Info()

			// Count files and calculate size
			fileCount, totalSize := countFiles(siteDir)

			sites = append(sites, SiteInfo{
				Name:      entry.Name(),
				Path:      siteDir,
				FileCount: fileCount,
				SizeBytes: totalSize,
				ModTime:   info.ModTime(),
			})
		}
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

// countFiles recursively counts files and calculates total size
func countFiles(dir string) (int, int64) {
	var count int
	var size int64

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		count++
		size += info.Size()
		return nil
	})

	return count, size
}

// CreateSite creates a new site directory
func CreateSite(subdomain string) error {
	if err := ValidateSubdomain(subdomain); err != nil {
		return err
	}

	siteDir := GetSiteDir(subdomain)
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		return fmt.Errorf("failed to create site directory: %w", err)
	}

	return nil
}

// DeleteSite removes a site directory and all its contents
func DeleteSite(subdomain string) error {
	siteDir := GetSiteDir(subdomain)

	// Validate the path to prevent directory traversal
	absPath, err := filepath.Abs(siteDir)
	if err != nil {
		return fmt.Errorf("invalid site path: %w", err)
	}

	baseAbs, err := filepath.Abs(sitesBaseDir)
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}

	// Ensure the site directory is within the sites base directory
	if !strings.HasPrefix(absPath, baseAbs+string(os.PathSeparator)) {
		return fmt.Errorf("invalid site path: directory traversal detected")
	}

	// Clean up WebSocket hub for this site (prevents goroutine leak)
	RemoveHub(subdomain)

	return os.RemoveAll(absPath)
}
