package hosting

import (
	"archive/zip"
	"crypto/rand"
	"database/sql"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DeployResult contains information about a deployment
type DeployResult struct {
	SiteID    string
	SizeBytes int64
	FileCount int
}

// DeploySite extracts a ZIP file to the VFS
func DeploySite(zipReader *zip.Reader, subdomain string) (*DeployResult, error) {
	// Validate subdomain
	if err := ValidateSubdomain(subdomain); err != nil {
		return nil, err
	}

	// Clear existing site files?
	// The VFS WriteFile does INSERT OR UPDATE, so files are overwritten.
	// But stale files (files removed in the new deploy) would remain.
	// Ideally we should delete the site first or track current files.
	// For now, let's delete the site first to ensure a clean state (Cartridge style).
	if err := fs.DeleteSite(subdomain); err != nil {
		return nil, fmt.Errorf("failed to clear existing site: %w", err)
	}

	var totalSize int64
	var fileCount int

	// Extract files
	for _, file := range zipReader.File {
		// Security: Prevent path traversal
		cleanPath := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") || strings.Contains(cleanPath, "\\") {
			continue // Skip files that try to escape
		}
		
		// Normalize path to forward slashes for DB consistency
		cleanPath = filepath.ToSlash(cleanPath)

		// Skip directories (we only store files)
		if file.FileInfo().IsDir() {
			continue
		}

		// Open file from zip
		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		// Determine MIME type
		ext := filepath.Ext(cleanPath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Write to VFS
		fileSize := file.FileInfo().Size()
		if err := fs.WriteFile(subdomain, cleanPath, src, fileSize, mimeType); err != nil {
			src.Close()
			return nil, fmt.Errorf("failed to write file %s: %w", cleanPath, err)
		}
		src.Close()

		totalSize += fileSize
		fileCount++
	}

	return &DeployResult{
		SiteID:    subdomain,
		SizeBytes: totalSize,
		FileCount: fileCount,
	}, nil
}

// ValidateAPIKey validates an API key against the database
func ValidateAPIKey(db *sql.DB, token string) (int64, string, error) {
	// Get all API keys from database
	rows, err := db.Query("SELECT id, name, key_hash FROM api_keys")
	if err != nil {
		return 0, "", fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, keyHash string
		if err := rows.Scan(&id, &name, &keyHash); err != nil {
			continue
		}

		// Compare token with hash
		if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(token)); err == nil {
			// Update last_used_at
			db.Exec("UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?", id)
			return id, name, nil
		}
	}

	return 0, "", fmt.Errorf("invalid API key")
}

// CreateAPIKey creates a new API key and returns the raw token
func CreateAPIKey(db *sql.DB, name string, scopes string) (string, error) {
	// Generate random token (32 bytes = 64 hex chars)
	token, err := generateRandomToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Store in database
	_, err = db.Exec(
		"INSERT INTO api_keys (name, key_hash, scopes) VALUES (?, ?, ?)",
		name, string(hash), scopes,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store API key: %w", err)
	}

	return token, nil
}

// generateRandomToken generates a random hex token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// RecordDeployment records a deployment in the database
func RecordDeployment(db *sql.DB, siteID string, sizeBytes int64, fileCount int, deployedBy string) error {
	_, err := db.Exec(
		"INSERT INTO deployments (site_id, size_bytes, file_count, deployed_by) VALUES (?, ?, ?, ?)",
		siteID, sizeBytes, fileCount, deployedBy,
	)
	return err
}

// ListAPIKeys lists all API keys (without the actual keys)
func ListAPIKeys(db *sql.DB) ([]APIKeyInfo, error) {
	rows, err := db.Query("SELECT id, name, scopes, created_at, last_used_at FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKeyInfo
	for rows.Next() {
		var k APIKeyInfo
		var lastUsed sql.NullTime
		if err := rows.Scan(&k.ID, &k.Name, &k.Scopes, &k.CreatedAt, &lastUsed); err != nil {
			continue
		}
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		keys = append(keys, k)
	}

	return keys, nil
}

// APIKeyInfo contains information about an API key
type APIKeyInfo struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Scopes     string     `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// DeleteAPIKey deletes an API key by ID
func DeleteAPIKey(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	return err
}
