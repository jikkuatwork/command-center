# Command Center - Auth & Config Upgrade Plan

## Project Overview
- **Current Version**: v0.1.0
- **Target Version**: v0.2.0
- **Goal**: Production-ready deployment with authentication and flexible configuration

---

## Phase 0: Configuration System Refactor (Commit #0)

**Duration**: 20-25 minutes

### Tasks:
- [x] Create config schema structure in `internal/config/config.go`:
  ```go
  type Config struct {
      Server   ServerConfig   `json:"server"`
      Database DatabaseConfig `json:"database"`
      Auth     AuthConfig     `json:"auth"`
      Ntfy     NtfyConfig     `json:"ntfy"`
  }

  type ServerConfig struct {
      Port   string `json:"port"`
      Domain string `json:"domain"`
      Env    string `json:"env"` // development/production
  }

  type DatabaseConfig struct {
      Path string `json:"path"`
  }

  type AuthConfig struct {
      Enabled  bool   `json:"enabled"`
      Username string `json:"username"`
      Password string `json:"password_hash"` // bcrypt hash
  }

  type NtfyConfig struct {
      Topic string `json:"topic"`
      URL   string `json:"url"`
  }
  ```

- [x] Add CLI flag parsing:
  - `--config` flag for custom config file path
  - `--db` flag for database path override
  - `--port` flag for port override
  - `--username` flag to set/reset username (updates config)
  - `--password` flag to set/reset password (hashes and updates config)
  - Default config path: `~/.config/cc/config.json`
  - Default database path: `~/.config/cc/data.db`

- [x] Implement config loading priority:
  1. CLI flags (highest priority)
  2. JSON config file
  3. Environment variables (backward compatibility)
  4. Built-in defaults (lowest priority)

- [x] Add config file creation and update logic:
  - Check if config directory exists (`~/.config/cc/`)
  - Create directory if missing
  - Generate default config if not found
  - If `--username` and `--password` flags provided:
    - Hash password with bcrypt
    - Update/create config file with new credentials
    - Enable auth in config
    - Exit after updating (or continue to start server)

- [x] Create example config file `config.example.json`:
  ```json
  {
    "server": {
      "port": "4698",
      "domain": "https://cc.toolbomber.com",
      "env": "production"
    },
    "database": {
      "path": "~/.config/cc/data.db"
    },
    "auth": {
      "enabled": true,
      "username": "admin",
      "password_hash": "$2a$10$..."
    },
    "ntfy": {
      "topic": "your-topic",
      "url": "https://ntfy.sh"
    }
  }
  ```

- [x] Update `cmd/server/main.go`:
  - Parse flags before loading config
  - Pass flags to config loader
  - Expand `~` in file paths to home directory

- [x] Add validation for config values:
  - Valid port number (1-65535)
  - Valid URL format for domain
  - Database path is writable
  - Password hash format validation

**Commit**: `feat: JSON configuration system with CLI flags`

---

## Phase 1: Password Management & Auth Package (Commit #1)

**Duration**: 15 minutes

### Tasks:
- [x] Create `internal/auth/password.go`:
  - Hash password function using bcrypt (cost 12)
  - Verify password function
  - Generate secure random salt
  - Error handling for invalid inputs

