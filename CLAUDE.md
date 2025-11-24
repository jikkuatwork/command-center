# fazt.sh - Project Context

## Current State

**Version:** v0.3.0 → v0.4.0 (in progress)
**Status:** CLI refactor implementation phase
**Next:** Implement v0.4.0 CLI changes (see `koder/start.md`)

## What Is This?

Personal PaaS combining **analytics tracking** with **static site hosting** and **serverless JavaScript**. Think: Surge.sh + Cloudflare Workers + Plausible Analytics in one binary.

## Quick Architecture

```
cmd/server/main.go (1,159 lines)  # Dual CLI/server entry point
internal/
  ├── hosting/     # PaaS core: JS runtime, deploy, sites
  ├── handlers/    # HTTP endpoints (analytics, deploy, dashboard)
  ├── auth/        # Sessions, rate limiting, passwords
  ├── middleware/  # Security, auth, body limits
  ├── config/      # JSON config with CLI overrides
  └── database/    # SQLite (WAL mode)
migrations/        # SQL schema
examples/          # Sample serverless apps
```

**Key Files:**
- `internal/hosting/runtime.go` - Goja JS runtime with SSRF protection
- `internal/config/config.go` - Configuration management
- `cmd/server/main.go` - CLI command routing

## Core Features

| Feature | Description | Limits |
|---------|-------------|--------|
| **Analytics** | Pageviews, pixels, redirects, webhooks | - |
| **Static Hosting** | Deploy via CLI (Surge-like) | 100MB/file |
| **Serverless JS** | `main.js` with fetch, KV store, WebSocket | 100ms timeout |
| **Key-Value Store** | Persistent data (site-isolated) | - |
| **WebSocket** | Real-time via `/ws` | - |

## CLI Commands (v0.4.0)

```bash
# Server Setup (NEW in v0.4.0)
fazt server init --username admin --password <pass> --domain <url>
fazt server status
fazt server set-credentials --username <user> --password <pass>
fazt server set-config --domain <url> --port <port> --env <env>
fazt server start [--port 8080]
fazt server stop

# Client
fazt client set-auth-token --token <TOKEN>
fazt client deploy --path <PATH> --domain <SUBDOMAIN>

# Shortcuts (NEW in v0.4.0)
fazt deploy --path <PATH> --domain <SUBDOMAIN>  # Alias for client deploy
```

**v0.4.0 Changes:**
- ✅ `init` command for first-time setup
- ✅ `status` command to view configuration
- ✅ `set-config` for updating settings
- ✅ `deploy` alias (top-level shortcut)
- ✅ Auth always required (removed `auth.enabled` flag)

## Configuration

**Location:** `~/.config/fazt/config.json` (0600 permissions)

```json
{
  "server": {"port": "4698", "domain": "https://fazt.sh", "env": "development"},
  "database": {"path": "~/.config/fazt/data.db"},
  "auth": {"username": "admin", "password_hash": "$2a$12$..."},
  "api_key": {"token": "...", "name": "..."}
}
```

**Priority:** CLI flags > Config file > Environment > Defaults

**Note:** `auth.enabled` field removed in v0.4.0 - auth is always required.

## Database Schema

```sql
events       -- Analytics data
api_keys     -- Deploy authentication (bcrypt)
kv_store     -- Serverless key-value (site-isolated)
env_vars     -- Environment variables (site-scoped)
deployments  -- Deployment history
redirects    -- URL shortening
webhooks     -- Webhook endpoints
```

See `migrations/*.sql` for full schema.

## JavaScript Runtime API

```javascript
// Request/Response
req.method, req.path, req.query, req.headers, req.body
res.send(html), res.json(obj), res.status(code), res.header(k, v)

// Storage (site-isolated)
db.get(key), db.set(key, value), db.delete(key)

// Environment & WebSocket
process.env.MY_SECRET
socket.broadcast(msg), socket.clients()

// HTTP (SSRF protected)
const resp = fetch(url, options)  // {status, body, headers, error}
```

See `internal/hosting/runtime.go` for implementation details.

## HTTP Endpoints

| Endpoint | Purpose | Auth |
|----------|---------|------|
| `/track` | Analytics tracking | None |
| `/pixel.gif` | Tracking pixel | None |
| `/r/{slug}` | URL redirect | None |
| `/webhook/{name}` | Webhook receiver | None |
| `/api/stats` | Dashboard stats | Required |
| `/api/deploy` | Deploy site | Bearer token |
| `/api/sites` | List sites | Required |
| `/login` | Web login page | None |
| `/` (subdomain) | Hosted sites | None |

