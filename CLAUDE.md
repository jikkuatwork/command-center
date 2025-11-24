# Command Center - AI Assistant Context

## Project Overview
Command Center v0.3.0 is a Personal PaaS combining analytics tracking with static site hosting and serverless JavaScript.

**Core Features**:
- Analytics: Pageviews, pixels, redirects, webhooks
- Static Site Hosting: Surge.sh-like deployment via CLI
- Serverless JavaScript: Run JS with `main.js` (100ms timeout)
- Key-Value Store: Persistent data for serverless apps
- WebSocket Support: Real-time communication via `/ws`

## Architecture
```
cmd/server/main.go       # Entry point (1,159 lines), dual CLI/server modes
internal/
  hosting/               # PaaS core (6 files)
    runtime.go           # Goja JS runtime with SSRF protection
    deploy.go            # ZIP extraction, API keys
    manager.go           # Site CRUD operations
  handlers/              # HTTP handlers (10 files)
  auth/                  # Session management, rate limiting
  middleware/            # Security, auth, tracing
  database/              # SQLite with WAL mode, migrations
  config/                # JSON config with CLI overrides
migrations/              # SQL schema (3 files)
examples/                # Sample apps
```

## Key Commands
```bash
# Build & Test
go build -o cc-server ./cmd/server/
go test ./... -v

# CLI Commands (subcommand-based interface)
./cc-server set-credentials --username admin --password secret123  # Set up auth
./cc-server start [--port 8080] [--config path]                     # Start server
./cc-server deploy --path <PATH> --domain <SUBDOMAIN>               # Deploy site
./cc-server stop                                                     # Stop server
./cc-server --version                                                # Show version
./cc-server --help                                                   # Show help

# Common workflows
./cc-server set-credentials --username admin --password secret123 && ./cc-server start
./cc-server deploy --path . --domain my-app                         # Deploy current dir
./cc-server deploy --path ~/project/build --domain app --server https://cc.example.com
```

## Database Schema
- `events` - Analytics events
- `api_keys` - Deploy tokens (bcrypt hashed)
- `kv_store` - Serverless data (site-isolated)
- `env_vars` - Environment variables (site-scoped)
- `deployments` - Deployment history
- `redirects` - URL shortener
- `webhooks` - Webhook endpoints

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

// HTTP requests (SSRF protected)
const resp = fetch(url, options)  // { status, body, headers, error }

// Logging
console.log(...)
```

## Configuration
**File**: `~/.config/cc/config.json` (auto-created with secure `0700` permissions)
**Priority**: CLI flags > Config file > Environment > Defaults

```json
{
  "server": { "port": "4698", "domain": "https://cc.toolbomber.com", "env": "development" },
  "database": { "path": "~/.config/cc/data.db" },
  "auth": { "enabled": true, "username": "admin", "password_hash": "bcrypt_hash" },
  "api_key": { "token": "generated_token", "name": "auto-generated" },
  "ntfy": { "topic": "", "url": "https://ntfy.sh" }
}
```

**Directory Creation**: `~/.config/cc/` automatically created by CLI commands with secure permissions.

## Important Constraints
- **100ms JS timeout** - Serverless execution limit
- **100MB file limit** - Per-file in ZIP uploads
- **1MB request body** - API limit (100MB for deploy)
- **5 deploys/minute** - Rate limiting per IP
- **Site isolation** - KV store and env vars scoped by site_id
- **SSRF protection** - Blocks internal IPs in fetch()

## HTTP API Endpoints
**Analytics**: `/track`, `/pixel.gif`, `/r/{slug}`, `/webhook/{name}`
**Dashboard**: `/api/stats`, `/api/events`, `/api/redirects` (auth required)
**PaaS**: `/api/deploy` (Bearer token), `/api/sites`, `/api/keys`, `/api/envvars`
**Auth**: `/login`, `/api/login`, `/api/logout`, `/api/auth/status`

## Request Routing
- **Dashboard** (localhost/main domain) → Authentication required
- **Sites** (subdomain.localhost) → Static files or serverless JS
- **Middleware**: Tracing → Logging → Body limit → Security → CORS → Auth

## Common Tasks
**Add API endpoint**: Handler in `internal/handlers/`, register in `main.go` dashboardMux
**Add JS function**: Edit `internal/hosting/runtime.go` vm.Set() section
**Add table**: Create migration file, update `internal/database/db.go`
**Security**: Validate inputs, use parameterized queries, apply rate limiting

## CLI Interface Design
The CLI uses a clear subcommand structure rather than flag-based modes:
- **Subcommands**: `set-credentials`, `deploy`, `start`, `stop`
- **Flag ordering**: Flags must come before positional arguments (Go flag package standard)
- **Error handling**: Clear error messages and usage guidance
- **Auto-creation**: Config directories created automatically with secure permissions