# Test Suite Guide for v0.4.0 CLI Refactor

## Overview

This guide provides comprehensive information about the test suite for the v0.4.0 CLI refactor. The tests are written in TDD (Test-Driven Development) style, meaning **the tests exist before the implementation**. Your job is to implement the features to make all tests pass.

## Test Structure

### 1. Unit Tests (`cmd/server/main_test.go`)
**Location:** `/home/testman/workspace/cmd/server/main_test.go`
**Purpose:** Test individual command functions in isolation
**Coverage:** ~40 test cases covering all new commands

### 2. Config Tests (`internal/config/config_test.go`)
**Location:** `/home/testman/workspace/internal/config/config_test.go`
**Purpose:** Test configuration structure and validation for v0.4.0
**Coverage:** Config validation, auth requirements, file operations

### 3. Integration Tests (`test_cli_refactor.sh`)
**Location:** `/home/testman/workspace/test_cli_refactor.sh`
**Purpose:** End-to-end workflow testing with real CLI invocations
**Coverage:** Complete user workflows from init through status

---

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./cmd/server -v

# Run specific test
go test ./cmd/server -v -run TestInitCommand_Success

# Run with coverage
go test ./cmd/server -v -cover

# Run config tests
go test ./internal/config -v
```

### Integration Tests

```bash
# First, build the binary
go build -o fazt ./cmd/server/

# Or use make
make build-local

