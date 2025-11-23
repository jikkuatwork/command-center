package hosting

import (
	"net/http"
	"path/filepath"
	"strings"
)

// SecureFileServer returns a file server that prevents path traversal
// and restricts access to the specified root directory
type SecureFileServer struct {
	root    string
	handler http.Handler
}

// NewSecureFileServer creates a secure file server for the given directory
func NewSecureFileServer(root string) *SecureFileServer {
	absRoot, _ := filepath.Abs(root)
	return &SecureFileServer{
		root:    absRoot,
		handler: http.FileServer(http.Dir(root)),
	}
}

// ServeHTTP handles requests with path traversal protection
func (s *SecureFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	cleanPath := filepath.Clean(r.URL.Path)

	// Block obvious traversal attempts
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify the requested path stays within root
	fullPath := filepath.Join(s.root, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil || !strings.HasPrefix(absPath, s.root) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Block access to hidden files (starting with .)
	parts := strings.Split(cleanPath, "/")
	for _, part := range parts {
		if len(part) > 0 && part[0] == '.' {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	s.handler.ServeHTTP(w, r)
}

// SecurityLimits defines resource limits for serverless execution
type SecurityLimits struct {
	MaxExecutionTime int64 // milliseconds
	MaxMemoryBytes   int64 // bytes (not strictly enforced by Goja)
	MaxFileSize      int64 // bytes for uploaded files
	MaxSiteSize      int64 // total bytes for a site
}

// DefaultLimits returns the default security limits
func DefaultLimits() *SecurityLimits {
	return &SecurityLimits{
		MaxExecutionTime: 100,                 // 100ms
		MaxMemoryBytes:   50 * 1024 * 1024,    // 50MB
		MaxFileSize:      100 * 1024 * 1024,   // 100MB per file
		MaxSiteSize:      500 * 1024 * 1024,   // 500MB total per site
	}
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove null bytes and other control characters
	var result strings.Builder
	for _, r := range input {
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateSiteID ensures a site ID is safe to use in file paths
func ValidateSiteID(siteID string) bool {
	// Must be non-empty
	if siteID == "" {
		return false
	}

	// Must not exceed length
	if len(siteID) > 63 {
		return false
	}

	// Must only contain safe characters
	for _, c := range siteID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	// Must not start or end with hyphen
	if siteID[0] == '-' || siteID[len(siteID)-1] == '-' {
		return false
	}

	// Must not contain consecutive hyphens
	if strings.Contains(siteID, "--") {
		return false
	}

	return true
}
