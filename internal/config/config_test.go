package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"home prefix", "~/test/path", filepath.Join(homeDir, "test/path")},
		{"absolute path", "/etc/config", "/etc/config"},
		{"relative path", "relative/path", "relative/path"},
		{"empty string", "", ""},
		{"just tilde", "~", "~"}, // Only ~/... is expanded
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid development config",
			config: Config{
				Server:   ServerConfig{Port: "8080", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
				Auth:     AuthConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "valid production config with auth",
			config: Config{
				Server:   ServerConfig{Port: "443", Domain: "example.com", Env: "production"},
				Database: DatabaseConfig{Path: "/var/data/app.db"},
				Auth:     AuthConfig{Enabled: true, Username: "admin", PasswordHash: "hash123"},
			},
			wantErr: false,
		},
		{
			name: "invalid port - not a number",
			config: Config{
				Server:   ServerConfig{Port: "abc", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Server:   ServerConfig{Port: "70000", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid port - zero",
			config: Config{
				Server:   ServerConfig{Port: "0", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid environment",
			config: Config{
				Server:   ServerConfig{Port: "8080", Domain: "localhost", Env: "staging"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "empty database path",
			config: Config{
				Server:   ServerConfig{Port: "8080", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: ""},
			},
			wantErr: true,
			errMsg:  "database path cannot be empty",
		},
		{
			name: "auth enabled but missing username",
			config: Config{
				Server:   ServerConfig{Port: "8080", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
				Auth:     AuthConfig{Enabled: true, Username: "", PasswordHash: "hash"},
			},
			wantErr: true,
			errMsg:  "username is empty",
		},
		{
			name: "auth enabled but missing password hash",
			config: Config{
				Server:   ServerConfig{Port: "8080", Domain: "localhost", Env: "development"},
				Database: DatabaseConfig{Path: "/tmp/test.db"},
				Auth:     AuthConfig{Enabled: true, Username: "admin", PasswordHash: ""},
			},
			wantErr: true,
			errMsg:  "password hash is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigEnvironmentMethods(t *testing.T) {
	devConfig := &Config{Server: ServerConfig{Env: "development"}}
	prodConfig := &Config{Server: ServerConfig{Env: "production"}}

	if !devConfig.IsDevelopment() {
		t.Error("IsDevelopment() should return true for development env")
	}
	if devConfig.IsProduction() {
		t.Error("IsProduction() should return false for development env")
	}

	if prodConfig.IsDevelopment() {
		t.Error("IsDevelopment() should return false for production env")
	}
	if !prodConfig.IsProduction() {
		t.Error("IsProduction() should return true for production env")
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	cfg := CreateDefaultConfig()

	if cfg.Server.Port != "4698" {
		t.Errorf("Default port should be 4698, got %s", cfg.Server.Port)
	}
	if cfg.Server.Env != "development" {
		t.Errorf("Default env should be development, got %s", cfg.Server.Env)
	}
	if cfg.Auth.Enabled {
		t.Error("Auth should be disabled by default")
	}
	if cfg.Ntfy.URL != "https://ntfy.sh" {
		t.Errorf("Default ntfy URL should be https://ntfy.sh, got %s", cfg.Ntfy.URL)
	}
	if cfg.Database.Path == "" {
		t.Error("Default database path should not be empty")
	}
}

func TestLegacyHelpers(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: "9000", Env: "production"},
		Database: DatabaseConfig{Path: "/data/db"},
		Ntfy:     NtfyConfig{Topic: "alerts", URL: "https://ntfy.example.com"},
	}

	if cfg.Port() != "9000" {
		t.Errorf("Port() = %s, want 9000", cfg.Port())
	}
	if cfg.DBPath() != "/data/db" {
		t.Errorf("DBPath() = %s, want /data/db", cfg.DBPath())
	}
	if cfg.NtfyTopic() != "alerts" {
		t.Errorf("NtfyTopic() = %s, want alerts", cfg.NtfyTopic())
	}
	if cfg.NtfyURL() != "https://ntfy.example.com" {
		t.Errorf("NtfyURL() = %s, want https://ntfy.example.com", cfg.NtfyURL())
	}
	if cfg.Environment() != "production" {
		t.Errorf("Environment() = %s, want production", cfg.Environment())
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {"port": "8080", "domain": "test.com", "env": "development"},
		"database": {"path": "/tmp/test.db"},
		"auth": {"enabled": false}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Port = %s, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Domain != "test.com" {
		t.Errorf("Domain = %s, want test.com", cfg.Server.Domain)
	}
}

func TestLoadFromFileErrors(t *testing.T) {
	// Non-existent file
	_, err := LoadFromFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadFromFile() should error for non-existent file")
	}

	// Invalid JSON
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(badPath, []byte("not valid json"), 0644)

	_, err = LoadFromFile(badPath)
	if err == nil {
		t.Error("LoadFromFile() should error for invalid JSON")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.json")

	cfg := &Config{
		Server:   ServerConfig{Port: "9999", Domain: "saved.com", Env: "production"},
		Database: DatabaseConfig{Path: "/saved/db"},
	}

	if err := SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	loaded, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.Server.Port != "9999" {
		t.Errorf("Loaded port = %s, want 9999", loaded.Server.Port)
	}
}
