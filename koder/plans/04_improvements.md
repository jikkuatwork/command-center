# Command Center - Phase 4: Improvements & Hardening

## Overview
- **Current Version**: v0.3.0 (Personal Cloud)
- **Target Version**: v0.3.1
- **Goal**: Fix bugs, improve security, enhance developer experience, expand test coverage
- **Constraint**: Maintain backward compatibility, single binary architecture

---

## Phase 1: Critical Bug Fixes (Commit #1)

**Task**: Fix known bugs that break functionality.

1. **Fix notifier/ntfy.go** - Lines 34, 63 use method syntax instead of struct fields
   - Change `cfg.NtfyTopic` to `cfg.Ntfy.Topic`
   - Change `cfg.NtfyURL` to `cfg.Ntfy.URL`

2. **Fix chat-app example** - main.js intercepts all requests, preventing index.html from loading
   - Add check: if `req.method === 'GET' && req.path === '/'` return false to serve static

**Verification**: `go test ./...` passes, `go build` succeeds.

---

## Phase 2: Developer Experience (Commit #2)

**Task**: Add project documentation for AI assistants and developers.

1. **Create CLAUDE.md** in project root with:
   - Project overview
   - Key file locations
   - Build/test commands
   - Architecture notes

2. **Create Makefile** with common targets:
   - `make build` - Build binary
   - `make test` - Run tests
   - `make run` - Build and run server
   - `make clean` - Clean artifacts

**Verification**: `make test` and `make build` work.

---

## Phase 3: Security Hardening (Commit #3)

**Task**: Address security concerns from review.

1. **Rate limiting for /api/deploy**
   - Add rate limiter (5 deploys per minute per IP)
   - Use existing auth/ratelimit.go pattern

2. **Request body size limit for serverless**
   - Limit POST body to 1MB in runtime.go
   - Return 413 if exceeded

3. **WebSocket authentication** (optional scope)
   - For now, just add site existence check before upgrade

**Verification**: Test rate limiting manually, verify body limit.

---

## Phase 4: Runtime Enhancements (Commit #4)

**Task**: Add fetch() to JavaScript runtime for HTTP requests.

1. **Add fetch() to runtime.go**
   ```javascript
   // Synchronous fetch (returns response object)
   const resp = fetch(url, { method: 'GET', headers: {} });
   // resp.status, resp.body, resp.headers
   ```

2. **Add timeout (5s) and size limit (1MB response)**

3. **Block localhost/internal IPs** (SSRF protection)

**Verification**: Test with example that fetches external API.

---

## Phase 5: CLI Improvements (Commit #5)

**Task**: Add site management commands.

1. **`cc-server sites`** - List all deployed sites
2. **`cc-server sites delete <name>`** - Delete a site
3. **`cc-server logs <site>`** - Tail logs (placeholder for Phase 6)

**Verification**: Test all CLI commands.

---

## Phase 6: Observability (Commit #6)

**Task**: Add logging infrastructure for serverless.

1. **Create `site_logs` table**
   ```sql
   CREATE TABLE site_logs (
     id INTEGER PRIMARY KEY,
     site_id TEXT NOT NULL,
     level TEXT NOT NULL, -- info, warn, error
     message TEXT NOT NULL,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );
   ```

2. **Update runtime.go console.log** to write to database

3. **Add `/api/logs?site_id=X&limit=50` endpoint**

4. **Update CLI `cc-server logs <site>`** to fetch from API

**Verification**: Console.log appears in logs API.

---

## Phase 7: Test Coverage Expansion (Commit #7)

**Task**: Add tests for critical paths.

1. **internal/handlers/** - Test deploy, hosting handlers
2. **internal/auth/** - Test session, password validation
3. **internal/config/** - Test config loading

**Target**: >50% coverage on core packages.

**Verification**: `go test ./... -cover` shows improvement.

---

## Execution Rules

1. **Test First**: Run `go test ./...` before committing
2. **Build Check**: Ensure `go build ./...` succeeds
3. **Incremental**: One phase per commit
4. **No Regressions**: Existing tests must pass

---
