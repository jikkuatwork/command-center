package hosting

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// FileSystem defines the interface for site storage
type FileSystem interface {
	WriteFile(siteID, path string, content io.Reader, size int64, mimeType string) error
	ReadFile(siteID, path string) (*File, error)
	DeleteSite(siteID string) error
	Exists(siteID, path string) (bool, error)
}

// File represents a file in the VFS
type File struct {
	Content  io.ReadCloser
	Size     int64
	MimeType string
	Hash     string
	ModTime  time.Time
}

// SQLFileSystem implements FileSystem using SQLite
type SQLFileSystem struct {
	db *sql.DB
}

// NewSQLFileSystem creates a new SQL-backed file system
func NewSQLFileSystem(db *sql.DB) *SQLFileSystem {
	return &SQLFileSystem{db: db}
}

// WriteFile writes a file to the database
func (fs *SQLFileSystem) WriteFile(siteID, path string, content io.Reader, size int64, mimeType string) error {
	// Read content to calculate hash and prepare for blob
	// Note: For very large files, this might be memory intensive.
	// Since we are targeting a 25MB binary/small VPS, keeping files in memory (up to 100MB limit)
	// before write is acceptable, but streaming is better if the driver supports it.
	// SQLite drivers often require the full []byte for BLOBs.
	
	data, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Insert or Replace
	query := `
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(site_id, path) DO UPDATE SET
			content = excluded.content,
			size_bytes = excluded.size_bytes,
			mime_type = excluded.mime_type,
			hash = excluded.hash,
			updated_at = CURRENT_TIMESTAMP
	`
	
	_, err = fs.db.Exec(query, siteID, path, data, size, mimeType, hashStr)
	if err != nil {
		return fmt.Errorf("failed to write file to DB: %w", err)
	}

	return nil
}

// ReadFile reads a file from the database
func (fs *SQLFileSystem) ReadFile(siteID, path string) (*File, error) {
	query := `
		SELECT content, size_bytes, mime_type, hash, updated_at
		FROM files WHERE site_id = ? AND path = ?
	`
	
	var data []byte
	var size int64
	var mimeType, hash string
	var modTime time.Time

	err := fs.db.QueryRow(query, siteID, path).Scan(&data, &size, &mimeType, &hash, &modTime)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found") // OS-agnostic error?
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &File{
		Content:  io.NopCloser(newByteReader(data)),
		Size:     size,
		MimeType: mimeType,
		Hash:     hash,
		ModTime:  modTime,
	}, nil
}

// DeleteSite deletes all files for a site
func (fs *SQLFileSystem) DeleteSite(siteID string) error {
	_, err := fs.db.Exec("DELETE FROM files WHERE site_id = ?", siteID)
	return err
}

// Exists checks if a file exists
func (fs *SQLFileSystem) Exists(siteID, path string) (bool, error) {
	var count int
	err := fs.db.QueryRow("SELECT COUNT(*) FROM files WHERE site_id = ? AND path = ?", siteID, path).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Helper for byte reader
type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
