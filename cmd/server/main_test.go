package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jikku/command-center/internal/config"
	"golang.org/x/crypto/bcrypt"
)

/*
===================================================================================
Test Suite for CLI Refactor (v0.4.0)
===================================================================================

This test suite provides comprehensive coverage for the new CLI commands:
- initCommand: Initialize server configuration
- setCredentialsCommand: Update username/password
- setConfigCommand: Update server settings
- statusCommand: Display server status

Implementation Guide:
--------------------

To implement these features, create the following functions in main.go:

1. initCommand(username, password, domain, port, env, configPath string) error
   - Checks if config already exists (return error if it does)
   - Creates config with provided values
   - Hashes password with bcrypt (cost 12)
   - Sets secure permissions (0600 on config file)
   - Returns nil on success

2. setCredentialsCommand(username, password, configPath string) error
   - Loads existing config (return error if not found)
   - Requires at least one of username or password
   - Updates provided fields only
   - Hashes new password if provided
   - Saves config back
   - Returns nil on success

3. setConfigCommand(domain, port, env, configPath string) error
   - Loads existing config (return error if not found)
   - Requires at least one field to update
   - Validates port (1-65535) and env (development|production)
   - Updates provided fields only
   - Saves config back
   - Returns nil on success

4. statusCommand(configPath, configDir string) (string, error)
   - Loads config (return error if not found)
   - Checks if server is running (reads PID file)
   - Formats and returns status string
   - Returns formatted output and nil on success

5. Update handleServerCommand() switch to include:
   case "init": call initCommand with parsed flags
   case "set-config": call setConfigCommand with parsed flags
   case "status": call statusCommand and print result

6. Update main() switch to include:
   case "deploy": handleDeployCommand() // alias

All tests will fail initially. Implement the functions to make them pass.
*/

// Test helpers

// createTempConfigDir creates a temporary config directory for testing
func createTempConfigDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "fazt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// loadConfigFromFile reads and parses config.json
func loadConfigFromFile(t *testing.T, path string) *config.Config {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	return &cfg
}

// createTestConfig creates a config file for testing
func createTestConfig(t *testing.T, dir string, cfg *config.Config) string {
	t.Helper()
	configPath := filepath.Join(dir, "config.json")
	if err := config.SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	return configPath
}

// ===================================================================================
// Init Command Tests
// ===================================================================================

// TestInitCommand_Success tests successful initialization
func TestInitCommand_Success(t *testing.T) {
	// Setup
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	// Test data
	username := "testadmin"
	password := "testpass123"
	domain := "https://test.example.com"
	port := "4698"
	env := "development"

	// Execute init command
	err := initCommand(username, password, domain, port, env, configPath)
	if err != nil {
		t.Fatalf("initCommand failed: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Verify file permissions (should be 0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Config file has incorrect permissions: got %o, want 0600", info.Mode().Perm())
	}

	// Load and verify config contents
	cfg := loadConfigFromFile(t, configPath)

	if cfg.Server.Domain != domain {
		t.Errorf("Domain mismatch: got %s, want %s", cfg.Server.Domain, domain)
	}
	if cfg.Server.Port != port {
		t.Errorf("Port mismatch: got %s, want %s", cfg.Server.Port, port)
	}
	if cfg.Server.Env != env {
		t.Errorf("Env mismatch: got %s, want %s", cfg.Server.Env, env)
	}
	if cfg.Auth.Username != username {
		t.Errorf("Username mismatch: got %s, want %s", cfg.Auth.Username, username)
	}

	// Verify password is hashed (not plaintext)
	if cfg.Auth.PasswordHash == password {
		t.Error("Password was not hashed")
	}

	// Verify password hash is valid bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(cfg.Auth.PasswordHash), []byte(password))
	if err != nil {
		t.Errorf("Password hash verification failed: %v", err)
	}

	// Verify database path is set
	expectedDBPath := filepath.Join(tmpDir, "data.db")
	if cfg.Database.Path != expectedDBPath {
		t.Errorf("Database path mismatch: got %s, want %s", cfg.Database.Path, expectedDBPath)
	}
}

