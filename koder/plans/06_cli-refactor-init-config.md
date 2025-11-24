# CLI Refactor: Init, Config Management, and Auth Simplification

**Status:** Planning
**Version:** v0.4.0
**Date:** 2025-11-24

## Overview

Refactor the CLI command structure to be more intuitive, scalable, and secure by default. This addresses issues with the current `set-credentials` command becoming overloaded, inconsistent config management, and optional authentication.

## Motivation

### Current Problems
1. **Semantic confusion**: `set-credentials --domain` doesn't make sense
2. **Command explosion risk**: If we add `set-domain`, `set-port`, `set-env`, etc., we'll have too many commands
3. **Unclear initialization**: No clear "first-time setup" command
4. **Optional auth**: `auth.enabled` flag is a security risk
5. **No visibility**: Can't easily see current configuration
6. **No deployment shortcut**: Most common operation requires `client deploy`

### Goals
- ✅ Clear first-time setup experience
- ✅ Unified config management
- ✅ Security by default (auth always required)
- ✅ Scalable command structure
- ✅ Better discoverability
- ✅ Convenience shortcuts for common operations

## Command Structure Changes

### Before (v0.3.0)
```bash
# Server commands
fazt server set-credentials --username admin --password secret123
fazt server start [--port 8080] [--domain https://example.com]
fazt server stop

# Client commands
fazt client set-auth-token --token <TOKEN>
fazt client deploy --path <PATH> --domain <SUBDOMAIN>
```

### After (v0.4.0)
```bash
# Server commands
fazt server init --username admin --password secret123 --domain https://fazt.example.com [--port 4698] [--env production]
fazt server set-credentials --username admin --password newsecret
fazt server set-config --domain https://new.example.com [--port 8080] [--env production]
fazt server status
fazt server start [--port 8080]
fazt server stop

# Client commands (unchanged)
fazt client set-auth-token --token <TOKEN>
fazt client deploy --path <PATH> --domain <SUBDOMAIN>

# Convenience aliases
fazt deploy --path <PATH> --domain <SUBDOMAIN>  # Alias for 'client deploy'
```

## Detailed Command Specifications

### 1. `fazt server init`

**Purpose:** First-time server initialization