## Request Routing

```
Main domain/localhost → Dashboard (auth required)
subdomain.domain      → Static files or main.js (serverless)

Middleware chain: Tracing → Logging → Body Limit → Security → CORS → Auth
```

## Security Constraints

| Feature | Limit | Reason |
|---------|-------|--------|
| JS execution | 100ms | Prevent runaway scripts |
| File upload | 100MB/file | Reasonable site size |
| Request body | 1MB (100MB deploy) | DoS prevention |
| Deploy rate | 5/min per IP | Rate limiting |
| KV/env isolation | Site-scoped | Multi-tenancy security |
| SSRF protection | Block internal IPs | Prevent network attacks |

## Common Development Tasks

| Task | Action |
|------|--------|
| Add API endpoint | Handler in `internal/handlers/`, register in `main.go` |
| Add JS function | Edit `internal/hosting/runtime.go` vm.Set() |
| Add DB table | Migration in `migrations/`, update schema |
| Update CLI | Modify `cmd/server/main.go` command handlers |
| Run tests | `go test ./... -v` |
| Build | `go build -o fazt ./cmd/server/` or `make build-local` |

## Important Files Reference

**Plans & Docs:**
- `koder/plans/` - Design documents for all phases
- `koder/plans/06_cli-refactor-init-config.md` - v0.4.0 CLI design
- `TEST_GUIDE_v0.4.0.md` - Complete v0.4.0 implementation guide
- `koder/start.md` - Current implementation task

**Tests:**
- `cmd/server/main_test.go` - CLI command tests (40+ cases)
- `internal/config/config_test.go` - Config validation tests
- `test_cli_refactor.sh` - Integration tests (40+ checks)

**Core Code:**
- `cmd/server/main.go` - CLI entry point & command routing
- `internal/config/config.go` - Configuration management
- `internal/hosting/runtime.go` - JavaScript runtime
- `internal/auth/password.go` - Bcrypt password hashing
- `internal/database/db.go` - Database operations

## Current Development Status

**Phase:** v0.4.0 CLI Refactor
**Status:** Implementation ready (TDD approach)
**What's Done:**
- ✅ Complete test suite (95+ tests)
- ✅ Implementation guide with all code
- ✅ Integration tests
- ✅ Config tests updated

**What's Next:**
1. Remove `auth.enabled` field from config
2. Implement 4 command functions
3. Add CLI handlers
4. Update routing
5. Verify all tests pass

**To implement:** See `koder/start.md` → Points to `TEST_GUIDE_v0.4.0.md`

## Build & Test Commands

```bash
# Development
go build -o fazt ./cmd/server/        # Build binary
make build-local                       # Or use Makefile
go test ./... -v                       # All tests
go test ./cmd/server -v                # CLI tests
go test ./internal/config -v           # Config tests

# Integration
./test_cli_refactor.sh                 # Full CLI workflow tests
./test_auth_flow.sh                    # Auth flow tests

# Usage
./fazt server init --username admin --password test123 --domain https://test.com
./fazt server status
./fazt deploy --path ./site --domain myapp
```

## Design Principles

1. **Security by default** - Auth required, secure permissions (0600/0700)
2. **Simple deployment** - Single binary, auto-setup config
3. **Clear CLI** - Hierarchical commands, helpful errors
4. **Site isolation** - KV store and env vars scoped by site
5. **Resource limits** - Timeouts, rate limits, size caps
6. **TDD approach** - Tests define behavior, guide implementation

## When Working on This Project

**Always:**
- Read relevant test files first to understand expectations
- Run tests after changes: `go test ./... -v`
- Check security implications (SSRF, injection, permissions)
- Use parameterized SQL queries
- Validate user inputs

**Never:**
- Skip authentication checks
- Use plaintext passwords
- Ignore file permissions
- Allow unbounded resource usage
- Break site isolation

**For v0.4.0 Implementation:**
Read `koder/start.md` for step-by-step instructions. All code examples in `TEST_GUIDE_v0.4.0.md`.

---

**Quick Links:**
- Implementation: `koder/start.md`
- Tests: `cmd/server/main_test.go`, `test_cli_refactor.sh`
- Design: `koder/plans/06_cli-refactor-init-config.md`
- Guide: `TEST_GUIDE_v0.4.0.md`
