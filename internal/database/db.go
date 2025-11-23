package database

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init initializes the database connection with WAL mode
func Init(dbPath string) error {
	var err error

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// runMigrations executes SQL migration files in order
func runMigrations() error {
	// Create migrations tracking table if it doesn't exist
	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version INTEGER UNIQUE NOT NULL,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations
	migrations := []struct {
		version int
		name    string
		file    string
	}{
		{1, "initial_schema", "./migrations/001_initial.sql"},
		{2, "paas_tables", "./migrations/002_paas.sql"},
		{3, "env_vars", "./migrations/003_env_vars.sql"},
	}

	// Run each migration if not already applied
	for _, migration := range migrations {
		// Check if migration has been applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = ?", migration.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("Migration %d (%s) already applied, skipping", migration.version, migration.name)
			continue
		}

		// Read migration file
		migrationSQL, err := os.ReadFile(migration.file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migration.file, err)
		}

		// Execute migration
		if _, err := db.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration %d (%s): %w", migration.version, migration.name, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO migrations (version, name) VALUES (?, ?)", migration.version, migration.name)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		log.Printf("Applied migration %d: %s", migration.version, migration.name)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// HealthCheck verifies the database connection is working
func HealthCheck() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}

// Backup creates a backup of the database
func Backup(dbPath string) (string, error) {
	// Create backup directory
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup_%s.db", timestamp))

	// Copy database file
	srcFile, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy database: %w", err)
	}

	log.Printf("Database backup created: %s", backupPath)

	// Cleanup old backups (keep last 5)
	if err := cleanupOldBackups(backupDir, 5); err != nil {
		log.Printf("Warning: failed to cleanup old backups: %v", err)
	}

	return backupPath, nil
}

// cleanupOldBackups removes old backup files, keeping only the most recent N
func cleanupOldBackups(backupDir string, keep int) error {
	files, err := filepath.Glob(filepath.Join(backupDir, "backup_*.db"))
	if err != nil {
		return err
	}

	// If we have fewer backups than the keep limit, nothing to do
	if len(files) <= keep {
		return nil
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: file, modTime: info.ModTime()})
	}

	// Sort by modification time
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.After(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Delete old backups
	deleteCount := len(fileInfos) - keep
	for i := 0; i < deleteCount; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			log.Printf("Warning: failed to remove old backup %s: %v", fileInfos[i].path, err)
		} else {
			log.Printf("Removed old backup: %s", fileInfos[i].path)
		}
	}

	return nil
}