# Then run integration tests
./test_cli_refactor.sh
```

---

## Functions You Need to Implement

The test suite expects these functions to exist in `cmd/server/main.go`:

### 1. `initCommand(username, password, domain, port, env, configPath string) error`

**Purpose:** Initialize server configuration for first-time setup

**Requirements:**
- Check if config already exists at `configPath`
  - If exists: return error "Server already initialized" or similar
- Validate required parameters:
  - username: must not be empty
  - password: must not be empty
  - domain: must not be empty
  - port: must be valid (1-65535)
  - env: must be "development" or "production"
- Create config directory with 0700 permissions
- Hash password using bcrypt (cost 12)
- Create config with all values:
  ```json
  {
    "server": {"port": port, "domain": domain, "env": env},
    "database": {"path": "{configDir}/data.db"},
    "auth": {"username": username, "password_hash": bcrypt_hash},
    "ntfy": {"topic": "", "url": "https://ntfy.sh"}
  }
  ```
- Save config with 0600 permissions
- Return nil on success, descriptive error on failure

**Example:**
```go
func initCommand(username, password, domain, port, env, configPath string) error {
    // Check if config exists
    if _, err := os.Stat(configPath); err == nil {
        return fmt.Errorf("Error: Server already initialized\nConfig exists at: %s", configPath)
    }

    // Validate inputs
    if username == "" || password == "" || domain == "" {
        return errors.New("Error: username, password, and domain are required")
    }

    // ... implement the rest
    return nil
}
```

**Tests:** 8 test cases
- TestInitCommand_Success
- TestInitCommand_ConfigAlreadyExists
- TestInitCommand_MissingRequiredFlags
- TestInitCommand_InvalidPort
- TestInitCommand_InvalidEnvironment
- TestInitCommand_SecurePermissions

---

### 2. `setCredentialsCommand(username, password, configPath string) error`

**Purpose:** Update username and/or password in existing config

**Requirements:**
- Load existing config from `configPath`
  - If not found: return error "Config not found" or "Server not initialized"
- Validate: at least one of username or password must be provided
  - If neither: return error "Error: at least one of --username or --password is required"
- Update provided fields only:
  - If username provided: update `cfg.Auth.Username`
  - If password provided: hash with bcrypt and update `cfg.Auth.PasswordHash`
  - Preserve other fields
- Save config back to file
- Return nil on success, descriptive error on failure

**Example:**
```go
func setCredentialsCommand(username, password, configPath string) error {
    // Validate at least one provided
    if username == "" && password == "" {
        return errors.New("Error: at least one of --username or --password is required")
    }

    // Load config
    cfg, err := config.LoadFromFile(configPath)
    if err != nil {
        return fmt.Errorf("Error: Config not found at %s\nRun 'fazt server init' first", configPath)
    }

    // Update fields
    if username != "" {
        cfg.Auth.Username = username
    }
    if password != "" {
        hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
        cfg.Auth.PasswordHash = string(hash)
    }

    // Save
    return config.SaveToFile(cfg, configPath)
}
```

**Tests:** 6 test cases
- TestSetCredentials_UpdatePassword
- TestSetCredentials_UpdateUsername
- TestSetCredentials_UpdateBoth
- TestSetCredentials_NoConfigExists
- TestSetCredentials_NoFlagsProvided

---

### 3. `setConfigCommand(domain, port, env, configPath string) error`

**Purpose:** Update server configuration settings

**Requirements:**
- Load existing config from `configPath`
  - If not found: return error
- Validate: at least one field must be provided
  - If none: return error "Error: at least one flag is required"
- Validate port if provided (1-65535)
- Validate env if provided ("development" or "production")
- Update provided fields only:
  - If domain != "": update `cfg.Server.Domain`
  - If port != "": update `cfg.Server.Port`
  - If env != "": update `cfg.Server.Env`
  - Preserve other fields
- Validate the updated config using `cfg.Validate()`
- Save config back to file
- Return nil on success, descriptive error on failure

**Example:**
```go
func setConfigCommand(domain, port, env, configPath string) error {
    // Validate at least one provided
    if domain == "" && port == "" && env == "" {
        return errors.New("Error: at least one of --domain, --port, or --env is required")
    }

    // Load config
    cfg, err := config.LoadFromFile(configPath)
    if err != nil {
        return fmt.Errorf("Error: Config not found")
    }

    // Update fields
    if domain != "" {
        cfg.Server.Domain = domain
    }
    if port != "" {
        cfg.Server.Port = port
    }
    if env != "" {
        cfg.Server.Env = env
    }

    // Validate
    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("Error: Invalid configuration: %v", err)
    }

    // Save
    return config.SaveToFile(cfg, configPath)
}
```

**Tests:** 10 test cases
- TestSetConfig_UpdateDomain
- TestSetConfig_UpdatePort
- TestSetConfig_UpdateEnvironment
- TestSetConfig_UpdateMultipleFields
- TestSetConfig_NoFlagsProvided
- TestSetConfig_InvalidPort
- TestSetConfig_InvalidEnvironment
- TestSetConfig_NoConfigExists

---

### 4. `statusCommand(configPath, configDir string) (string, error)`

**Purpose:** Display current configuration and server status

**Requirements:**
- Load config from `configPath`
  - If not found: return error
- Check if server is running:
  - Read PID file at `{configDir}/cc-server.pid`
  - If exists and contains valid PID: server is running
  - Otherwise: server is stopped
- Get database file size (if exists)
- Get sites directory info (if exists)
- Format and return status string with all information
- Return formatted string and nil on success, empty string and error on failure

**Expected Output Format:**
```
Server Status
═══════════════════════════════════════════════════════════
Config:       ~/.config/fazt/config.json
Domain:       https://fazt.toolbomber.com
Port:         4698
Environment:  production
Username:     admin
Database:     ~/.config/fazt/data.db (42.3 MB)
Sites:        ~/.config/fazt/sites/ (3 sites, 15.2 MB)

Server:       ● Running (PID: 12345, uptime: 2d 5h)
              ○ Not running