- [x] Integrate password hashing into main.go:
  - When `--username` and `--password` flags are provided:
    - Load or create config file
    - Hash the password
    - Update config with username and password_hash
    - Set `auth.enabled = true`
    - Write config back to file
    - Log success message with credentials set
    - Exit gracefully (don't start server in this mode)

- [x] Add password strength validation:
  - Minimum 8 characters
  - Warning for weak passwords (log to console)
  - Recommendations for strong passwords

- [x] Update `.gitignore`:
  - Add `config.json` (don't commit passwords!)
  - Add `.env`
  - Keep `config.example.json` in repo

- [x] Add helpful messages:
  ```
  Example usage:
  ./cc-server --username admin --password mysecurepass123
  > Config updated at ~/.config/cc/config.json
  > Username: admin
  > Password hash saved
  > Authentication enabled

  To start server:
  ./cc-server
  ```

**Commit**: `feat: password management integrated into main binary`

---

## Phase 2: Session Management (Commit #2)

**Duration**: 25-30 minutes

### Tasks:
- [ ] Create `internal/auth/session.go`:
  - Session store (in-memory with expiry)
  - Generate secure session IDs (crypto/rand)
  - Session validation
  - Session cleanup (expire after 24 hours of inactivity)
  - Concurrent access safety (mutex)

- [ ] Implement session structure:
  ```go
  type Session struct {
      ID        string
      Username  string
      CreatedAt time.Time
      ExpiresAt time.Time
      LastSeen  time.Time
  }

  type SessionStore struct {
      sessions map[string]*Session
      mu       sync.RWMutex
      ttl      time.Duration
  }
  ```

- [ ] Add session methods:
  - `CreateSession(username string) (sessionID string, error)`
  - `ValidateSession(sessionID string) (bool, error)`
  - `DeleteSession(sessionID string)`
  - `CleanupExpired()` - background goroutine
  - `RefreshSession(sessionID string)` - update LastSeen

- [ ] Cookie handling:
  - HTTPOnly cookies for security
  - Secure flag in production
  - SameSite=Strict
  - Path=/
  - MaxAge=86400 (24 hours)

- [ ] Add session persistence option (future):
  - Store sessions in SQLite table
  - Survive server restarts
  - Configurable via config file

**Commit**: `feat: session management with secure cookies`

---

## Phase 3: Authentication Middleware (Commit #3)

**Duration**: 20 minutes

### Tasks:
- [ ] Create `internal/middleware/auth.go`:
  - Authentication middleware
  - Check session cookie
  - Validate session with store
  - Redirect to login if unauthorized
  - Skip auth for public endpoints

- [ ] Implement middleware:
  ```go
  func AuthMiddleware(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          // Check if path requires auth
          if !requiresAuth(r.URL.Path) {
              next.ServeHTTP(w, r)
              return
          }

          // Validate session
          cookie, err := r.Cookie("session_id")
          if err != nil || !auth.ValidateSession(cookie.Value) {
              http.Redirect(w, r, "/login", http.StatusSeeOther)
              return
          }

          next.ServeHTTP(w, r)
      })
  }
  ```

- [ ] Define public endpoints (no auth required):
  - `/track` - tracking endpoint
  - `/pixel.gif` - tracking pixel
  - `/r/*` - redirects
  - `/webhook/*` - webhooks
  - `/static/*` - static files
  - `/login` - login page
  - `/api/login` - login API
  - `/health` - health check

- [ ] Protect private endpoints:
  - `/` - dashboard
  - `/api/stats` - analytics
  - `/api/events` - event list
  - `/api/redirects` - redirect management
  - `/api/webhooks` - webhook management
  - `/api/domains` - domain list
  - `/api/tags` - tag list
  - All other `/api/*` routes

- [ ] Add auth bypass for development mode:
  - If `auth.enabled = false` in config, skip all auth checks
  - Log warning when auth is disabled

**Commit**: `feat: authentication middleware for protected routes`

---

## Phase 4: Login Handler & UI (Commit #4)

**Duration**: 30 minutes

### Tasks:
- [ ] Create `internal/handlers/auth.go`:

  **POST /api/login**:
  - Accept JSON: `{username, password}`
  - Validate credentials against config
  - Create session on success
  - Set session cookie
  - Return success/error JSON
  - Rate limiting (max 5 attempts per 15 min per IP)

  **POST /api/logout**:
  - Clear session cookie
  - Delete session from store
  - Return success

  **GET /api/auth/status**:
  - Check if user is authenticated
  - Return session info if valid

- [ ] Create `web/templates/login.html`:
  - Clean, minimal login form
  - Tabler styling for consistency
  - Username field
  - Password field (type=password)
  - "Remember me" checkbox (extend session to 7 days)
  - Login button
  - Error message display
  - Dark mode support
  - Mobile responsive
  - No external dependencies

- [ ] Add login page route in `cmd/server/main.go`:
  ```go
  mux.HandleFunc("/login", handlers.LoginPageHandler)
  mux.HandleFunc("/api/login", handlers.LoginHandler)
  mux.HandleFunc("/api/logout", handlers.LogoutHandler)
  ```

- [ ] Implement rate limiting:
  - Track login attempts by IP
  - Max 5 failed attempts per 15 minutes
  - Return 429 Too Many Requests
  - Exponential backoff
  - Clear attempts after 15 minutes

- [ ] Add CSRF protection (basic):
  - Generate CSRF token on login page load
  - Validate token on login submission
  - Store in session

- [ ] Security headers:
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - Referrer-Policy: no-referrer

**Commit**: `feat: login handler and UI with rate limiting`

---

## Phase 5: Dashboard Auth Integration (Commit #5)

**Duration**: 15 minutes

### Tasks:
- [ ] Update `cmd/server/main.go`:
  - Apply auth middleware to protected routes only
  - Keep tracking endpoints public
  - Add auth check before serving dashboard

- [ ] Update `internal/handlers/api.go`:
  - Add auth context to handlers (username from session)
  - Log authenticated actions
  - Return 401 Unauthorized for invalid sessions

- [ ] Add logout button to dashboard:
  - Top-right corner in header
  - Icon + "Logout" text
  - Confirm before logout (optional)
  - Redirect to login page after logout

- [ ] Add session info to dashboard:
  - Show logged-in username
  - Show session expiry time
  - Auto-refresh on activity

- [ ] Handle session expiry gracefully:
  - Show toast notification "Session expired"
  - Redirect to login page
  - Preserve return URL (redirect back after login)

- [ ] Update settings page:
  - Add "Change Password" section
  - Require current password
  - Validate new password strength
  - Update config file with new hash
  - Note: Requires server restart to reload config

**Commit**: `feat: dashboard authentication integration`

---

## Phase 6: Config Management UI (Commit #6)

**Duration**: 25-30 minutes

### Tasks:
- [ ] Add Settings > Configuration tab with:

  **Server Settings** (read-only display):
  - Current port
  - Domain
  - Environment (dev/prod)
  - Config file path
  - Database path
  - Note: "Edit config.json to change these settings"

  **Authentication Settings**:
  - Enable/disable auth toggle (dangerous, show warning)
  - Change password form (current + new + confirm)
  - Session timeout configuration
  - Active sessions list with revoke option

  **Database Settings** (read-only):
  - Database path
  - Database size
  - Total records
  - Last backup time (future feature)

  **Advanced Settings**:
  - Export current config as JSON (download)
  - Validate config button (checks syntax)
  - Reload config button (requires auth re-check)

- [ ] Create `internal/handlers/config.go`:

  **GET /api/config**:
  - Return current config (sanitized - no password hashes!)
  - Only show safe fields
  - Require authentication

  **POST /api/config/password**:
  - Change password endpoint
  - Validate current password
  - Hash new password
  - Update config file
  - Return success/error

  **POST /api/config/reload**:
  - Reload config from file
  - Re-initialize necessary components
  - Invalidate all sessions (force re-login)
  - Return new config

- [ ] Add config file hot-reload:
  - Watch config file for changes (optional)
  - Graceful reload without restart
  - Preserve active sessions if auth config unchanged

- [ ] Config validation API:
  - Validate JSON syntax
  - Check required fields
  - Validate port range
  - Validate URL formats
  - Return detailed error messages

**Commit**: `feat: configuration management UI and API`

---

## Phase 7: Enhanced Security Features (Commit #7)

**Duration**: 20 minutes

### Tasks:
- [ ] Add security headers middleware:
  - Content-Security-Policy
  - X-XSS-Protection
  - Strict-Transport-Security (HSTS) in production
  - Permissions-Policy

- [ ] Implement request sanitization:
  - HTML escaping for all user inputs
  - SQL injection prevention (already using prepared statements)
  - Path traversal prevention
  - Max request size limits

- [ ] Add audit logging:
  - Log all authentication events (success/failure)
  - Log config changes
  - Log admin actions (create/delete redirects, webhooks)
  - Store in SQLite table or log file
  - Include: timestamp, IP, username, action, result

- [ ] Create audit log table:
  ```sql
  CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    username TEXT,
    ip_address TEXT,
    action TEXT NOT NULL,
    resource TEXT,
    result TEXT,
    details TEXT
  );
  ```

- [ ] Add audit log viewer in dashboard:
  - Settings > Audit Logs tab
  - Filterable table (date range, action type, user)
  - Export to CSV
  - Auto-cleanup old logs (>90 days)

- [ ] Security recommendations display:
  - Check if default password is being used
  - Check if auth is disabled in production
  - Check if running on default port
  - Display warnings in dashboard

**Commit**: `feat: enhanced security and audit logging`

---

## Phase 8: Database Migration System (Commit #8)

**Duration**: 20 minutes

### Tasks:
- [ ] Update database initialization in `internal/database/db.go`:
  - Check database path directory exists
  - Create `~/.config/cc/` if missing
  - Better error messages for permission issues

- [ ] Add migration tracking:
  ```sql
  CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version INTEGER UNIQUE NOT NULL,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```

- [ ] Create migration files:
  - `migrations/002_audit_logs.sql` - audit logging table
  - `migrations/003_sessions.sql` - session persistence (optional)

- [ ] Implement migration runner:
  - Check which migrations are applied
  - Run pending migrations in order
  - Rollback on error
  - Log migration status

- [ ] Add database backup utility:
  - Create backup before migrations
  - Store in `~/.config/cc/backups/`
  - Keep last 5 backups
  - Restore command (via CLI flag)

- [ ] Add CLI commands:
  - `--migrate` - run pending migrations
  - `--backup` - create manual backup
  - `--restore <file>` - restore from backup

**Commit**: `feat: database migration system with backups`

---

## Phase 9: Config Validation & Startup (Commit #9)

**Duration**: 15 minutes

### Tasks:
- [ ] Add config validation on startup:
  - Check all required fields present
  - Warn about missing optional fields (ntfy topic, etc)
  - Fail fast with helpful error messages
  - If no config exists, create default with auth disabled

- [ ] Add helpful startup messages:
  - If auth is disabled, show warning in production mode
  - If config was auto-generated, show message:
    ```
    No config found, created default at ~/.config/cc/config.json
    To enable authentication, run:
      ./cc-server --username admin --password yourpassword
    ```

- [ ] Add Makefile targets:
  ```makefile
  run:
  	go run cmd/server/main.go

  run-with-config:
  	go run cmd/server/main.go --config ~/.config/cc/config.json

  build:
  	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cc-server ./cmd/server
  ```

- [ ] Update documentation:
  - Add quickstart guide
  - Explain setup process (simple flags)
  - Common troubleshooting

**Commit**: `feat: config validation and helpful startup messages`

---

## Phase 10: Environment-Specific Configs (Commit #10)

**Duration**: 15 minutes

### Tasks:
- [ ] Support multiple config profiles:
  - `config.development.json`
  - `config.production.json`
  - `config.staging.json`
  - Load based on `ENV` flag or auto-detect

- [ ] Add `--env` flag:
  - `--env development` loads `config.development.json`
  - `--env production` loads `config.production.json`
  - Default: auto-detect or use `development`

- [ ] Create default configs for each environment:
  - Development: auth disabled, verbose logging, mock data
  - Production: auth required, minimal logging, no mock data
  - Staging: mix of both

- [ ] Add config merging:
  - Base config + environment-specific overrides
  - Deep merge of JSON objects
  - Precedence: env-specific > base > defaults

- [ ] Update documentation:
  - Explain environment configs
  - Best practices for each environment
  - Security considerations

**Commit**: `feat: environment-specific configuration profiles`

---

## Phase 11: Improved CLI Experience (Commit #11)

**Duration**: 20 minutes

### Tasks:
- [ ] Add `--version` flag:
  - Show version number
  - Show build info (Go version, build date, commit hash)
  - Exit after displaying

- [ ] Add `--help` flag:
  - Comprehensive help text
  - List all available flags
  - Usage examples
  - Config file location info

- [ ] Improve startup messages:
  - Show banner/logo (ASCII art)
  - Display loaded config file path
  - Show auth status (enabled/disabled)
  - Show database path
  - Show server URL
  - Color-coded messages (green=success, yellow=warning, red=error)

- [ ] Add verbose mode `--verbose`:
  - Detailed logging
  - Show config values on startup
  - Debug information
  - Request logging

- [ ] Add quiet mode `--quiet`:
  - Minimal output
  - Only errors
  - Useful for systemd services

- [ ] Signal handling improvements:
  - Graceful shutdown on SIGTERM
  - Config reload on SIGHUP
  - Status dump on SIGUSR1

- [ ] Add `--check` flag:
  - Validate config and exit
  - Check database accessibility
  - Test all dependencies
  - Don't start server
  - Useful for CI/CD

**Commit**: `feat: improved CLI experience with better flags and output`

---

## Phase 12: Security Hardening (Commit #12)

**Duration**: 20 minutes

### Tasks:
- [ ] Add brute-force protection:
  - Track failed login attempts globally
  - Implement exponential backoff
  - Temporary IP bans after 10 failures
  - Unban after 1 hour
  - Admin notification on potential attacks

- [ ] Implement session security:
  - Rotate session IDs after login
  - Bind sessions to IP addresses (optional, configurable)
  - Bind sessions to User-Agent (optional)
  - Detect session hijacking attempts

- [ ] Add password policies:
  - Minimum length requirement
  - Complexity requirements (optional)
  - Password history (prevent reuse)
  - Password expiry (optional, configurable)
  - Force password change on first login (if default detected)

- [ ] Secure file permissions:
  - Set config.json to 0600 (owner read/write only)
  - Warn if permissions too open
  - Set database to 0600
  - Create directories with 0700

- [ ] Add integrity checks:
  - Checksum config file
  - Detect unauthorized modifications
  - Verify database integrity on startup
  - Hash tracking for static files (optional)

- [ ] Add security endpoints:
  - GET /api/security/status - security posture check
  - GET /api/security/recommendations - actionable advice
  - Display in dashboard Settings > Security tab

**Commit**: `feat: security hardening and brute-force protection`

---

## Phase 13: Testing & Validation (Commit #13)

**Duration**: 30 minutes

### Tasks:
- [ ] Create comprehensive test suite:
  - Unit tests for auth package
  - Integration tests for login flow
  - Test config loading from different sources
  - Test CLI flag precedence
  - Test password hashing/verification
  - Test session creation/validation/expiry

- [ ] Test scripts:
  - `test_auth.sh` - test authentication flows
  - `test_config.sh` - test config loading
  - Test with various config combinations
  - Test with missing config
  - Test with invalid config

- [ ] Security testing:
  - Test rate limiting
  - Test session hijacking prevention
  - Test CSRF protection
  - Test with malicious inputs
  - Test authentication bypass attempts

- [ ] Edge cases:
  - Empty config file
  - Corrupted config file
  - Missing config directory
  - No write permissions
  - Database locked
  - Concurrent login attempts

- [ ] Performance testing:
  - Login performance under load
  - Session validation overhead
  - Config reload impact
  - Memory usage with many sessions

- [ ] Add test coverage reporting:
  - Generate coverage report
  - Target: >80% coverage for auth code
  - Add to CI pipeline (if using)

**Commit**: `test: comprehensive auth and config testing`

---

## Phase 14: Documentation (Commit #14)

**Duration**: 30 minutes

### Tasks:
- [ ] Update README.md with:
  - New configuration system explanation
  - Simple setup instructions (using --username and --password flags)
  - Config file format documentation
  - Authentication setup guide
  - CLI flags reference
  - Environment configuration guide
  - Migration guide from v0.1.0

- [ ] Create SECURITY.md:
  - Security best practices
  - How to report vulnerabilities
  - Secure deployment checklist
  - Password requirements
  - Session security notes
  - Audit logging guide

- [ ] Create CONFIGURATION.md:
  - Detailed config reference
  - All available options
  - Examples for different scenarios
  - Environment-specific configs
  - CLI flag priority explanation
  - Troubleshooting common issues

- [ ] Create UPGRADE.md:
  - v0.1.0 to v0.2.0 upgrade guide
  - Breaking changes
  - Migration steps
  - Config file conversion
  - Database migration
  - Rollback procedure

- [ ] Add inline code comments:
  - Document all auth functions
  - Document config loading logic
  - Add examples in comments
  - Explain security decisions

- [ ] Update deployment guide:
  - Include auth setup
  - Config file deployment
  - systemd service with config path
  - nginx auth pass-through
  - Security headers in nginx

**Commit**: `docs: comprehensive documentation for auth and config`

---

## Phase 15: Polish & Final Testing (Commit #15)

**Duration**: 25 minutes

### Tasks:
- [ ] UI polish for login page:
  - Add "Command Center" branding
  - Smooth transitions
  - Better error messages
  - Loading states
  - Success animation

- [ ] Dashboard UI updates:
  - Add user menu in header
  - Show session expiry countdown
  - Auto-logout warning (5 min before expiry)
  - Session refresh on activity

- [ ] Settings page enhancements:
  - Better organization of auth settings
  - Visual indicators for security status
  - Quick actions (change password, logout all sessions)
  - Export config (sanitized)

- [ ] Error message improvements:
  - User-friendly error messages
  - Actionable suggestions
  - Link to documentation
  - Support contact info

- [ ] Code cleanup:
  - Remove debug logs
  - Consistent error handling
  - Code formatting (gofmt)
  - Remove unused code
  - Update dependencies

- [ ] Final testing checklist:
  - [ ] Fresh install with --username/--password flags
  - [ ] Login/logout flow
  - [ ] Session expiry
  - [ ] Password change
  - [ ] Config reload
  - [ ] Rate limiting
  - [ ] All dashboard pages with auth
  - [ ] Tracking endpoints still public
  - [ ] Mobile responsive login
  - [ ] Dark mode login page
  - [ ] Multiple concurrent sessions
  - [ ] Config file validation
  - [ ] CLI flags override config
  - [ ] Default database location
  - [ ] Custom database location
  - [ ] Migration from v0.1.0

**Commit**: `polish: final UI refinements and testing`

---

## Phase 16: Build & Release (Commit #16)

**Duration**: 20 minutes

### Tasks:
- [ ] Version bump to v0.2.0:
  - Update version constant in code
  - Update README.md
  - Update CHANGELOG.md

- [ ] Create release builds:
  ```bash
  # Linux x64
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
    -ldflags="-w -s -X main.Version=v0.2.0" \
    -o cc-server-linux-amd64 ./cmd/server

  # macOS x64
  GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build \
    -ldflags="-w -s -X main.Version=v0.2.0" \
    -o cc-server-darwin-amd64 ./cmd/server

  # macOS ARM64
  GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build \
    -ldflags="-w -s -X main.Version=v0.2.0" \
    -o cc-server-darwin-arm64 ./cmd/server
  ```

- [ ] Create release packages:
  ```bash
  # Linux package
  tar -czf command-center-v0.2.0-linux-amd64.tar.gz \
    cc-server-linux-amd64 \
    web/ \
    migrations/ \
    config.example.json \
    README.md \
    SECURITY.md \
    CONFIGURATION.md \
    UPGRADE.md

  # macOS package (similar)
  ```

- [ ] Update CHANGELOG.md:
  ```markdown
  ## [0.2.0] - 2024-XX-XX

  ### Added
  - JSON configuration file support with CLI flags
  - Authentication system with username/password
  - Session management with secure cookies
  - Simple credential setup via --username/--password flags
  - Configuration management UI
  - Audit logging for security events
  - Database migration system
  - Environment-specific configs
  - Enhanced security features
  - Comprehensive documentation

  ### Changed
  - Config loading now uses JSON instead of env vars (backward compatible)
  - Default database location moved to ~/.config/cc/data.db
  - Dashboard now requires authentication (configurable)

  ### Security
  - Added brute-force protection
  - Implemented rate limiting for login
  - Added CSRF protection
  - Enhanced session security
  - Secure file permissions enforcement
  ```

- [ ] Create GitHub release:
  - Tag: v0.2.0
  - Title: "Command Center v0.2.0 - Auth & Config Upgrade"
  - Attach binary packages
  - Include CHANGELOG excerpt
  - Migration guide link

- [ ] Update deployment documentation:
  - New installation steps
  - Upgrade from v0.1.0 instructions
  - systemd service file with new paths
  - nginx config updates

**Commit**: `release: v0.2.0 - authentication and configuration upgrade`

---

## Migration Guide (v0.1.0 → v0.2.0)

### For Existing Installations:

1. **Backup your database**:
   ```bash
   cp cc.db cc.db.backup
   ```

2. **Stop the server**:
   ```bash
   systemctl stop command-center  # or kill process
   ```

3. **Download v0.2.0 binary**:
   ```bash
   wget https://github.com/yourusername/command-center/releases/download/v0.2.0/command-center-v0.2.0-linux-amd64.tar.gz
   tar -xzf command-center-v0.2.0-linux-amd64.tar.gz
   ```

4. **Set up authentication** (optional but recommended):
   ```bash
   ./cc-server-linux-amd64 --username admin --password your-secure-password
   # This creates the config file at ~/.config/cc/config.json
   ```

5. **Update your systemd service** to use config:
   ```ini
   ExecStart=/opt/command-center/cc-server-linux-amd64 \
     --config /home/user/.config/cc/config.json \
     --db /home/user/.config/cc/data.db
   ```

6. **Copy old database** to new location:
   ```bash
   mkdir -p ~/.config/cc
   cp cc.db ~/.config/cc/data.db
   ```

7. **Run database migrations**:
   ```bash
   ./cc-server-linux-amd64 --migrate --config ~/.config/cc/config.json
   ```

8. **Start server**:
   ```bash
   systemctl start command-center
   ```

9. **Login to dashboard** with credentials you set in step 4

### Backward Compatibility:

- Environment variables still work but are deprecated
- Old database location (`./cc.db`) is supported with warning
- Tracking endpoints remain unauthenticated by default
- All existing redirects and webhooks continue working

---

## Testing Checklist

Before marking complete, verify:

- [ ] --username/--password flags create valid config
- [ ] Server starts with config file
- [ ] Server starts with CLI flags
- [ ] Login works with correct credentials
- [ ] Login fails with wrong credentials
- [ ] Rate limiting prevents brute force
- [ ] Session persists across requests
- [ ] Session expires after timeout
- [ ] Logout clears session
- [ ] Dashboard requires auth when enabled
- [ ] Tracking endpoints don't require auth
- [ ] Config reload works
- [ ] Password change works
- [ ] Audit logs capture events
- [ ] Database migrations run successfully
- [ ] Backup/restore works
- [ ] Multiple environment configs work
- [ ] CLI flags override config file
- [ ] Default paths work (~/.config/cc/)
- [ ] Custom paths work (--db, --config)
- [ ] Mobile login page responsive
- [ ] Dark mode on login page
- [ ] Security headers present
- [ ] HTTPS redirect in production
- [ ] File permissions are restrictive
- [ ] Migration from v0.1.0 works

---

## Success Criteria

By end of implementation:

- [ ] All 16 phases complete with commits
- [ ] Authentication system fully functional
- [ ] Config system uses JSON files
- [ ] --username/--password flags create valid configs
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Binary builds for all platforms
- [ ] Release packages created
- [ ] Migration guide tested
- [ ] Security audit passed
- [ ] Ready for production deployment

---

## Notes

- **Security First**: All auth decisions prioritize security over convenience
- **Backward Compatible**: v0.1.0 installations can upgrade smoothly
- **Configurable**: Auth can be disabled for trusted environments
- **Well Documented**: Every feature has comprehensive docs
- **Production Ready**: Suitable for internet-facing deployment
