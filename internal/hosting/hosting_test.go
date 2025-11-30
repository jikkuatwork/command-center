package hosting

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary in-memory database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	// Enable WAL (though not strictly needed for :memory:)
	db.Exec("PRAGMA journal_mode=WAL")

	// Create schema
	schema := `
	CREATE TABLE files (
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content BLOB,
		size_bytes INTEGER NOT NULL,
		mime_type TEXT,
		hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (site_id, path)
	);
	CREATE TABLE api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		scopes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	);
	CREATE TABLE deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		site_id TEXT NOT NULL,
		size_bytes INTEGER,
		file_count INTEGER,
		deployed_by TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestVFS_WriteAndRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fs := NewSQLFileSystem(db)
	
	// Test Write
	content := []byte("Hello World")
	err := fs.WriteFile("site1", "index.html", bytes.NewReader(content), int64(len(content)), "text/html")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test Read
	file, err := fs.ReadFile("site1", "index.html")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	defer file.Content.Close()

	readContent, _ := io.ReadAll(file.Content)
	if string(readContent) != string(content) {
		t.Errorf("Content mismatch. Got %s, want %s", readContent, content)
	}
	if file.MimeType != "text/html" {
		t.Errorf("MimeType mismatch. Got %s, want text/html", file.MimeType)
	}
}

func TestDeploySite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Initialize hosting with DB
	Init(db)

	// Create a mock zip file
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	
	files := map[string]string{
		"index.html": "<h1>Hello</h1>",
		"css/style.css": "body { color: red; }",
		"main.js": "console.log('test');",
	}

	for name, content := range files {
		f, _ := zipWriter.Create(name)
		f.Write([]byte(content))
	}
	zipWriter.Close()

	// Create reader from buffer
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to create zip reader: %v", err)
	}

	// Deploy
	res, err := DeploySite(zipReader, "test-site")
	if err != nil {
		t.Fatalf("DeploySite failed: %v", err)
	}

	if res.FileCount != 3 {
		t.Errorf("Expected 3 files, got %d", res.FileCount)
	}

	// Verify files in DB
	fs := GetFileSystem()
	
	// Check index.html
	exists, err := fs.Exists("test-site", "index.html")
	if !exists || err != nil {
		t.Error("index.html not found in VFS")
	}

	// Check subdirectory file
	exists, err = fs.Exists("test-site", "css/style.css")
	if !exists || err != nil {
		t.Error("css/style.css not found in VFS")
	}
}

func TestSiteExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	Init(db)
	fs := GetFileSystem()

	// Case 1: No site
	if SiteExists("ghost") {
		t.Error("SiteExists returned true for non-existent site")
	}

	// Case 2: Static site (index.html)
	fs.WriteFile("static", "index.html", strings.NewReader("hi"), 2, "text/html")
	if !SiteExists("static") {
		t.Error("SiteExists returned false for static site")
	}

	// Case 3: Serverless site (main.js only)
	fs.WriteFile("app", "main.js", strings.NewReader("code"), 4, "text/javascript")
	if !SiteExists("app") {
		t.Error("SiteExists returned false for serverless app")
	}
}