```

**Example:**
```go
func statusCommand(configPath, configDir string) (string, error) {
    // Load config
    cfg, err := config.LoadFromFile(configPath)
    if err != nil {
        return "", fmt.Errorf("Error: Config not found")
    }

    var output strings.Builder
    output.WriteString("Server Status\n")
    output.WriteString("═══════════════════════════════════════\n")
    output.WriteString(fmt.Sprintf("Config:       %s\n", configPath))
    output.WriteString(fmt.Sprintf("Domain:       %s\n", cfg.Server.Domain))
    output.WriteString(fmt.Sprintf("Port:         %s\n", cfg.Server.Port))
    output.WriteString(fmt.Sprintf("Environment:  %s\n", cfg.Server.Env))
    output.WriteString(fmt.Sprintf("Username:     %s\n", cfg.Auth.Username))

    // Check PID file
    pidFile := filepath.Join(configDir, "cc-server.pid")
    if pidData, err := os.ReadFile(pidFile); err == nil {
        output.WriteString(fmt.Sprintf("\nServer:       ● Running (PID: %s)\n", strings.TrimSpace(string(pidData))))
    } else {
        output.WriteString("\nServer:       ○ Not running\n")
    }

    return output.String(), nil
}
```

**Tests:** 4 test cases
- TestStatus_OutputFormat
- TestStatus_ServerRunning
- TestStatus_ServerStopped
- TestStatus_NoConfigExists

---

## CLI Integration

### Update `handleServerCommand()` in main.go

Add new cases to the switch statement:

```go
func handleServerCommand(args []string) {
    if len(args) < 1 {
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
    // ... other cases
    }
}
```

### Create Handler Functions

Each handler should:
1. Create a FlagSet
2. Define flags
3. Parse os.Args[3:]
4. Validate required flags
5. Call the command function
6. Handle errors and exit codes

**Example for init:**
```go
func handleInitCommand() {
    flags := flag.NewFlagSet("init", flag.ExitOnError)
    username := flags.String("username", "", "Admin username (required)")
    password := flags.String("password", "", "Admin password (required)")
    domain := flags.String("domain", "", "Server domain (required)")
    port := flags.String("port", "4698", "Server port")
    env := flags.String("env", "development", "Environment (development|production)")

    flags.Usage = func() {
        fmt.Println("Usage: fazt server init [flags]")
        fmt.Println()
        fmt.Println("Initialize fazt.sh server configuration")
        fmt.Println()
        flags.PrintDefaults()
    }

    if err := flags.Parse(os.Args[3:]); err != nil {
        os.Exit(1)
    }

    // Get config path
    homeDir, _ := os.UserHomeDir()
    configPath := filepath.Join(homeDir, ".config", "fazt", "config.json")

    // Call command function
    if err := initCommand(*username, *password, *domain, *port, *env, configPath); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    fmt.Println("✓ Server initialized successfully")
    fmt.Printf("  Config saved to: %s\n", configPath)
}
```

### Add Deploy Alias

Update main() to handle "deploy" as a top-level command:

```go
func main() {
    if len(os.Args) < 2 {
        printUsage()
        return
    }

    command := os.Args[1]

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
```

---

## Config Changes

### Remove `auth.enabled` Field

**File:** `internal/config/config.go`

**Changes needed:**

1. Update `AuthConfig` struct:
```go
type AuthConfig struct {
    // Enabled field removed in v0.4.0
    Username     string `json:"username"`
    PasswordHash string `json:"password_hash"`
}
```

2. Update `Validate()` method:
```go
func (c *Config) Validate() error {
    // ... existing validation ...

    // v0.4.0: Auth always required
    if c.Auth.Username == "" {
        return errors.New("auth username is required")
    }
    if c.Auth.PasswordHash == "" {
        return errors.New("auth password hash is required")
    }

    return nil
}
```

3. Update `CreateDefaultConfig()`:
```go
func CreateDefaultConfig() *Config {
    return &Config{
        // ...
        Auth: AuthConfig{
            // No Enabled field
            // Username and PasswordHash empty (to be set by init)
            Username:     "",
            PasswordHash: "",
        },
        // ...
    }
}
```

4. Remove all references to `cfg.Auth.Enabled` throughout the codebase

---

## TDD Workflow

### Step-by-Step Implementation Process

1. **Start with failing tests:**
   ```bash
   go test ./cmd/server -v
   # All tests should fail with "undefined: initCommand" etc.
   ```

2. **Implement initCommand:**
   - Add the function signature
   - Run tests: `go test ./cmd/server -v -run TestInitCommand`
   - Implement until all init tests pass

3. **Implement setCredentialsCommand:**
   - Run tests: `go test ./cmd/server -v -run TestSetCredentials`
   - Implement until all tests pass

4. **Implement setConfigCommand:**
   - Run tests: `go test ./cmd/server -v -run TestSetConfig`
   - Implement until all tests pass

5. **Implement statusCommand:**
   - Run tests: `go test ./cmd/server -v -run TestStatus`
   - Implement until all tests pass

6. **Run full workflow test:**
   ```bash
   go test ./cmd/server -v -run TestFullWorkflow
   ```

7. **Update config package:**
   ```bash
   go test ./internal/config -v
   ```

8. **Add CLI handlers and routing**

9. **Run integration tests:**
   ```bash
   make build-local
   ./test_cli_refactor.sh
   ```

10. **All tests should pass!**

---

## Test Coverage Goals

### Target Coverage
- **Unit Tests:** 80%+ coverage for new command functions
- **Config Tests:** 100% coverage for validation logic
- **Integration Tests:** All critical user workflows

### Checking Coverage

```bash
# Overall coverage
go test ./cmd/server -cover

# Detailed coverage report
go test ./cmd/server -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage for config package
go test ./internal/config -cover
```

---

## Common Implementation Pitfalls

### 1. File Permissions
**Problem:** Config file created with wrong permissions
**Solution:** Always use 0600 for config files, 0700 for directories

### 2. Password Hashing
**Problem:** Forgetting to hash password or using wrong cost
**Solution:** Use `bcrypt.GenerateFromPassword([]byte(password), 12)`

### 3. Preserving Fields
**Problem:** Overwriting entire config instead of updating specific fields
**Solution:** Load config first, update only specified fields, then save

### 4. Error Messages
**Problem:** Generic errors that don't help user
**Solution:** Include context: what failed, what file, what to do next

```go
// Bad
return errors.New("config error")

// Good
return fmt.Errorf("Error: Config not found at %s\n\nRun 'fazt server init' to create configuration", configPath)
```

### 5. Validation Order
**Problem:** Validating after saving (too late)
**Solution:** Validate inputs → load config → update → validate again → save

---

## Debugging Failed Tests

### Unit Test Failures

```bash
# Run single test with verbose output
go test ./cmd/server -v -run TestInitCommand_Success

# Check what the test expects
cat cmd/server/main_test.go | grep -A 20 "TestInitCommand_Success"
```

### Integration Test Failures

```bash
# Run with set -x for debug output
bash -x ./test_cli_refactor.sh

# Check specific test section
./test_cli_refactor.sh 2>&1 | grep -A 10 "Testing 'fazt server init'"
```

### Config Test Failures

```bash
# Run config tests
go test ./internal/config -v

# Run specific config test
go test ./internal/config -v -run TestConfigValidation_AlwaysRequiresAuth
```

---

## Success Criteria

### You're done when:

1. ✅ All unit tests pass:
   ```bash
   go test ./cmd/server -v
   # All tests pass
   ```

2. ✅ Config tests pass:
   ```bash
   go test ./internal/config -v
   # All tests pass
   ```

3. ✅ Integration tests pass:
   ```bash
   ./test_cli_refactor.sh
   # All integration tests passed!
   ```

4. ✅ Manual testing works:
   ```bash
   ./fazt server init --username admin --password test123 --domain https://test.com
   ./fazt server status
   ./fazt server set-config --domain https://new.com
   ./fazt deploy --help
   ```

---

## Additional Resources

### Plan Document
- **File:** `koder/plans/06_cli-refactor-init-config.md`
- Contains detailed specifications and design decisions

### Existing Tests to Reference
- `internal/auth/password_test.go` - Password hashing examples
- `internal/config/config_test.go` - Config testing patterns
- `test_auth_flow.sh` - Integration test examples

### Go Testing Resources
- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [bcrypt Documentation](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

---

## Questions?

If you encounter issues:

1. Read the test that's failing - it shows exactly what's expected
2. Check the function signature and requirements in this guide
3. Look at similar existing code (e.g., `handleSetCredentials`)
4. Verify your error messages match what tests expect
5. Run tests incrementally, one function at a time

**Remember:** The tests are your specification. If a test fails, it's telling you exactly what needs to be fixed!

Good luck! 🚀
