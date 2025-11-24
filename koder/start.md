# Implementation Task: v0.4.0 CLI Refactor

**Task:** Implement the v0.4.0 CLI refactor by making all tests pass.

**Approach:** TDD (Test-Driven Development) - tests are written, implement code to make them pass.

---

## Quick Start

```bash
# 1. Understand what you're building
cat TEST_GUIDE_v0.4.0.md

# 2. See current test status (will fail - that's expected)
go test ./cmd/server -v

# 3. Follow the implementation phases below

# 4. Verify success
go test ./cmd/server -v && go test ./internal/config -v && ./test_cli_refactor.sh
```

---

## Implementation Phases

### Phase 1: Remove auth.enabled Field

**File:** `internal/config/config.go`

1. Remove `Enabled bool` field from `AuthConfig` struct
2. Update `CreateDefaultConfig()` - remove `Enabled: false`
3. Update `Validate()` - always require username and password (remove `if cfg.Auth.Enabled` check)
4. Search and remove all `cfg.Auth.Enabled` references in codebase

**Verify:** `go test ./internal/config -v` (all pass)

### Phase 2: Implement Command Functions

**File:** `cmd/server/main.go`

Add these 4 functions (see TEST_GUIDE_v0.4.0.md for complete code):

1. **initCommand(username, password, domain, port, env, configPath string) error**
   - Check config doesn't exist
   - Validate required fields
   - Hash password with bcrypt cost 12
   - Create and save config
   - Verify: `go test ./cmd/server -v -run TestInitCommand`

2. **setCredentialsCommand(username, password, configPath string) error**
   - Require at least one field
   - Load existing config
   - Update provided fields
   - Hash password if provided
   - Save config
   - Verify: `go test ./cmd/server -v -run TestSetCredentials`

3. **setConfigCommand(domain, port, env, configPath string) error**
   - Require at least one field
   - Load existing config
   - Update provided fields
   - Validate config
   - Save config
   - Verify: `go test ./cmd/server -v -run TestSetConfig`

4. **statusCommand(configPath, configDir string) (string, error)**
   - Load config
   - Check PID file for server status
   - Format and return status string
   - Verify: `go test ./cmd/server -v -run TestStatus`

### Phase 3: Add CLI Handlers

**File:** `cmd/server/main.go`

Add handler functions:
- `handleInitCommand()` - Parse flags, call initCommand
- `handleSetConfigCommand()` - Parse flags, call setConfigCommand
- `handleStatusCommand()` - Parse flags, call statusCommand

### Phase 4: Update Routing

**File:** `cmd/server/main.go`

1. **Update handleServerCommand:**
   ```go
   case "init":
       handleInitCommand()
   case "set-config":
       handleSetConfigCommand()
   case "status":
       handleStatusCommand()
   ```

2. **Update main() for deploy alias:**
   ```go
   case "deploy":
       handleDeployCommand()
   ```

3. **Update help texts:**
   - Add new commands to `printServerHelp()`
   - Mention deploy alias in `printUsage()`

4. **Update version:**
   ```go
   const Version = "v0.4.0"
   ```

### Phase 5: Test & Verify

```bash
# All unit tests should pass
go test ./cmd/server -v

# All config tests should pass
go test ./internal/config -v

# Build binary
make build-local

# Integration tests should pass
./test_cli_refactor.sh

# Manual verification
./fazt server init --username admin --password test123 --domain https://test.com
./fazt server status
./fazt server set-config --port 8080
./fazt deploy --help
```

---

## Complete Implementation Guide

**Critical:** Read `TEST_GUIDE_v0.4.0.md` for:
- Complete function implementations with code
- Expected behavior for each function
- Common pitfalls and solutions
- Debugging guide

**File locations:**
- Test specs: `cmd/server/main_test.go`
- Config tests: `internal/config/config_test.go`
- Integration tests: `test_cli_refactor.sh`
- Implementation guide: `TEST_GUIDE_v0.4.0.md`
- Plan document: `koder/plans/06_cli-refactor-init-config.md`

---

## Success Criteria

✅ `go test ./cmd/server -v` - All 40+ tests pass
✅ `go test ./internal/config -v` - All 15+ tests pass
✅ `./test_cli_refactor.sh` - All 40+ checks pass
✅ Manual CLI testing works

---

## Execution Rules

1. **Follow phases in order** - Don't skip ahead
2. **Test after each phase** - Verify before moving on
3. **Reference the guide** - All code examples are in TEST_GUIDE_v0.4.0.md
4. **Let tests guide you** - Failing test shows exactly what's needed
5. **Don't stop** - Complete all phases

---

## If You Get Stuck

1. Read the failing test - shows exact expectations
2. Check TEST_GUIDE_v0.4.0.md for implementation examples
3. Look at similar existing code (e.g., handleSetCredentials)
4. Verify error messages match test expectations

---

**Start now!** Begin with Phase 1. The tests will guide you to success. 🚀