func TestInitCommand_ConfigAlreadyExists(t *testing.T) {
	// Setup
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	// Create existing config
	existingConfig := &config.Config{
		Server: config.ServerConfig{
			Port:   "5000",
			Domain: "https://existing.com",
			Env:    "production",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(tmpDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     "existing",
			PasswordHash: "existinghash",
		},
	}
	createTestConfig(t, tmpDir, existingConfig)

	// Try to init again
	err := initCommand("admin", "pass123", "https://test.com", "4698", "development", configPath)
	if err == nil {
		t.Fatal("initCommand should fail when config exists")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already initialized") &&
	   !strings.Contains(strings.ToLower(err.Error()), "exists") {
		t.Errorf("Error should mention 'already initialized' or 'exists', got: %v", err)
	}

	// Verify original config was not modified
	cfg := loadConfigFromFile(t, configPath)
	if cfg.Auth.Username != "existing" {
		t.Error("Existing config was modified when it shouldn't have been")
	}
	if cfg.Server.Port != "5000" {
		t.Error("Existing config port was modified")
	}
}

func TestInitCommand_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		domain   string
		wantErr  bool
	}{
		{"missing username", "", "pass123", "https://test.com", true},
		{"missing password", "admin", "", "https://test.com", true},
		{"missing domain", "admin", "pass123", "", true},
		{"all provided", "admin", "pass123", "https://test.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTempConfigDir(t)
			configPath := filepath.Join(tmpDir, "config.json")

			err := initCommand(tt.username, tt.password, tt.domain, "4698", "development", configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("initCommand() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Error message should mention which field is missing
				errMsg := strings.ToLower(err.Error())
				if !strings.Contains(errMsg, "required") && !strings.Contains(errMsg, "missing") && !strings.Contains(errMsg, "empty") {
					t.Errorf("Error should mention required field, got: %v", err)
				}
			}
		})
	}
}

