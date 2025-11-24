package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jikku/command-center/internal/audit"
	"github.com/jikku/command-center/internal/auth"
	"github.com/jikku/command-center/internal/config"
	"github.com/jikku/command-center/internal/database"
	"github.com/jikku/command-center/internal/handlers"
	"github.com/jikku/command-center/internal/hosting"
	"github.com/jikku/command-center/internal/middleware"
	"github.com/jikku/command-center/internal/security"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

const Version = "v0.4.0"

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	showHelp    = flag.Bool("help", false, "Show help and exit")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	quiet       = flag.Bool("quiet", false, "Quiet mode (errors only)")
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	// Handle help/version flags first
	if command == "--version" || command == "-version" {
		printVersion()
		return
	}
	if command == "--help" || command == "-help" || command == "-h" {
		printUsage()
		return
	}

	// Handle top-level subcommands
	switch command {
	case "server":
		handleServerCommand(os.Args[2:])
	case "client":
		handleClientCommand(os.Args[2:])
	case "deploy":
		handleDeployCommand() // Alias for client deploy
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// ===================================================================================
// CLI Command Functions (v0.4.0)
// ===================================================================================

// initCommand initializes server configuration for first-time setup
func initCommand(username, password, domain, port, env, configPath string) error {
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("Error: Server already initialized\nConfig exists at: %s", configPath)
	}

	// Validate required fields
	if username == "" || password == "" || domain == "" {
		return errors.New("Error: username, password, and domain are required")
	}

	// Validate port
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf("Error: invalid port '%s' (must be 1-65535)", port)
	}

	// Validate environment
	if env != "development" && env != "production" {
		return fmt.Errorf("Error: invalid environment '%s' (must be 'development' or 'production')", env)
	}

	// Hash password with bcrypt cost 12
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("Error: failed to hash password: %v", err)
	}

	// Create config directory with secure permissions
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("Error: failed to create config directory: %v", err)
	}

	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   port,
			Domain: domain,
			Env:    env,
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(configDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     username,
			PasswordHash: string(passwordHash),
		},
		Ntfy: config.NtfyConfig{
			Topic: "",
			URL:   "https://ntfy.sh",
		},
	}

	// Save config with secure permissions
	if err := config.SaveToFile(cfg, configPath); err != nil {
		return fmt.Errorf("Error: failed to save config: %v", err)
	}

	return nil
}

// setCredentialsCommand updates username and/or password in existing config
func setCredentialsCommand(username, password, configPath string) error {
	// Validate at least one field is provided
	if username == "" && password == "" {
		return errors.New("Error: at least one of --username or --password is required")
	}

	// Load existing config
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: Config not found at %s\nRun 'fazt server init' first", configPath)
		}
		return fmt.Errorf("Error: Failed to load config: %v", err)
	}

	// Update provided fields
	if username != "" {
		cfg.Auth.Username = username
	}
	if password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return fmt.Errorf("Error: Failed to hash password: %v", err)
		}
		cfg.Auth.PasswordHash = string(passwordHash)
	}

	// Save config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		return fmt.Errorf("Error: Failed to save config: %v", err)
	}

	return nil
}

// setConfigCommand updates server configuration settings
func setConfigCommand(domain, port, env, configPath string) error {
	// Validate at least one field is provided
	if domain == "" && port == "" && env == "" {
		return errors.New("Error: at least one of --domain, --port, or --env is required")
	}

	// Load existing config
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: Config not found at %s\nRun 'fazt server init' first", configPath)
		}
		return fmt.Errorf("Error: Failed to load config: %v", err)
	}

	// Validate and update port if provided
	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("Error: invalid port '%s' (must be 1-65535)", port)
		}
		cfg.Server.Port = port
	}

	// Validate and update environment if provided
	if env != "" {
		if env != "development" && env != "production" {
			return fmt.Errorf("Error: invalid environment '%s' (must be 'development' or 'production')", env)
		}
		cfg.Server.Env = env
	}

	// Update domain if provided
	if domain != "" {
		cfg.Server.Domain = domain
	}

	// Validate the updated config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("Error: Invalid configuration: %v", err)
	}

	// Save config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		return fmt.Errorf("Error: Failed to save config: %v", err)
	}

	return nil
}

