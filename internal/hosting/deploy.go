package hosting

import (
	"archive/zip"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"os"
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

// DeploySite extracts a ZIP file to the site directory
func DeploySite(zipReader *zip.Reader, subdomain string) (*DeployResult, error) {
	// Validate subdomain
	if err := ValidateSubdomain(subdomain); err != nil {
		return nil, err
	}

	siteDir := GetSiteDir(subdomain)

	// Clean existing site directory
	if err := os.RemoveAll(siteDir); err != nil {
		return nil, fmt.Errorf("failed to clean existing site: %w", err)
	}

	// Create site directory
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create site directory: %w", err)
	}

	var totalSize int64
	var fileCount int

	// Extract files
	for _, file := range zipReader.File {
		// Security: Prevent path traversal
		cleanPath := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
			continue // Skip files that try to escape
		}

		destPath := filepath.Join(siteDir, cleanPath)

		// Verify the destination is within the site directory
		if !strings.HasPrefix(destPath, filepath.Clean(siteDir)+string(os.PathSeparator)) {
			continue // Skip files that escape via symlinks or other tricks
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", cleanPath, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Extract file
		if err := extractFile(file, destPath); err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", cleanPath, err)
		}

		totalSize += file.FileInfo().Size()
		fileCount++
	}

	return &DeployResult{
		SiteID:    subdomain,
		SizeBytes: totalSize,
		FileCount: fileCount,
	}, nil
}

// extractFile extracts a single file from the ZIP archive
func extractFile(file *zip.File, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	// Limit file size to prevent zip bombs (100MB per file)
	limited := io.LimitReader(src, 100*1024*1024)
	_, err = io.Copy(dst, limited)
	return err
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
