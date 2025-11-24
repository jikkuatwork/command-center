package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth AuthConfig     `json:"auth"`
	Ntfy NtfyConfig     `json:"ntfy"`
	APIKey APIKeyConfig `json:"api_key,omitempty"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port   string `json:"port"`
	Domain string `json:"domain"`
	Env    string `json:"env"` // development/production
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"` // bcrypt hash
}

// NtfyConfig holds notification configuration
type NtfyConfig struct {
	Topic string `json:"topic"`
	URL   string `json:"url"`
}

// APIKeyConfig holds API key configuration for deployment
type APIKeyConfig struct {
	Token string `json:"token,omitempty"`
	Name  string `json:"name,omitempty"`
}

var appConfig *Config

// CLIFlags holds command-line flags
type CLIFlags struct {
	ConfigPath string
	DBPath     string
	Port       string
	Username   string
	Password   string
}

// ParseFlags parses command-line flags (for backward compatibility)
func ParseFlags() *CLIFlags {
	flags := &CLIFlags{}

	// Get default config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultConfigPath := filepath.Join(homeDir, ".config", "fazt", "config.json")
	defaultDBPath := filepath.Join(homeDir, ".config", "fazt", "data.db")

	flag.StringVar(&flags.ConfigPath, "config", defaultConfigPath, "Path to config file")
	flag.StringVar(&flags.DBPath, "db", "", "Database file path (overrides config)")
	flag.StringVar(&flags.Port, "port", "", "Server port (overrides config)")
	flag.StringVar(&flags.Username, "username", "", "Set/update username (updates config)")
	flag.StringVar(&flags.Password, "password", "", "Set/update password (updates config)")

	flag.Parse()

	// If no db path provided via flag, use default
	if flags.DBPath == "" {
		flags.DBPath = defaultDBPath
	}

	return flags
}

// Load reads configuration from multiple sources with priority:
// 1. CLI flags (highest)
// 2. JSON config file
// 3. Environment variables
// 4. Built-in defaults (lowest)
func Load(flags *CLIFlags) (*Config, error) {
	if appConfig != nil {
		return appConfig, nil
	}

	// Expand home directory in paths
	configPath := ExpandPath(flags.ConfigPath)

	// Try to load from JSON file
	cfg, err := LoadFromFile(configPath)
	if err != nil {
		// If file doesn't exist, create default config
		if os.IsNotExist(err) {
			log.Printf("Config file not found at %s, creating default config...", configPath)
			cfg = CreateDefaultConfig()
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Apply environment variables (backward compatibility)
	applyEnvVars(cfg)

	// Apply CLI flags (highest priority)
	applyCLIFlags(cfg, flags)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	appConfig = cfg
	log.Printf("Configuration loaded: Environment=%s, Port=%s, Auth=required",
		cfg.Server.Env, cfg.Server.Port)

	return appConfig, nil
}

// LoadFromFile loads configuration from a JSON file (exported for use in main.go)
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// SaveToFile saves configuration to a JSON file
func SaveToFile(cfg *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with restrictive permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Config saved to %s", path)
	return nil
}

// CreateDefaultConfig creates a default configuration (exported for use in main.go)
func CreateDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(homeDir, ".config", "fazt", "data.db")

	return &Config{
		Server: ServerConfig{
			Port:   "4698",
			Domain: "https://fazt.sh",
			Env:    "development",
		},
		Database: DatabaseConfig{
			Path: defaultDBPath,
		},
		Auth: AuthConfig{
			Username:     "",
			PasswordHash: "",
		},
		Ntfy: NtfyConfig{
			Topic: "",
			URL:   "https://ntfy.sh",
		},
	}
}

// applyEnvVars applies environment variables to config (backward compatibility)
func applyEnvVars(cfg *Config) {
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if env := os.Getenv("ENV"); env != "" {
		cfg.Server.Env = env
	}
	if domain := os.Getenv("FAZT_DOMAIN"); domain != "" {
		cfg.Server.Domain = domain
	}
	if ntfyTopic := os.Getenv("NTFY_TOPIC"); ntfyTopic != "" {
		cfg.Ntfy.Topic = ntfyTopic
	}
	if ntfyURL := os.Getenv("NTFY_URL"); ntfyURL != "" {
		cfg.Ntfy.URL = ntfyURL
	}
}

// applyCLIFlags applies CLI flags to config (highest priority)
func applyCLIFlags(cfg *Config, flags *CLIFlags) {
	if flags.Port != "" {
		cfg.Server.Port = flags.Port
	}
	if flags.DBPath != "" {
		cfg.Database.Path = ExpandPath(flags.DBPath)
	}
}

// ExpandPath expands ~ to home directory (exported for use in main.go)
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port is a number between 1-65535
	port, err := strconv.Atoi(c.Server.Port)
	if err != nil {
		return fmt.Errorf("invalid port: %s (must be a number)", c.Server.Port)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", port)
	}

	// Validate environment
	if c.Server.Env != "development" && c.Server.Env != "production" {
		return fmt.Errorf("invalid environment: %s (must be 'development' or 'production')", c.Server.Env)
	}

	// Ensure DB path is set
	if c.Database.Path == "" {
		return errors.New("database path cannot be empty")
	}

	// Expand database path
	c.Database.Path = ExpandPath(c.Database.Path)

	// Validate auth config (v0.4.0: auth always required)
	if c.Auth.Username == "" {
		return errors.New("auth username is required")
	}
	if c.Auth.PasswordHash == "" {
		return errors.New("auth password hash is required")
	}

	return nil
}

// Get returns the loaded configuration
func Get() *Config {
	if appConfig == nil {
		log.Fatal("Configuration not loaded. Call Load() first.")
	}
	return appConfig
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// Legacy helpers for backward compatibility
func (c *Config) Port() string {
	return c.Server.Port
}

func (c *Config) DBPath() string {
	return c.Database.Path
}

func (c *Config) NtfyTopic() string {
	return c.Ntfy.Topic
}

func (c *Config) NtfyURL() string {
	return c.Ntfy.URL
}

func (c *Config) Environment() string {
	return c.Server.Env
}

// GetAPIKey returns the stored API key token
func (c *Config) GetAPIKey() string {
	return c.APIKey.Token
}

// SetAPIKey stores the API key token and name in config
func (c *Config) SetAPIKey(token, name string) {
	c.APIKey.Token = token
	c.APIKey.Name = name
}
