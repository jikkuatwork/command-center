package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	DBPath      string
	NtfyTopic   string
	NtfyURL     string
	Environment string
}

var appConfig *Config

// Load reads configuration from environment variables with fallback defaults
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		Port:        getEnv("PORT", "4698"),
		DBPath:      getEnv("DB_PATH", "./cc.db"),
		NtfyTopic:   getEnv("NTFY_TOPIC", ""),
		NtfyURL:     getEnv("NTFY_URL", "https://ntfy.sh"),
		Environment: getEnv("ENV", "development"),
	}

	// Validate configuration
	if err := appConfig.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded: Environment=%s, Port=%s", appConfig.Environment, appConfig.Port)
	return appConfig
}

// Get returns the loaded configuration
func Get() *Config {
	if appConfig == nil {
		return Load()
	}
	return appConfig
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port is a number
	if _, err := strconv.Atoi(c.Port); err != nil {
		log.Printf("Warning: PORT is not a valid number, using default 4698")
		c.Port = "4698"
	}

	// Ensure DB path is set
	if c.DBPath == "" {
		log.Printf("Warning: DB_PATH not set, using default ./cc.db")
		c.DBPath = "./cc.db"
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