// statusCommand displays current configuration and server status
func statusCommand(configPath, configDir string) (string, error) {
	// Load config
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return "", fmt.Errorf("Error: Config not found at %s\nRun 'fazt server init' first", configPath)
	}

	var output strings.Builder
	output.WriteString("Server Status\n")
	output.WriteString("═══════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("Config:       %s\n", configPath))
	output.WriteString(fmt.Sprintf("Domain:       %s\n", cfg.Server.Domain))
	output.WriteString(fmt.Sprintf("Port:         %s\n", cfg.Server.Port))
	output.WriteString(fmt.Sprintf("Environment:  %s\n", cfg.Server.Env))
	output.WriteString(fmt.Sprintf("Username:     %s\n", cfg.Auth.Username))

	// Check database file
	if stat, err := os.Stat(cfg.Database.Path); err == nil {
		size := float64(stat.Size()) / (1024 * 1024) // Convert to MB
		output.WriteString(fmt.Sprintf("Database:     %s (%.1f MB)\n", cfg.Database.Path, size))
	} else {
		output.WriteString(fmt.Sprintf("Database:     %s (not found)\n", cfg.Database.Path))
	}

	// Check sites directory
	sitesDir := filepath.Join(configDir, "sites")
	if stat, err := os.Stat(sitesDir); err == nil && stat.IsDir() {
		if entries, err := os.ReadDir(sitesDir); err == nil {
			output.WriteString(fmt.Sprintf("Sites:        %s/ (%d sites)\n", sitesDir, len(entries)))
		} else {
			output.WriteString(fmt.Sprintf("Sites:        %s/ (error reading)\n", sitesDir))
		}
	} else {
		output.WriteString(fmt.Sprintf("Sites:        %s/ (not found)\n", sitesDir))
	}

	// Check PID file for server status
	pidFile := filepath.Join(configDir, "cc-server.pid")
	if pidData, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		output.WriteString(fmt.Sprintf("\nServer:       ● Running (PID: %s)\n", pidStr))
	} else {
		output.WriteString("\nServer:       ○ Not running\n")
	}

	return output.String(), nil
}

// handleServerCommand handles server-related subcommands
func handleServerCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: server command requires a subcommand")
		printServerHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		handleInitCommand()
	case "set-credentials":
		handleSetCredentials()
	case "set-config":
		handleSetConfigCommand()
	case "status":
		handleStatusCommand()
	case "start":
		handleStartCommand()
	case "stop":
		handleStopCommand()
	case "--help", "-h", "help":
		printServerHelp()
	default:
		fmt.Printf("Unknown server command: %s\n\n", subcommand)
		printServerHelp()
		os.Exit(1)
	}
}

// handleClientCommand handles client-related subcommands
func handleClientCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: client command requires a subcommand")
		printClientHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "set-auth-token":
		handleSetAuthToken()
	case "deploy":
		handleDeployCommand()
	case "--help", "-h", "help":
		printClientHelp()
	default:
		fmt.Printf("Unknown client command: %s\n\n", subcommand)
		printClientHelp()
		os.Exit(1)
	}
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			log.Printf("[%s] %s %s %d %v", requestID, r.Method, r.URL.Path, wrapped.statusCode, duration)
		} else {
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
		}
	})
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()

		if cfg.IsDevelopment() {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics and logs the error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("fazt.sh %s\n", Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// createRootHandler creates a handler that routes based on the Host header
// - Requests to the main domain (or localhost) go to the dashboard
// - Requests to subdomains (*.domain.com or *.localhost) go to the site handler
func createRootHandler(cfg *config.Config, dashboardMux *http.ServeMux, sessionStore *auth.SessionStore) http.Handler {
	// Parse the main domain from config
	mainDomain := extractDomain(cfg.Server.Domain)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		// Remove port from host if present
		if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
			// Check if this is IPv6 (has brackets)
			if !strings.Contains(host, "]") || strings.LastIndex(host, "]") < colonIdx {
				host = host[:colonIdx]
			}
		}

		// Check if this is the main domain or localhost (no subdomain)
		if isDashboardHost(host, mainDomain, cfg.Server.Port) {
			// Apply auth middleware only to dashboard routes
			middleware.AuthMiddleware(sessionStore)(dashboardMux).ServeHTTP(w, r)
			return
		}

		// Extract subdomain and serve the site
		subdomain := extractSubdomain(host, mainDomain)
		if subdomain != "" {
			siteHandler(w, r, subdomain)
			return
		}

		// Fallback to dashboard
		middleware.AuthMiddleware(sessionStore)(dashboardMux).ServeHTTP(w, r)
	})
}

