package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jikku/command-center/internal/audit"
	"github.com/jikku/command-center/internal/auth"
	"github.com/jikku/command-center/internal/config"
	"github.com/jikku/command-center/internal/database"
	"github.com/jikku/command-center/internal/handlers"
	"github.com/jikku/command-center/internal/middleware"
	"github.com/jikku/command-center/internal/security"
)

const Version = "v0.2.0"

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	showHelp    = flag.Bool("help", false, "Show help and exit")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	quiet       = flag.Bool("quiet", false, "Quiet mode (errors only)")
)

func main() {
	// Check for --version or --help before parsing other flags
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			printVersion()
			return
		}
		if arg == "--help" || arg == "-help" || arg == "-h" {
			printHelp()
			return
		}
	}

	// Configure logging based on flags
	if *quiet {
		log.SetOutput(io.Discard)
	}

	if !*quiet {
		log.Println("Starting Command Center...")
	}

	// Parse CLI flags
	flags := config.ParseFlags()

	// Handle credential setup mode (--username and --password flags)
	if flags.Username != "" && flags.Password != "" {
		handleCredentialSetup(flags)
		return
	}

	// Load configuration
	cfg, err := config.Load(flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Ensure secure file permissions
	security.EnsureSecurePermissions(config.ExpandPath(flags.ConfigPath), cfg.Database.Path)

	// Display startup information
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("           Command Center v0.2.0 - Starting Up")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Environment:  %s\n", cfg.Server.Env)
	fmt.Printf("  Port:         %s\n", cfg.Server.Port)
	fmt.Printf("  Domain:       %s\n", cfg.Server.Domain)
	fmt.Printf("  Database:     %s\n", cfg.Database.Path)
	fmt.Printf("  Config File:  %s\n", config.ExpandPath(flags.ConfigPath))
	fmt.Println()

	// Initialize session store
	sessionStore := auth.NewSessionStore(auth.SessionTTL)
	defer sessionStore.Stop()

	// Initialize rate limiter
	rateLimiter := auth.NewRateLimiter()

	// Initialize auth handlers with session store and rate limiter
	handlers.InitAuth(sessionStore, rateLimiter)

	// Display auth status with warnings
	if cfg.Auth.Enabled {
		fmt.Printf("  Authentication: ✓ Enabled (user: %s)\n", cfg.Auth.Username)
	} else {
		fmt.Println("  Authentication: ✗ Disabled")
		if cfg.IsProduction() {
			fmt.Println()
			fmt.Println("  ⚠️  WARNING: Running in PRODUCTION without authentication!")
			fmt.Println("  ⚠️  Your dashboard is publicly accessible.")
			fmt.Println()
			fmt.Println("  To enable authentication, run:")
			fmt.Printf("    ./cc-server --username admin --password yourpassword\n")
			fmt.Println()
		}
	}
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

	// Create router
	mux := http.NewServeMux()

	// Apply middleware (order: logging -> security -> auth -> cors -> recovery -> mux)
	handler := loggingMiddleware(
		middleware.SecurityHeaders(
			middleware.AuthMiddleware(sessionStore)(
				corsMiddleware(
					recoveryMiddleware(mux),
				),
			),
		),
	)

	// Authentication routes
	mux.HandleFunc("/login", handlers.LoginPageHandler)
	mux.HandleFunc("/api/login", handlers.LoginHandler)
	mux.HandleFunc("/api/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)

	// API routes - Tracking
	mux.HandleFunc("/track", handlers.TrackHandler)
	mux.HandleFunc("/pixel.gif", handlers.PixelHandler)
	mux.HandleFunc("/r/", handlers.RedirectHandler)
	mux.HandleFunc("/webhook/", handlers.WebhookHandler)

	// API routes - Dashboard
	mux.HandleFunc("/api/stats", handlers.StatsHandler)
	mux.HandleFunc("/api/events", handlers.EventsHandler)
	mux.HandleFunc("/api/redirects", handlers.RedirectsHandler)
	mux.HandleFunc("/api/domains", handlers.DomainsHandler)
	mux.HandleFunc("/api/tags", handlers.TagsHandler)
	mux.HandleFunc("/api/webhooks", handlers.WebhooksHandler)
	mux.HandleFunc("/api/config", handlers.ConfigHandler)

	// Static files
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Dashboard (root)
	mux.HandleFunc("/", handlers.DashboardHandler)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(); err != nil {
			http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
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

// handleCredentialSetup sets up or updates credentials when --username and --password flags are provided
func handleCredentialSetup(flags *config.CLIFlags) {
	log.Println("Setting up authentication credentials...")

	// Validate password strength
	isStrong, warnings := auth.ValidatePasswordStrength(flags.Password)
	if len(warnings) > 0 {
		log.Println("Password strength warnings:")
		for _, warning := range warnings {
			log.Printf("  - %s", warning)
		}
	}
	if isStrong {
		log.Println("Password strength: Strong")
	} else {
		log.Println("Password strength: Weak (but acceptable)")
	}

	// Hash the password
	passwordHash, err := auth.HashPassword(flags.Password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Load or create config
	configPath := config.ExpandPath(flags.ConfigPath)
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file not found, creating new config at %s", configPath)
			cfg = config.CreateDefaultConfig()
		} else {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	// Update auth configuration
	cfg.Auth.Enabled = true
	cfg.Auth.Username = flags.Username
	cfg.Auth.PasswordHash = passwordHash

	// Save config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	// Success message
	fmt.Println()
	fmt.Println("✓ Authentication configured successfully!")
	fmt.Println()
	fmt.Printf("Config file: %s\n", configPath)
	fmt.Printf("Username:    %s\n", flags.Username)
	fmt.Println("Password:    [hashed and saved]")
	fmt.Println("Auth:        enabled")
	fmt.Println()
	fmt.Println("To start the server:")
	fmt.Println("  ./cc-server")
	fmt.Println()
	fmt.Println("Or with custom config:")
	fmt.Printf("  ./cc-server --config %s\n", configPath)
	fmt.Println()
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("Command Center %s\n", Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// printHelp displays usage information
func printHelp() {
	fmt.Println("Command Center v0.2.0 - Universal Tracking & Analytics Server")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cc-server [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  --config <path>       Path to config file (default: ~/.config/cc/config.json)")
	fmt.Println("  --env <environment>   Load environment-specific config (development/production)")
	fmt.Println("  --db <path>           Database file path (overrides config)")
	fmt.Println("  --port <port>         Server port (overrides config)")
	fmt.Println("  --username <user>     Set/update username (creates/updates config)")
	fmt.Println("  --password <pass>     Set/update password (creates/updates config)")
	fmt.Println("  --version             Show version and exit")
	fmt.Println("  --help, -h            Show this help")
	fmt.Println("  --verbose             Enable verbose logging")
	fmt.Println("  --quiet               Quiet mode (errors only)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Setup authentication")
	fmt.Println("  cc-server --username admin --password mysecurepass")
	fmt.Println()
	fmt.Println("  # Start server with default config")
	fmt.Println("  cc-server")
	fmt.Println()
	fmt.Println("  # Start with specific environment")
	fmt.Println("  cc-server --env production")
	fmt.Println()
	fmt.Println("  # Start with custom config and database")
	fmt.Println("  cc-server --config /path/to/config.json --db /path/to/data.db")
	fmt.Println()
	fmt.Println("  # Start on custom port")
	fmt.Println("  cc-server --port 8080")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/jikkuatwork/command-center")
}
