# Command Center - AI Assistant Context

## Project Overview
Command Center is a Personal PaaS (Platform as a Service) combining:
- **Analytics & Tracking**: Pageviews, pixels, redirects, webhooks
- **Static Site Hosting**: Deploy static sites via CLI (Surge-like)
- **Serverless JavaScript**: Run JS functions with `main.js`
- **Key-Value Store**: Persistent data for serverless apps
- **WebSocket Support**: Real-time communication

**Single binary, SQLite-backed, local filesystem storage.**

## Architecture

```
cmd/server/main.go       # Entry point, CLI, HTTP routing
internal/
  hosting/               # PaaS core logic
    manager.go           # Site CRUD operations
    deploy.go            # ZIP extraction, API keys
    runtime.go           # Goja JS runtime
    security.go          # Path traversal protection
    ws.go                # WebSocket hub
  handlers/              # HTTP handlers
    deploy.go            # POST /api/deploy
    hosting.go           # Sites, keys, env vars APIs
  auth/                  # Session management, rate limiting
  database/              # SQLite with migrations
  config/                # JSON config loading
migrations/              # SQL schema files
web/templates/           # HTML templates
examples/                # Sample apps
```

## Key Commands

```bash
# Build
go build -o cc-server ./cmd/server/

# Run tests
go test ./...

# Run server
./cc-server

# Deploy a site
cd my-site && ../cc-server deploy sitename --token <api-key>

# Setup auth
./cc-server --username admin --password secret
```

## Database Schema

- `events` - Analytics events
- `api_keys` - Deploy tokens (bcrypt hashed)
- `kv_store` - Serverless key-value data (site-scoped)
- `env_vars` - Environment variables (site-scoped)
- `deployments` - Deployment history
- `migrations` - Schema version tracking

## JavaScript Runtime API

```javascript
// Request
req.method, req.path, req.query, req.headers, req.body

// Response
res.send(html), res.json(obj), res.status(code), res.header(k, v)

// Storage (site-isolated)
db.get(key), db.set(key, value), db.delete(key)

// Environment
process.env.MY_SECRET

// WebSocket
socket.broadcast(msg), socket.clients()

// Logging
console.log(...)

// HTTP requests (synchronous, 5s timeout, 1MB limit, SSRF protected)
const resp = fetch(url, { method: 'POST', headers: {}, body: '' });
// resp.status, resp.body, resp.headers, resp.error
```

## Configuration

Config file: `~/.config/cc/config.json`
```json
{
  "server": { "port": "4698", "domain": "localhost", "env": "development" },
  "database": { "path": "~/.config/cc/data.db" },
  "auth": { "enabled": false, "username": "", "password_hash": "" }
}
```

## Important Constraints

1. **100ms JS timeout** - Scripts killed after 100ms
2. **100MB file limit** - Per-file limit in ZIP uploads
3. **Site isolation** - KV store and env vars scoped by site_id
4. **Path traversal protection** - ZIP extraction and file serving secured

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/hosting/...
```

## Common Tasks

### Add a new API endpoint
1. Add handler in `internal/handlers/`
2. Register route in `cmd/server/main.go` (dashboardMux)
3. Add auth middleware if needed

### Add new JS runtime function
1. Edit `internal/hosting/runtime.go`
2. Add to `vm.Set()` calls before `vm.RunString()`

### Add database table
1. Create `migrations/00X_name.sql`
2. Add to migrations list in `internal/database/db.go`