// extractDomain extracts the domain from a URL (removes protocol and path)
func extractDomain(rawURL string) string {
	// Handle URLs with protocol
	if strings.Contains(rawURL, "://") {
		if parsed, err := url.Parse(rawURL); err == nil {
			return parsed.Hostname()
		}
	}
	// Handle bare domains
	if colonIdx := strings.Index(rawURL, ":"); colonIdx != -1 {
		return rawURL[:colonIdx]
	}
	return rawURL
}

// isDashboardHost checks if the host should be routed to the dashboard
func isDashboardHost(host, mainDomain, port string) bool {
	// Exact match with main domain
	if host == mainDomain {
		return true
	}

	// localhost without subdomain
	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	return false
}

// extractSubdomain extracts the subdomain from a host
// e.g., "blog.example.com" with mainDomain "example.com" returns "blog"
// e.g., "blog.localhost" returns "blog"
func extractSubdomain(host, mainDomain string) string {
	host = strings.ToLower(host)
	mainDomain = strings.ToLower(mainDomain)

	// Handle *.localhost pattern
	if strings.HasSuffix(host, ".localhost") {
		return strings.TrimSuffix(host, ".localhost")
	}

	// Handle *.127.0.0.1 pattern (rare but possible)
	if strings.HasSuffix(host, ".127.0.0.1") {
		return strings.TrimSuffix(host, ".127.0.0.1")
	}

	// Handle *.mainDomain pattern
	suffix := "." + mainDomain
	if strings.HasSuffix(host, suffix) {
		subdomain := strings.TrimSuffix(host, suffix)
		// Don't return empty subdomain or subdomain with dots (nested subdomains)
		if subdomain != "" && !strings.Contains(subdomain, ".") {
			return subdomain
		}
	}

	return ""
}

// siteHandler handles requests for hosted sites
// Serves static files from ~/.config/fazt/sites/{subdomain}/
// If main.js exists, executes serverless JavaScript instead
// WebSocket connections at /ws are handled by the WebSocket hub
func siteHandler(w http.ResponseWriter, r *http.Request, subdomain string) {
	// Check if site exists
	if !hosting.SiteExists(subdomain) {
		serveSiteNotFound(w, subdomain)
		return
	}

	// Handle WebSocket connections at /ws
	if r.URL.Path == "/ws" {
		hosting.HandleWebSocket(w, r, subdomain)
		return
	}

	// Log analytics event for site visits
	logSiteVisit(r, subdomain)

	// Get the site directory
	siteDir := hosting.GetSiteDir(subdomain)

	// Check for serverless (main.js)
	if hosting.HasServerless(siteDir) {
		db := database.GetDB()
		if hosting.RunServerless(w, r, siteDir, db, subdomain) {
			return // Serverless handled the request
		}
	}

	// Create a secure file server for this site (prevents path traversal)
	fileServer := hosting.NewSecureFileServer(siteDir)

	// Serve the request
	// Strip nothing since we're serving from root of site directory
	fileServer.ServeHTTP(w, r)
}

// logSiteVisit logs an analytics event for a site visit
func logSiteVisit(r *http.Request, subdomain string) {
	db := database.GetDB()
	if db == nil {
		return
	}

	// Insert event into database
	_, err := db.Exec(`
		INSERT INTO events (domain, source_type, event_type, path, referrer, user_agent, ip_address, query_params)
		VALUES (?, 'hosting', 'pageview', ?, ?, ?, ?, ?)
	`,
		subdomain,
		r.URL.Path,
		r.Referer(),
		r.UserAgent(),
		r.RemoteAddr,
		r.URL.RawQuery,
	)

	if err != nil {
		log.Printf("Failed to log site visit: %v", err)
	}
}