func TestInitCommand_InvalidPort(t *testing.T) {
	tests := []struct {
		port    string
		wantErr bool
	}{
		{"0", true},      // Too low
		{"65536", true},  // Too high
		{"abc", true},    // Not a number
		{"-100", true},   // Negative
		{"4698", false},  // Valid
		{"8080", false},  // Valid
		{"443", false},   // Valid
	}

	for _, tt := range tests {
		t.Run("port="+tt.port, func(t *testing.T) {
			tmpDir := createTempConfigDir(t)
			configPath := filepath.Join(tmpDir, "config.json")

			err := initCommand("admin", "pass123", "https://test.com", tt.port, "development", configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("initCommand() with port %s: error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestInitCommand_InvalidEnvironment(t *testing.T) {
	tests := []struct {
		env     string
		wantErr bool
	}{
		{"prod", true},        // Should be "production"
		{"dev", true},         // Should be "development"
		{"staging", true},     // Not supported
		{"test", true},        // Not supported
		{"development", false}, // Valid
		{"production", false},  // Valid
	}

	for _, tt := range tests {
		t.Run("env="+tt.env, func(t *testing.T) {
			tmpDir := createTempConfigDir(t)
			configPath := filepath.Join(tmpDir, "config.json")

			err := initCommand("admin", "pass123", "https://test.com", "4698", tt.env, configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("initCommand() with env %s: error = %v, wantErr %v", tt.env, err, tt.wantErr)
			}
		})
	}
}

func TestInitCommand_SecurePermissions(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	// Ensure directory has secure permissions
	os.MkdirAll(tmpDir, 0700)

	err := initCommand("admin", "pass123", "https://test.com", "4698", "development", configPath)
	if err != nil {
		t.Fatalf("initCommand failed: %v", err)
	}

	// Check directory permissions
	dirInfo, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("Directory permissions = %o, want 0700", dirInfo.Mode().Perm())
	}

	// Check file permissions
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("File permissions = %o, want 0600", fileInfo.Mode().Perm())
	}
}

// ===================================================================================
// Set-Credentials Command Tests
// ===================================================================================

func TestSetCredentials_UpdatePassword(t *testing.T) {
	// Setup - create existing config
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	oldPassword := "oldpass123"
	oldHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "4698",
			Domain: "https://test.com",
			Env:    "development",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(tmpDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: string(oldHash),
		},
	}
	createTestConfig(t, tmpDir, cfg)

	// Update password
	newPassword := "newpass456"
	err := setCredentialsCommand("", newPassword, configPath)
	if err != nil {
		t.Fatalf("setCredentialsCommand failed: %v", err)
	}

	// Verify password was updated
	updatedCfg := loadConfigFromFile(t, configPath)
	err = bcrypt.CompareHashAndPassword([]byte(updatedCfg.Auth.PasswordHash), []byte(newPassword))
	if err != nil {
		t.Error("New password hash verification failed")
	}

	// Verify old password no longer works
	err = bcrypt.CompareHashAndPassword([]byte(updatedCfg.Auth.PasswordHash), []byte(oldPassword))
	if err == nil {
		t.Error("Old password still works (password was not updated)")
	}

	// Verify username was preserved
	if updatedCfg.Auth.Username != "admin" {
		t.Error("Username was changed when it shouldn't have been")
	}

	// Verify other config fields were preserved
	if updatedCfg.Server.Port != "4698" {
		t.Error("Port was changed when it shouldn't have been")
	}
	if updatedCfg.Server.Domain != "https://test.com" {
		t.Error("Domain was changed when it shouldn't have been")
	}
}

func TestSetCredentials_UpdateUsername(t *testing.T) {
	// Setup
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "4698",
			Domain: "https://test.com",
			Env:    "development",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(tmpDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     "oldadmin",
			PasswordHash: string(oldHash),
		},
	}
	createTestConfig(t, tmpDir, cfg)

	// Update username
	newUsername := "newadmin"
	err := setCredentialsCommand(newUsername, "", configPath)
	if err != nil {
		t.Fatalf("setCredentialsCommand failed: %v", err)
	}

	// Verify username was updated
	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Auth.Username != newUsername {
		t.Errorf("Username not updated: got %s, want %s", updatedCfg.Auth.Username, newUsername)
	}

	// Verify password hash was preserved
	if updatedCfg.Auth.PasswordHash != string(oldHash) {
		t.Error("Password hash was changed when it shouldn't have been")
	}
}

func TestSetCredentials_UpdateBoth(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "oldadmin", PasswordHash: "oldhash"},
	}
	createTestConfig(t, tmpDir, cfg)

	// Update both
	err := setCredentialsCommand("newadmin", "newpass", configPath)
	if err != nil {
		t.Fatalf("setCredentialsCommand failed: %v", err)
	}

	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Auth.Username != "newadmin" {
		t.Error("Username was not updated")
	}
	// Password should be hashed
	err = bcrypt.CompareHashAndPassword([]byte(updatedCfg.Auth.PasswordHash), []byte("newpass"))
	if err != nil {
		t.Error("New password verification failed")
	}
}

func TestSetCredentials_NoConfigExists(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	// Config doesn't exist
	err := setCredentialsCommand("admin", "pass123", configPath)
	if err == nil {
		t.Fatal("setCredentialsCommand should fail when config doesn't exist")
	}
	errMsg := strings.ToLower(err.Error())
	if !strings.Contains(errMsg, "not found") && !strings.Contains(errMsg, "not initialized") && !strings.Contains(errMsg, "does not exist") {
		t.Errorf("Error should mention config not found, got: %v", err)
	}
}

func TestSetCredentials_NoFlagsProvided(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	// Call with no flags
	err := setCredentialsCommand("", "", configPath)
	if err == nil {
		t.Fatal("setCredentialsCommand should fail when no flags provided")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "at least one") &&
	   !strings.Contains(strings.ToLower(err.Error()), "required") {
		t.Errorf("Error should mention 'at least one flag', got: %v", err)
	}
}

