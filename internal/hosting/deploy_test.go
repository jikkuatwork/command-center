package hosting

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func createTestZip(files map[string]string) (*zip.Reader, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
}

func TestDeploySite(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-deploy-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Create test ZIP
	zipReader, err := createTestZip(map[string]string{
		"index.html": "<h1>Hello</h1>",
		"style.css":  "body { color: red; }",
		"js/app.js":  "console.log('hi');",
	})
	if err != nil {
		t.Fatalf("Failed to create test zip: %v", err)
	}

	// Deploy
	result, err := DeploySite(zipReader, "testsite")
	if err != nil {
		t.Fatalf("DeploySite() failed: %v", err)
	}

	// Verify result
	if result.SiteID != "testsite" {
		t.Errorf("result.SiteID = %q, want %q", result.SiteID, "testsite")
	}
	if result.FileCount != 3 {
		t.Errorf("result.FileCount = %d, want 3", result.FileCount)
	}

	// Verify files exist
	siteDir := GetSiteDir("testsite")
	files := []string{"index.html", "style.css", "js/app.js"}
	for _, f := range files {
		path := filepath.Join(siteDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s was not created", f)
		}
	}
}

func TestDeploySitePathTraversal(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-deploy-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Create ZIP with path traversal attempt
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Try to create a file outside the site directory
	f, _ := w.Create("../../../etc/passwd")
	f.Write([]byte("malicious content"))

	// Also add a normal file
	f2, _ := w.Create("index.html")
	f2.Write([]byte("<h1>Normal</h1>"))

	w.Close()

	zipReader, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))

	// Deploy should succeed but skip the malicious file
	result, err := DeploySite(zipReader, "safeside")
	if err != nil {
		t.Fatalf("DeploySite() failed: %v", err)
	}

	// Only the safe file should be deployed
	if result.FileCount != 1 {
		t.Errorf("result.FileCount = %d, want 1 (malicious file should be skipped)", result.FileCount)
	}

	// Verify malicious file was not created
	maliciousPath := filepath.Join(tmpDir, "etc", "passwd")
	if _, err := os.Stat(maliciousPath); !os.IsNotExist(err) {
		t.Error("Malicious file was created - path traversal not blocked!")
	}
}

func TestDeploySiteInvalidSubdomain(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-deploy-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	zipReader, _ := createTestZip(map[string]string{"index.html": "test"})

	// Note: "My-Site" is actually valid because it's lowercased to "my-site"
	invalidNames := []string{"", "../bad", "test.site", "test_site"}
	for _, name := range invalidNames {
		_, err := DeploySite(zipReader, name)
		if err == nil {
			t.Errorf("DeploySite(%q) should have failed", name)
		}
	}
}

func TestAPIKeyOperations(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create api_keys table
	_, err = db.Exec(`
		CREATE TABLE api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			scopes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create API key
	token, err := CreateAPIKey(db, "test-key", "deploy")
	if err != nil {
		t.Fatalf("CreateAPIKey() failed: %v", err)
	}
	if token == "" {
		t.Error("CreateAPIKey() returned empty token")
	}

	// Validate the key
	id, name, err := ValidateAPIKey(db, token)
	if err != nil {
		t.Fatalf("ValidateAPIKey() failed: %v", err)
	}
	if name != "test-key" {
		t.Errorf("name = %q, want %q", name, "test-key")
	}
	if id == 0 {
		t.Error("id should not be 0")
	}

	// Validate with wrong key
	_, _, err = ValidateAPIKey(db, "wrong-token")
	if err == nil {
		t.Error("ValidateAPIKey() should fail with wrong token")
	}

	// List keys
	keys, err := ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys() failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("len(keys) = %d, want 1", len(keys))
	}

	// Delete key
	if err := DeleteAPIKey(db, id); err != nil {
		t.Fatalf("DeleteAPIKey() failed: %v", err)
	}

	// Verify deleted
	keys, _ = ListAPIKeys(db)
	if len(keys) != 0 {
		t.Errorf("len(keys) = %d, want 0 after delete", len(keys))
	}
}