// serveSiteNotFound renders the 404 page for non-existent sites
func serveSiteNotFound(w http.ResponseWriter, subdomain string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Site Not Found</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
               display: flex; justify-content: center; align-items: center;
               height: 100vh; margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 40px; background: white;
                     border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 10px; }
        p { color: #666; }
        .subdomain { font-family: monospace; background: #f0f0f0; padding: 2px 8px; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>404 - Site Not Found</h1>
        <p>The site <span class="subdomain">%s</span> does not exist.</p>
    </div>
</body>
</html>`, subdomain)
}

// createDeployZip creates a ZIP archive of the directory
func createDeployZip(dir string) (*bytes.Buffer, int, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	fileCount := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Create ZIP entry
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		fileCount++
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, 0, err
	}

	return buf, fileCount, nil
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleSetCredentials handles the set-credentials subcommand
func handleSetCredentials() {
	flags := flag.NewFlagSet("set-credentials", flag.ExitOnError)
	username := flags.String("username", "", "Username for authentication")
	password := flags.String("password", "", "Password for authentication")
	configPath := flags.String("config", "", "Config file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server set-credentials [flags]")
		fmt.Println()
		fmt.Println("Update authentication credentials for the fazt.sh dashboard.")
		fmt.Println("At least one of --username or --password must be provided.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server set-credentials --username newuser")
		fmt.Println("  fazt server set-credentials --password newpass")
		fmt.Println("  fazt server set-credentials --username admin --password secret123")
		fmt.Println("  fazt server set-credentials --username admin --config /path/to/config.json")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Get config path
	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		*configPath = filepath.Join(homeDir, ".config", "fazt", "config.json")
	}

	// Call command function
	if err := setCredentialsCommand(*username, *password, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Credentials updated successfully")
	if *username != "" {
		fmt.Printf("  Username: %s\n", *username)
	}
	if *password != "" {
		fmt.Println("  Password: [updated and hashed]")
	}
	fmt.Println()
}

// handleInitCommand handles the init subcommand
func handleInitCommand() {
	flags := flag.NewFlagSet("init", flag.ExitOnError)
	username := flags.String("username", "", "Admin username (required)")
	password := flags.String("password", "", "Admin password (required)")
	domain := flags.String("domain", "", "Server domain (required)")
	port := flags.String("port", "4698", "Server port")
	env := flags.String("env", "development", "Environment (development|production)")
	configPath := flags.String("config", "", "Config file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server init [flags]")
		fmt.Println()
		fmt.Println("Initialize fazt.sh server configuration")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server init --username admin --password secret123 --domain https://mydomain.com")
		fmt.Println("  fazt server init --username admin --password secret123 --domain https://mydomain.com --port 8080 --env production")
		fmt.Println("  fazt server init --username admin --password secret123 --domain https://mydomain.com --config /path/to/config.json")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Get config path
	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		*configPath = filepath.Join(homeDir, ".config", "fazt", "config.json")
	}

	// Call command function
	if err := initCommand(*username, *password, *domain, *port, *env, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Server initialized successfully")
	fmt.Printf("  Config saved to: %s\n", *configPath)
	fmt.Println()
	fmt.Println("To start the server:")
	fmt.Println("  fazt server start")
	fmt.Println()
}

// handleSetConfigCommand handles the set-config subcommand
func handleSetConfigCommand() {
	flags := flag.NewFlagSet("set-config", flag.ExitOnError)
	domain := flags.String("domain", "", "Server domain")
	port := flags.String("port", "", "Server port")
	env := flags.String("env", "", "Environment (development|production)")
	configPath := flags.String("config", "", "Config file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server set-config [flags]")
		fmt.Println()
		fmt.Println("Update server configuration settings")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server set-config --domain https://newdomain.com")
		fmt.Println("  fazt server set-config --port 8080")
		fmt.Println("  fazt server set-config --env production")
		fmt.Println("  fazt server set-config --domain https://prod.com --port 443 --env production")
		fmt.Println("  fazt server set-config --domain https://prod.com --config /path/to/config.json")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Get config path
	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		*configPath = filepath.Join(homeDir, ".config", "fazt", "config.json")
	}

	// Call command function
	if err := setConfigCommand(*domain, *port, *env, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Configuration updated successfully")
	if *domain != "" {
		fmt.Printf("  Domain: %s\n", *domain)
	}
	if *port != "" {
		fmt.Printf("  Port: %s\n", *port)
	}
	if *env != "" {
		fmt.Printf("  Environment: %s\n", *env)
	}
	fmt.Println()
}

// handleStatusCommand handles the status subcommand
func handleStatusCommand() {
	flags := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := flags.String("config", "", "Config file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server status [flags]")
		fmt.Println()
		fmt.Println("Display server configuration and status")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Shows:")
		fmt.Println("  Configuration file location")
		fmt.Println("  Server settings (domain, port, environment)")
		fmt.Println("  Authentication status")
		fmt.Println("  Database information")
		fmt.Println("  Site deployment directory")
		fmt.Println("  Server running status")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server status")
		fmt.Println("  fazt server status --config /path/to/config.json")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Get config path and directory
	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		*configPath = filepath.Join(homeDir, ".config", "fazt", "config.json")
	}
	configDir := filepath.Dir(*configPath)

	// Call command function
	output, err := statusCommand(*configPath, configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

// handleSetAuthToken handles the set-auth-token subcommand
func handleSetAuthToken() {
	flags := flag.NewFlagSet("set-auth-token", flag.ExitOnError)
	token := flags.String("token", "", "Authentication token (required)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client set-auth-token --token <TOKEN>")
		fmt.Println()
		fmt.Println("Sets the authentication token for site deployments.")
		fmt.Println("Generate a token in the web interface at /hosting,")
		fmt.Println("then configure it with this command.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cc-server set-auth-token --token abc123def456789")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *token == "" {
		fmt.Println("Error: --token is required")
		flags.Usage()
		os.Exit(1)
	}

	// Load or create config
	flagsConfig := config.ParseFlags()
	configPath := config.ExpandPath(flagsConfig.ConfigPath)
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating new config at %s\n", configPath)
			cfg = config.CreateDefaultConfig()
		} else {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	// Validate token format (basic validation)
	if len(*token) < 10 {
		fmt.Println("Warning: Token seems too short (minimum 10 characters recommended)")
	}

	// Set token in config (simplified - no name needed)
	cfg.SetAPIKey(*token, "deployment-token")

	// Save config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	// Success message
	fmt.Println()
	fmt.Println("✓ Authentication token configured successfully!")
	fmt.Println()
	fmt.Printf("Token:  %s...%s (truncated)\n", (*token)[:4], (*token)[len(*token)-4:])
	fmt.Println("Config: ~/.config/fazt/config.json")
	fmt.Println()
	fmt.Println("You can now deploy sites:")
	fmt.Println("  fazt client deploy --path . --domain my-site")
	fmt.Println()
}
func handleDeployCommand() {
	flags := flag.NewFlagSet("deploy", flag.ExitOnError)
	path := flags.String("path", "", "Directory to deploy (required)")
	domain := flags.String("domain", "", "Domain/subdomain for the site (required)")
	server := flags.String("server", "http://localhost:4698", "fazt.sh server URL")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client deploy --path <PATH> --domain <SUBDOMAIN>")
		fmt.Println()
		fmt.Println("Deploys a directory to a fazt.sh server.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cc-server deploy --path . --domain my-site")
		fmt.Println("  cc-server deploy --path ~/Desktop/site --domain example --server https://cc.example.com")
		fmt.Println("  cc-server deploy --domain my-site --path .")
	}

	// Determine args offset based on whether this is "deploy" or "client deploy"
	argsOffset := 3
	if len(os.Args) > 1 && os.Args[1] == "deploy" {
		argsOffset = 2
	}

	if err := flags.Parse(os.Args[argsOffset:]); err != nil {
		os.Exit(1)
	}

	deployPath := *path
	if deployPath == "" {
		fmt.Println("Error: --path is required")
		flags.Usage()
		os.Exit(1)
	}

	if *domain == "" {
		fmt.Println("Error: --domain is required")
		flags.Usage()
		os.Exit(1)
	}

	// Validate the path exists
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		fmt.Printf("Error: Path '%s' does not exist\n", deployPath)
		os.Exit(1)
	}

	// Load config to get API key
	flagsConfig := config.ParseFlags()
	cfg, err := config.Load(flagsConfig)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get API key from config
	token := cfg.GetAPIKey()
	if token == "" {
		fmt.Println("Error: No API key found in config")
		fmt.Println("Please ensure you have an API key configured in ~/.config/fazt/config.json")
		os.Exit(1)
	}

	fmt.Printf("Deploying %s to %s as '%s'...\n", deployPath, *server, *domain)

	// Change to the deploy directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(deployPath); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", deployPath, err)
		os.Exit(1)
	}
	defer os.Chdir(originalDir)

	// Create ZIP of the directory
	zipBuffer, fileCount, err := createDeployZip(".")
	if err != nil {
		fmt.Printf("Error creating ZIP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Zipped %d files (%d bytes)\n", fileCount, zipBuffer.Len())

	// Create HTTP request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add domain field
	if err := writer.WriteField("site_name", *domain); err != nil {
		fmt.Printf("Error creating form: %v\n", err)
		os.Exit(1)
	}

	// Add file field
	part, err := writer.CreateFormFile("file", "deploy.zip")
	if err != nil {
		fmt.Printf("Error creating file field: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.Copy(part, zipBuffer); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}
	writer.Close()

	// Make request
	req, err := http.NewRequest("POST", *server+"/api/deploy", &body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error deploying: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("✗ Deployment failed!\n")
		fmt.Printf("  Status: %s\n", resp.Status)
		fmt.Printf("  Error: %s\n", string(respBody))
		os.Exit(1)
	}

	// Parse success response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if success, ok := result["success"].(bool); ok && success {
			fmt.Printf("✓ Deployment successful!\n")
			if site, ok := result["site"].(string); ok {
				// Extract server URL for display
				serverURL := *server
				serverURL = strings.TrimPrefix(serverURL, "http://")
				serverURL = strings.TrimPrefix(serverURL, "https://")
				fmt.Printf("  Site: http://%s.%s\n", site, serverURL)
			}
			if fileCount, ok := result["file_count"].(float64); ok {
				fmt.Printf("  Files: %.0f\n", fileCount)
			}
			if sizeBytes, ok := result["size_bytes"].(float64); ok {
				fmt.Printf("  Size: %.0f bytes\n", sizeBytes)
			}
			return
		}
	}

	fmt.Printf("✓ Deployment completed! (Status: %s)\n", resp.Status)
}

// handleStartCommand handles the start subcommand
func handleStartCommand() {
	flags := flag.NewFlagSet("start", flag.ExitOnError)
	port := flags.String("port", "", "Server port (overrides config)")
	db := flags.String("db", "", "Database file path (overrides config)")
	configFile := flags.String("config", "", "Config file path")
	domain := flags.String("domain", "", "Server domain (overrides config)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server start [options]")
		fmt.Println()
		fmt.Println("Starts the fazt.sh server.")
		fmt.Println()
		fmt.Println("Domain Configuration:")
		fmt.Println("  Default: https://fazt.sh (for project use)")
		fmt.Println("  Override: --domain yourdomain.com")
		fmt.Println("  Environment: FAZT_DOMAIN environment variable")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cc-server server start")
		fmt.Println("  cc-server server start --port 8080")
		fmt.Println("  cc-server server start --domain mysite.com")
		fmt.Println("  cc-server server start --config /path/to/config.json")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  FAZT_DOMAIN=fazt.sh cc-server server start")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Set up configuration
	if !*quiet {
		log.Println("Starting fazt.sh...")
	}

	// Use default flags structure but override with our specific flags
	cliFlags := config.ParseFlags()
	if *port != "" {
		cliFlags.Port = *port
	}
	if *db != "" {
		cliFlags.DBPath = *db
	}
	if *configFile != "" {
		cliFlags.ConfigPath = *configFile
	}
	// Load configuration
	cfg, err := config.Load(cliFlags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply domain override if provided (highest priority)
	if *domain != "" {
		cfg.Server.Domain = *domain
	}

	// Ensure secure file permissions
	security.EnsureSecurePermissions(config.ExpandPath(cliFlags.ConfigPath), cfg.Database.Path)

	// Display startup information
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("             fazt.sh v0.3.0 - Starting Up")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Environment:  %s\n", cfg.Server.Env)
	fmt.Printf("  Port:         %s\n", cfg.Server.Port)
	fmt.Printf("  Domain:       %s\n", cfg.Server.Domain)
	fmt.Printf("  Database:     %s\n", cfg.Database.Path)
	fmt.Printf("  Config File:  %s\n", config.ExpandPath(cliFlags.ConfigPath))
	fmt.Println()

	// Initialize session store
	sessionStore := auth.NewSessionStore(auth.SessionTTL)
	defer sessionStore.Stop()

	// Initialize rate limiter
	rateLimiter := auth.NewRateLimiter()

	// Initialize auth handlers with session store and rate limiter
	handlers.InitAuth(sessionStore, rateLimiter)

	// Display auth status (v0.4.0: auth always required)
	fmt.Printf("  Authentication: ✓ Enabled (user: %s)\n", cfg.Auth.Username)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize database
	if err := database.Init(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize audit logging
	if err := audit.Init(database.GetDB()); err != nil {
		log.Fatalf("Failed to initialize audit logging: %v", err)
	}

	// Initialize hosting system
	configDir := filepath.Dir(cfg.Database.Path)
	if err := hosting.Init(configDir); err != nil {
		log.Fatalf("Failed to initialize hosting: %v", err)
	}
	log.Printf("Hosting initialized: %s", hosting.GetSitesDir())

	// Generate mock data in development mode
	if cfg.IsDevelopment() {
		log.Println("Development mode: Checking for existing data...")
		// Only generate mock data if database is empty
		db := database.GetDB()
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
		if err == nil && count == 0 {
			log.Println("Database is empty, generating mock data...")
			if err := database.GenerateMockData(); err != nil {
				log.Printf("Warning: Failed to generate mock data: %v", err)
			}
		} else {
			log.Printf("Database already has %d events, skipping mock data generation", count)
		}
	}

	// Create dashboard router (existing dashboard functionality)
	dashboardMux := http.NewServeMux()

	// Authentication routes
	dashboardMux.HandleFunc("/login", handlers.LoginPageHandler)
	dashboardMux.HandleFunc("/api/login", handlers.LoginHandler)
	dashboardMux.HandleFunc("/api/logout", handlers.LogoutHandler)
	dashboardMux.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)

	// API routes - Tracking
	dashboardMux.HandleFunc("/track", handlers.TrackHandler)
	dashboardMux.HandleFunc("/pixel.gif", handlers.PixelHandler)
	dashboardMux.HandleFunc("/r/", handlers.RedirectHandler)
	dashboardMux.HandleFunc("/webhook/", handlers.WebhookHandler)

	// API routes - Dashboard
	dashboardMux.HandleFunc("/api/stats", handlers.StatsHandler)
	dashboardMux.HandleFunc("/api/events", handlers.EventsHandler)
	dashboardMux.HandleFunc("/api/redirects", handlers.RedirectsHandler)
	dashboardMux.HandleFunc("/api/domains", handlers.DomainsHandler)
	dashboardMux.HandleFunc("/api/tags", handlers.TagsHandler)
	dashboardMux.HandleFunc("/api/webhooks", handlers.WebhooksHandler)
	dashboardMux.HandleFunc("/api/config", handlers.ConfigHandler)

	// API routes - Hosting/Deploy
	dashboardMux.HandleFunc("/api/deploy", handlers.DeployHandler)
	dashboardMux.HandleFunc("/api/sites", handlers.SitesHandler)
	dashboardMux.HandleFunc("/api/keys", handlers.APIKeysHandler)
	dashboardMux.HandleFunc("/api/deployments", handlers.DeploymentsHandler)
	dashboardMux.HandleFunc("/api/envvars", handlers.EnvVarsHandler)

	// Hosting management page
	dashboardMux.HandleFunc("/hosting", handlers.HostingPageHandler)

	// Static files
	fs := http.FileServer(http.Dir("./web/static"))
	dashboardMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Dashboard (root)
	dashboardMux.HandleFunc("/", handlers.DashboardHandler)

	// Health check (available on both dashboard and sites)
	dashboardMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(); err != nil {
			http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create the root handler with host-based routing
	rootHandler := createRootHandler(cfg, dashboardMux, sessionStore)

	// Apply middleware (order: tracing -> logging -> body limit -> security -> cors -> recovery -> root)
	handler := middleware.RequestTracing(
		loggingMiddleware(
			middleware.BodySizeLimit(middleware.MaxBodySize)(
				middleware.SecurityHeaders(
					corsMiddleware(
						recoveryMiddleware(rootHandler),
					),
				),
			),
		),
	)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Write PID file for stop command
	pidFile := filepath.Join(filepath.Dir(cfg.Database.Path), "cc-server.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		log.Printf("Warning: Failed to write PID file: %v", err)
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%s", cfg.Server.Port)
		log.Printf("Dashboard: http://localhost:%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Clean up PID file
	os.Remove(pidFile)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// handleStopCommand handles the stop subcommand
func handleStopCommand() {
	flags := flag.NewFlagSet("stop", flag.ExitOnError)

	flags.Usage = func() {
		fmt.Println("Usage: fazt server stop")
		fmt.Println()
		fmt.Println("Stops a running fazt.sh server.")
		fmt.Println()
		fmt.Println("Looks for a PID file in ~/.config/fazt/ to gracefully shutdown the server.")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Get default config directory to find PID file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	pidFile := filepath.Join(homeDir, ".config", "cc", "cc-server.pid")

	// Read PID file
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No running server found (no PID file)")
			os.Exit(1)
		}
		log.Fatalf("Failed to read PID file: %v", err)
	}

	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		log.Fatalf("Invalid PID file format: %v", err)
	}

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("Failed to find process: %v", err)
	}

	// Send SIGTERM for graceful shutdown
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Printf("Warning: Could not send signal to process %d: %v\n", pid, err)
		fmt.Println("The process may have already stopped")
	} else {
		fmt.Printf("Sent shutdown signal to process %d\n", pid)
		fmt.Println("Waiting for graceful shutdown...")

		// Wait a bit and check if process is still running
		time.Sleep(2 * time.Second)
		err = process.Signal(syscall.Signal(0))
		if err == nil {
			fmt.Println("Process is still running, sending forceful shutdown...")
			process.Kill()
		}
	}

	// Clean up PID file
	os.Remove(pidFile)
	fmt.Println("Server stopped successfully")
}

// printUsage displays the usage information
func printUsage() {
	fmt.Println("fazt.sh v0.3.0 - Analytics & Personal Cloud Platform")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt <command> [options]")
	fmt.Println()
	fmt.Println("MAIN COMMANDS:")
	fmt.Println("  server           Server management commands")
	fmt.Println("  client           Client/deployment commands")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println("  --version        Show version and exit")
	fmt.Println()
	fmt.Println("For detailed help:")
	fmt.Println("  fazt server --help     # Server commands")
	fmt.Println("  fazt client --help     # Client commands")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/jikkuatwork/fazt.sh")
}

// printServerHelp displays server-specific help
func printServerHelp() {
	fmt.Println("fazt.sh v0.3.0 - Server Commands")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt server <command> [options]")
	fmt.Println()
	fmt.Println("SERVER COMMANDS:")
	fmt.Println("  set-credentials  Set up authentication credentials")
	fmt.Println("  start            Start the fazt.sh server")
	fmt.Println("  stop             Stop a running fazt.sh server")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Set up authentication")
	fmt.Println("  fazt server set-credentials --username admin --password secret123")
	fmt.Println()
	fmt.Println("  # Start the server")
	fmt.Println("  fazt server start")
	fmt.Println()
	fmt.Println("  # Start on custom port")
	fmt.Println("  fazt server start --port 8080")
	fmt.Println()
	fmt.Println("  # Stop the server")
	fmt.Println("  fazt server stop")
	fmt.Println()
}

// printClientHelp displays client-specific help
func printClientHelp() {
	fmt.Println("fazt.sh v0.3.0 - Client Commands")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt client <command> [options]")
	fmt.Println()
	fmt.Println("CLIENT COMMANDS:")
	fmt.Println("  set-auth-token   Set authentication token for deployments")
	fmt.Println("  deploy           Deploy a directory to a site")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Set authentication token")
	fmt.Println("  fazt client set-auth-token --token abc123def456")
	fmt.Println()
	fmt.Println("  # Deploy current directory")
	fmt.Println("  fazt client deploy --path . --domain my-site")
	fmt.Println()
	fmt.Println("  # Deploy to remote server")
	fmt.Println("  fazt client deploy --path ./build --domain app --server https://fazt.sh")
	fmt.Println()
	fmt.Println("WORKFLOW:")
	fmt.Println("  1. Start server: fazt server start")
	fmt.Println("  2. Visit /hosting in your browser to generate token")
	fmt.Println("  3. Set token: fazt client set-auth-token --token <TOKEN>")
	fmt.Println("  4. Deploy sites: fazt client deploy --path . --domain my-site")
	fmt.Println()
}