// ===================================================================================
// Set-Config Command Tests
// ===================================================================================

func TestSetConfig_UpdateDomain(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "4698",
			Domain: "https://old.example.com",
			Env:    "development",
		},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	newDomain := "https://new.example.com"
	err := setConfigCommand(newDomain, "", "", configPath)
	if err != nil {
		t.Fatalf("setConfigCommand failed: %v", err)
	}

	// Verify domain was updated
	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Server.Domain != newDomain {
		t.Errorf("Domain not updated: got %s, want %s", updatedCfg.Server.Domain, newDomain)
	}

	// Verify other fields preserved
	if updatedCfg.Server.Port != "4698" {
		t.Error("Port was changed when it shouldn't have been")
	}
	if updatedCfg.Server.Env != "development" {
		t.Error("Environment was changed when it shouldn't have been")
	}
	if updatedCfg.Auth.Username != "admin" {
		t.Error("Username was changed when it shouldn't have been")
	}
}

func TestSetConfig_UpdatePort(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	newPort := "8080"
	err := setConfigCommand("", newPort, "", configPath)
	if err != nil {
		t.Fatalf("setConfigCommand failed: %v", err)
	}

	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Server.Port != newPort {
		t.Errorf("Port not updated: got %s, want %s", updatedCfg.Server.Port, newPort)
	}
}

func TestSetConfig_UpdateEnvironment(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	newEnv := "production"
	err := setConfigCommand("", "", newEnv, configPath)
	if err != nil {
		t.Fatalf("setConfigCommand failed: %v", err)
	}

	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Server.Env != newEnv {
		t.Errorf("Environment not updated: got %s, want %s", updatedCfg.Server.Env, newEnv)
	}
}

func TestSetConfig_UpdateMultipleFields(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://old.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	newDomain := "https://prod.example.com"
	newPort := "443"
	newEnv := "production"

	err := setConfigCommand(newDomain, newPort, newEnv, configPath)
	if err != nil {
		t.Fatalf("setConfigCommand failed: %v", err)
	}

	updatedCfg := loadConfigFromFile(t, configPath)
	if updatedCfg.Server.Domain != newDomain {
		t.Errorf("Domain not updated: got %s, want %s", updatedCfg.Server.Domain, newDomain)
	}
	if updatedCfg.Server.Port != newPort {
		t.Errorf("Port not updated: got %s, want %s", updatedCfg.Server.Port, newPort)
	}
	if updatedCfg.Server.Env != newEnv {
		t.Errorf("Environment not updated: got %s, want %s", updatedCfg.Server.Env, newEnv)
	}
}

func TestSetConfig_NoFlagsProvided(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	err := setConfigCommand("", "", "", configPath)
	if err == nil {
		t.Fatal("setConfigCommand should fail when no flags provided")
	}
}

func TestSetConfig_InvalidPort(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	err := setConfigCommand("", "99999", "", configPath)
	if err == nil {
		t.Fatal("setConfigCommand should fail with invalid port")
	}
}

func TestSetConfig_InvalidEnvironment(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	err := setConfigCommand("", "", "staging", configPath)
	if err == nil {
		t.Fatal("setConfigCommand should fail with invalid environment")
	}
}

func TestSetConfig_NoConfigExists(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	err := setConfigCommand("https://test.com", "", "", configPath)
	if err == nil {
		t.Fatal("setConfigCommand should fail when config doesn't exist")
	}
}

// ===================================================================================
// Status Command Tests
// ===================================================================================

func TestStatus_OutputFormat(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "4698",
			Domain: "https://test.example.com",
			Env:    "production",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(tmpDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: "hash",
		},
	}
	createTestConfig(t, tmpDir, cfg)

	// Create a fake database file
	os.WriteFile(cfg.Database.Path, []byte("fake db"), 0600)

	output, err := statusCommand(configPath, tmpDir)
	if err != nil {
		t.Fatalf("statusCommand failed: %v", err)
	}

	// Verify output contains expected information
	expectedStrings := []string{
		"Server Status",
		"Config:",
		"Domain:",
		"https://test.example.com",
		"Port:",
		"4698",
		"Environment:",
		"production",
		"Username:",
		"admin",
		"Database:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Status output missing '%s'\nGot output:\n%s", expected, output)
		}
	}
}

