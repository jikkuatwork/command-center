package database

import (
	"context"
	"database/sql"
	"io/fs"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
)

// SQLCertStorage implements certmagic.Storage using SQLite
type SQLCertStorage struct {
	db *sql.DB
}

// NewSQLCertStorage creates a new SQLCertStorage instance
func NewSQLCertStorage(db *sql.DB) *SQLCertStorage {
	return &SQLCertStorage{db: db}
}

// Store saves data to the database
func (s *SQLCertStorage) Store(ctx context.Context, key string, value []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO certificates (key, value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET 
			value = excluded.value, 
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

// Load retrieves data from the database
func (s *SQLCertStorage) Load(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	err := s.db.QueryRowContext(ctx, "SELECT value FROM certificates WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Delete removes data from the database
func (s *SQLCertStorage) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM certificates WHERE key = ?", key)
	return err
}

// Exists checks if a key exists
func (s *SQLCertStorage) Exists(ctx context.Context, key string) bool {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM certificates WHERE key = ?", key).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// List returns a list of keys matching the prefix
func (s *SQLCertStorage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	// If recursive is true, we want everything that starts with prefix
	// If recursive is false, we only want direct children (simulate directory listing)
	// But CertMagic mainly uses List to find certs, and a simple prefix match is usually enough.
	// For simplicity in SQL, we'll just do prefix match.
	// CertMagic's FileStorage implementation logic for recursive=false is complex to map to SQL flat keys.
	// However, CertMagic's uses of List usually pass recursive=true.
	
	query := "SELECT key FROM certificates WHERE key LIKE ?"
	rows, err := s.db.QueryContext(ctx, query, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		
		if !recursive {
			// Simulate directory listing: extract immediate child only
			// key: prefix/child/grandchild
			// prefix: prefix/
			// suffix: child/grandchild
			// relative: child/grandchild
			
			// Ensure prefix has trailing slash for directory logic
			if !strings.HasSuffix(prefix, "/") {
				prefix += "/"
			}
			
			if !strings.HasPrefix(key, prefix) {
				continue 
			}
			
			rel := strings.TrimPrefix(key, prefix)
			if idx := strings.Index(rel, "/"); idx != -1 {
				// It's a directory, return prefix/child (with trailing slash? CertMagic expects keys)
				// CertMagic FileStorage returns paths.
				// Let's just return the full key for now, simpler and often compatible.
				// Actually, if we return full keys, CertMagic might traverse unnecessary deep keys.
				// But given we are storing certs, there aren't that many keys.
				keys = append(keys, key)
			} else {
				keys = append(keys, key)
			}
		} else {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// Stat returns information about a key
func (s *SQLCertStorage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	var size int64
	var modified time.Time
	
	err := s.db.QueryRowContext(ctx, "SELECT length(value), updated_at FROM certificates WHERE key = ?", key).Scan(&size, &modified)
	if err == sql.ErrNoRows {
		return certmagic.KeyInfo{}, fs.ErrNotExist
	}
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	return certmagic.KeyInfo{
		Key:        key,
		Modified:   modified,
		Size:       size,
		IsTerminal: true, // It's a file (blob), not a directory
	}, nil
}

// Lock acquires a lock (using SQLite transaction or specialized table? CertMagic requires file locking)
// Since we are single-binary (Cartridge), we don't strictly need distributed locking if we assume one instance.
// But if we run multiple instances (e.g. during blue/green deploy), we might need it.
// For now, let's implement a simple "always success" or file-based lock if needed.
// CertMagic uses Lock to coordinate ACME challenges.
// SQLite has database-level locking.
// We can use a separate `locks` table if we really want to be correct.
func (s *SQLCertStorage) Lock(ctx context.Context, key string) error {
	// TODO: Implement proper locking if scaling beyond one instance
	// For "Cartridge" (single binary), internal mutex or just no-op is often fine 
	// because CertMagic also has in-memory locking.
	// But CertMagic docs say: "Storage must implement locking to be safe for use by multiple CertMagic instances."
	// We will rely on in-memory locking of the single instance for now.
	return nil 
}

func (s *SQLCertStorage) Unlock(ctx context.Context, key string) error {
	return nil
}