**Flags:**
- `--username <string>` (required) - Admin username
- `--password <string>` (required) - Admin password
- `--domain <string>` (required) - Server domain (e.g., https://fazt.example.com)
- `--port <string>` (optional, default: 4698) - Server port
- `--env <string>` (optional, default: development) - Environment (development|production)

**Behavior:**
- Creates `~/.config/fazt/` directory with secure permissions (0700)
- Generates `config.json` with all settings
- Hashes password with bcrypt (cost 12)
- Sets secure file permissions (0600 for config.json)
- **Fails if config already exists** with error message:
  ```
  Error: Server already initialized
  Config exists at: ~/.config/fazt/config.json

  To update configuration:
    - Change credentials: fazt server set-credentials
    - Change settings: fazt server set-config
    - View current config: fazt server status
  ```

**Examples:**
```bash
# Minimal (use defaults)
fazt server init --username admin --password secret123 --domain https://fazt.example.com

# Full control
fazt server init \
  --username admin \
  --password secret123 \
  --domain https://fazt.toolbomber.com \
  --port 8080 \
  --env production
```

### 2. `fazt server set-credentials`

**Purpose:** Update username and/or password

**Flags:**
- `--username <string>` (optional) - New username
- `--password <string>` (optional) - New password

**Behavior:**
- Loads existing config
- Updates username if provided
- Hashes new password if provided
- Saves config back
- At least one flag must be provided

**Examples:**
```bash
# Change password only
fazt server set-credentials --password newsecret

# Change username only
fazt server set-credentials --username newadmin

# Change both
fazt server set-credentials --username newadmin --password newsecret
```

### 3. `fazt server set-config` (NEW)

**Purpose:** Update server configuration settings

**Flags:**
- `--domain <string>` (optional) - Server domain
- `--port <string>` (optional) - Server port
- `--env <string>` (optional) - Environment (development|production)

**Behavior:**
- Loads existing config
- Updates specified fields only
- Validates values (port 1-65535, env must be development|production)
- Preserves other settings
- At least one flag must be provided

**Examples:**
```bash
# Change domain
fazt server set-config --domain https://new.example.com

# Change port
fazt server set-config --port 8080

# Change environment
fazt server set-config --env production

# Change multiple at once
fazt server set-config --domain https://prod.example.com --port 443 --env production
```

### 4. `fazt server status` (NEW)

**Purpose:** Display current configuration and server status

**Flags:** None

**Behavior:**
- Reads config file
- Checks if server is running (via PID file)
- Displays formatted output

**Output:**
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

API Keys:     2 active keys
Last Deploy:  2025-11-24 10:30:45 (my-app by token-1)
```

### 5. `fazt deploy` (NEW ALIAS)

**Purpose:** Convenience shortcut for `fazt client deploy`

**Behavior:**
- Exactly equivalent to `fazt client deploy`
- All flags pass through unchanged
- Implemented as direct command routing, not shell alias

**Examples:**
```bash
# These are identical
fazt deploy --path ./site --domain myapp
fazt client deploy --path ./site --domain myapp
```

## Configuration Structure Changes

### Before
```json
{
  "server": {
    "port": "4698",
    "domain": "https://fazt.sh",
    "env": "development"
  },
  "database": {
    "path": "~/.config/fazt/data.db"
  },
  "auth": {
    "enabled": true,              ← REMOVE THIS
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "api_key": {
    "token": "...",
    "name": "..."
  },
  "ntfy": {
    "topic": "",
    "url": "https://ntfy.sh"
  }
}
```

### After
```json
{
  "server": {
    "port": "4698",
    "domain": "https://fazt.toolbomber.com",
    "env": "production"
  },
  "database": {
    "path": "~/.config/fazt/data.db"
  },
  "auth": {
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "api_key": {
    "token": "...",
    "name": "..."
  },
  "ntfy": {
    "topic": "",
    "url": "https://ntfy.sh"
  }
}
```

**Changes:**
- ❌ Remove `auth.enabled` field
- ✅ Auth is always required
- ✅ Domain is explicitly set during init
- ✅ Config must have valid auth credentials

## Implementation Plan

### Phase 1: Config Structure Changes
**Files:** `internal/config/config.go`

1. Remove `Enabled` field from `AuthConfig` struct
2. Update `CreateDefaultConfig()` - remove enabled field
3. Update `Validate()` - always require credentials
4. Remove any `auth.enabled` checks
5. Update config examples

### Phase 2: Implement New Commands
**Files:** `cmd/server/main.go`

1. **Implement `fazt server init`:**
   - Parse flags (username, password, domain, port, env)
   - Check if config exists, error if it does
   - Create config with all settings
   - Hash password
   - Save with secure permissions
   - Print success message with next steps

2. **Refactor `fazt server set-credentials`:**
   - Remove initialization logic
   - Focus only on updating credentials
   - Require at least one flag
   - Load, update, save pattern

3. **Implement `fazt server set-config`:**
   - Parse flags (domain, port, env)
   - Require at least one flag
   - Load existing config
   - Update specified fields
   - Validate
   - Save config

4. **Implement `fazt server status`:**
   - Load config file
   - Check PID file for server status
   - Get database file size
   - Count sites in sites directory
   - Query API keys from database
   - Format and display output

5. **Implement `fazt deploy` alias:**
   - Add routing in main() switch
   - Call handleDeployCommand() directly
   - Update help text

### Phase 3: Update Help & Documentation
**Files:** `cmd/server/main.go`, `README.md`, `CLAUDE.md`, `CONFIGURATION.md`

1. Update all help messages
2. Update command examples in docs
3. Add migration guide
4. Update quick start guides

### Phase 4: Testing

**Unit Tests** (`cmd/server/main_test.go` - NEW FILE)

```go
// Test init command
- TestInitCommand_Success
- TestInitCommand_ConfigExists
- TestInitCommand_MissingRequiredFlags
- TestInitCommand_InvalidPort
- TestInitCommand_InvalidEnv
- TestInitCommand_SecurePermissions

// Test set-credentials command
- TestSetCredentials_Success
- TestSetCredentials_NoConfigExists
- TestSetCredentials_NoFlagsProvided
- TestSetCredentials_PasswordHashing

// Test set-config command
- TestSetConfig_Domain
- TestSetConfig_Port
- TestSetConfig_Env
- TestSetConfig_MultipleFlags
- TestSetConfig_NoFlagsProvided
- TestSetConfig_InvalidValues

// Test status command
- TestStatus_ServerRunning
- TestStatus_ServerStopped
- TestStatus_NoConfig
- TestStatus_OutputFormat
```

**Integration Tests** (`test_cli_refactor.sh` - NEW FILE)

```bash
#!/bin/bash
# Test full workflow

# 1. Init prevents double initialization
# 2. Set-credentials updates credentials
# 3. Set-config updates settings
# 4. Status shows correct information
# 5. Server start requires init
# 6. Deploy alias works
```

**Regression Tests**
- Ensure `test_auth_flow.sh` still passes
- Update any tests that reference old commands

### Phase 5: Migration Guide

**Breaking Changes:**
1. `auth.enabled` field removed from config
2. `server start` now requires prior initialization
3. Config structure change (minor)

**Migration Steps:**

For users upgrading from v0.3.0:

```bash
# 1. Backup existing config
cp ~/.config/fazt/config.json ~/.config/fazt/config.json.backup

# 2. Manual edit: Remove "enabled" field from auth section
# Edit ~/.config/fazt/config.json and remove:
#   "enabled": true,

# 3. Or: Reinitialize (recommended)
rm ~/.config/fazt/config.json
fazt server init --username admin --password secret123 --domain https://your-domain.com

# 4. Restore API keys if needed
# (Keys are in database, not affected)
```

**Automated Migration Script** (`migrate-v0.3-to-v0.4.sh`):
```bash
#!/bin/bash
# Automatically migrate config from v0.3 to v0.4
# - Removes auth.enabled field
# - Preserves all other settings
```

## Testing Strategy

### Unit Test Coverage
- **Target:** 80%+ coverage for new command handlers
- **Focus:** Command parsing, validation, error handling
- **Mock:** File system operations, config loading

### Integration Test Coverage
- Full command workflows
- Error scenarios
- Permission checks
- Config file integrity

### Manual Testing Checklist
- [ ] Init creates config with correct values
- [ ] Init fails on second run
- [ ] Set-credentials updates password
- [ ] Set-config updates domain
- [ ] Status shows server running/stopped
- [ ] Deploy alias works
- [ ] Server start fails without init
- [ ] All help messages are correct
- [ ] Migration from v0.3 config works

## Security Considerations

1. **Auth Always Required**
   - Server won't start without credentials
   - No way to disable authentication
   - Prevents accidental public deployments

2. **Secure Defaults**
   - Config directory: 0700
   - Config file: 0600
   - Database: 0600

3. **Password Handling**
   - Never store plaintext
   - Bcrypt with cost factor 12
   - No password in logs or output

4. **Validation**
   - Port range: 1-65535
   - Environment: development|production only
   - Domain: Must be valid URL

## Backward Compatibility

### What Breaks
- `auth.enabled` field in config (removed)
- Server behavior without config (now fails, previously used defaults)
- `set-credentials` as initialization method (now separate `init` command)

### What Doesn't Break
- `client deploy` command (unchanged)
- `server start/stop` commands (unchanged, but start now requires init)
- Config file location (still `~/.config/fazt/`)
- API key management (unchanged)

## Success Criteria

- [ ] All new commands implemented and working
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Documentation updated
- [ ] Migration guide written
- [ ] No regression in existing functionality
- [ ] Help messages are clear and helpful
- [ ] `fazt server status` provides useful information

## Timeline

**Estimated:** 4-6 hours of development + testing

1. **Phase 1:** Config changes (30 min)
2. **Phase 2:** Command implementation (2 hours)
3. **Phase 3:** Documentation (1 hour)
4. **Phase 4:** Testing (1-2 hours)
5. **Phase 5:** Migration guide (30 min)

## Future Enhancements (Not in v0.4.0)

- Interactive `init` command (prompt for values)
- `fazt server config edit` - Open config in $EDITOR
- `fazt server logs` - View recent logs
- Tab completion for bash/zsh
- Config validation command: `fazt server config validate`

## Questions/Decisions Needed

- [ ] Should `init` support `--force` flag to overwrite existing config?
- [ ] Should `status` command show deployed sites list?
- [ ] Should we add `fazt logs` as another top-level alias?
- [ ] Migration script: automatic or manual?

---

**Next Steps:**
1. Review this plan
2. Get approval on command structure
3. Begin implementation
4. Write tests alongside implementation
5. Update documentation
6. Create migration guide