func TestStatus_ServerRunning(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	// Create fake PID file
	pidFile := filepath.Join(tmpDir, "cc-server.pid")
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0600)

	output, err := statusCommand(configPath, tmpDir)
	if err != nil {
		t.Fatalf("statusCommand failed: %v", err)
	}

	if !strings.Contains(output, "Running") || !strings.Contains(output, fmt.Sprintf("%d", os.Getpid())) {
		t.Errorf("Status should indicate server is running with PID\nGot: %s", output)
	}
}

func TestStatus_ServerStopped(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: "4698", Domain: "https://test.com", Env: "development"},
		Database: config.DatabaseConfig{Path: filepath.Join(tmpDir, "data.db")},
		Auth: config.AuthConfig{Username: "admin", PasswordHash: "hash"},
	}
	createTestConfig(t, tmpDir, cfg)

	// No PID file = server not running

	output, err := statusCommand(configPath, tmpDir)
	if err != nil {
		t.Fatalf("statusCommand failed: %v", err)
	}

	if !strings.Contains(output, "Not running") && !strings.Contains(output, "not running") {
		t.Errorf("Status should indicate server is not running\nGot: %s", output)
	}
}

func TestStatus_NoConfigExists(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	_, err := statusCommand(configPath, tmpDir)
	if err == nil {
		t.Fatal("statusCommand should fail when config doesn't exist")
	}
}

// ===================================================================================
// Integration-like Tests
// ===================================================================================

func TestFullWorkflow_InitSetConfigStatus(t *testing.T) {
	tmpDir := createTempConfigDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	// 1. Init
	err := initCommand("admin", "pass123", "https://test.com", "4698", "development", configPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 2. Verify init worked
	cfg := loadConfigFromFile(t, configPath)
	if cfg.Server.Domain != "https://test.com" {
		t.Error("Init didn't set domain correctly")
	}

	// 3. Update credentials
	err = setCredentialsCommand("newadmin", "newpass", configPath)
	if err != nil {
		t.Fatalf("set-credentials failed: %v", err)
	}

	// 4. Update config
	err = setConfigCommand("https://new.com", "8080", "production", configPath)
	if err != nil {
		t.Fatalf("set-config failed: %v", err)
	}

	// 5. Verify updates worked
	cfg = loadConfigFromFile(t, configPath)
	if cfg.Server.Domain != "https://new.com" {
		t.Error("set-config didn't update domain")
	}
	if cfg.Server.Port != "8080" {
		t.Error("set-config didn't update port")
	}
	if cfg.Server.Env != "production" {
		t.Error("set-config didn't update environment")
	}
	if cfg.Auth.Username != "newadmin" {
		t.Error("set-credentials didn't update username")
	}

	// 6. Check status
	output, err := statusCommand(configPath, tmpDir)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(output, "https://new.com") {
		t.Error("Status doesn't show updated domain")
	}
	if !strings.Contains(output, "8080") {
		t.Error("Status doesn't show updated port")
	}
	if !strings.Contains(output, "newadmin") {
		t.Error("Status doesn't show updated username")
	}
}

// ===================================================================================
// Run Instructions
// ===================================================================================

/*
To run these tests:

1. Run all tests:
   go test ./cmd/server -v

2. Run specific test:
   go test ./cmd/server -v -run TestInitCommand_Success

3. Run tests with coverage:
   go test ./cmd/server -v -cover

These tests will fail until you implement the required functions.
Start by implementing initCommand, then setCredentialsCommand,
then setConfigCommand, and finally statusCommand.

Each function should:
- Have clear error messages
- Validate inputs
- Handle edge cases
- Return descriptive errors

Good luck!
*/
