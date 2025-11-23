package hosting

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{"valid simple", "mysite", false},
		{"valid with numbers", "site123", false},
		{"valid with hyphen", "my-site", false},
		{"empty", "", true},
		{"too long", "this-is-a-very-long-subdomain-name-that-exceeds-sixty-three-characters-limit", true},
		{"starts with hyphen", "-mysite", true},
		{"ends with hyphen", "mysite-", true},
		{"double hyphen", "my--site", false}, // double hyphens are allowed
		{"uppercase", "MySite", false},     // converted to lowercase, so valid
		{"underscore", "my_site", true},
		{"dot", "my.site", true},
		{"space", "my site", true},
		{"special chars", "my@site", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubdomain(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestInitAndGetSitesDir(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Check sites directory was created
	sitesDir := GetSitesDir()
	if sitesDir == "" {
		t.Error("GetSitesDir() returned empty string")
	}

	expectedPath := filepath.Join(tmpDir, "sites")
	if sitesDir != expectedPath {
		t.Errorf("GetSitesDir() = %q, want %q", sitesDir, expectedPath)
	}

	// Check directory exists
	if _, err := os.Stat(sitesDir); os.IsNotExist(err) {
		t.Error("Sites directory was not created")
	}
}

func TestSiteOperations(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test creating a site
	siteName := "testsite"
	if err := CreateSite(siteName); err != nil {
		t.Fatalf("CreateSite() failed: %v", err)
	}

	// Check site exists
	if !SiteExists(siteName) {
		t.Error("SiteExists() returned false for created site")
	}

	// Check site directory
	siteDir := GetSiteDir(siteName)
	expectedDir := filepath.Join(tmpDir, "sites", siteName)
	if siteDir != expectedDir {
		t.Errorf("GetSiteDir() = %q, want %q", siteDir, expectedDir)
	}

	// Test listing sites
	sites, err := ListSites()
	if err != nil {
		t.Fatalf("ListSites() failed: %v", err)
	}
	if len(sites) != 1 || sites[0].Name != siteName {
		t.Errorf("ListSites() = %v, want [%s]", sites, siteName)
	}

	// Test deleting a site
	if err := DeleteSite(siteName); err != nil {
		t.Fatalf("DeleteSite() failed: %v", err)
	}

	// Verify site no longer exists
	if SiteExists(siteName) {
		t.Error("SiteExists() returned true for deleted site")
	}
}

func TestSiteExistsForNonexistent(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if SiteExists("nonexistent") {
		t.Error("SiteExists() returned true for nonexistent site")
	}
}

func TestHasServerless(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "cc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize hosting
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Create a site without main.js
	if err := CreateSite("static"); err != nil {
		t.Fatalf("CreateSite() failed: %v", err)
	}

	staticDir := GetSiteDir("static")
	if HasServerless(staticDir) {
		t.Error("HasServerless() returned true for site without main.js")
	}

	// Create a site with main.js
	if err := CreateSite("serverless"); err != nil {
		t.Fatalf("CreateSite() failed: %v", err)
	}

	serverlessDir := GetSiteDir("serverless")
	mainJS := filepath.Join(serverlessDir, "main.js")
	if err := os.WriteFile(mainJS, []byte("res.send('hello');"), 0644); err != nil {
		t.Fatalf("Failed to write main.js: %v", err)
	}

	if !HasServerless(serverlessDir) {
		t.Error("HasServerless() returned false for site with main.js")
	}
}
