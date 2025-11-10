# Command Center - Development Start Guide

## Context & Mission

You are building **Command Center (CC)** - a unified analytics, monitoring, and tracking platform. This is a complete v0.1 build meant to be done in one autonomous session while the developer sleeps.

**Critical**: DO NOT STOP or ask for permission. You have blanket freedom to make decisions and complete the build. The developer will return in ~8 hours expecting a working system.

## Key Information

- **Project Location**: Root of repository (NOT in koder/)
- **Meta Folder**: `koder/` contains plans and docs - reference only
- **Build Plan**: `koder/plans/01_initial-build.md` - Your detailed roadmap (23 phases)
- **Deployment Manual**: `koder/docs/meta/01_manual.md` - For post-build deployment
- **Port**: 4698
- **Domain**: https://cc.toolbomber.com
- **Target**: x64 Linux binary (cross-compile from sandbox)

## Project Specifications

**Stack**:
- Backend: Go + SQLite (WAL mode)
- Frontend: Tabler (vanilla JS, no bundling)
- Charts: Chart.js
- PWA: manifest.json + service worker

**Core Features**:
1. Universal tracking endpoint (auto-detect domain from referer)
2. Tracking pixel (1x1 GIF)
3. Redirect service with click tracking
4. Webhook receiver
5. Dashboard with analytics, filters, graphs
6. ntfy.sh integration for alerts
7. PWA support for mobile use

**Frontend Requirements**:
- Clean, minimal aesthetics
- Light/dark mode + themeable
- Responsive (mobile PWA + 14" MacBook Air)
- Vivid graphs
- String truncation to prevent layout breaks
- No bundlers - plain HTML/CSS/JS

**Query String Support**:
- `domain`: Override auto-detected domain
- `tags`: Comma-separated tags (e.g., `app,campaign-3846`)

## Execution Strategy

### Phase Tracking
The plan in `koder/plans/01_initial-build.md` has checkboxes. **Update the plan file as you complete tasks** - this serves as your progress tracker and allows resumption after crashes.

### Commit Strategy
Commit after EVERY phase (23 commits total). Format: `feat/docs/perf: description` as specified in plan.

### Decision Making
- **Don't ask, just build**: Make sensible defaults for anything unspecified
- **Mock external services**: ntfy.sh calls should log instead of actual HTTP during development
- **Test locally**: Use curl/browser at localhost:4698 to verify each component
- **Handle errors gracefully**: Add proper error handling, don't let nil pointers crash

### Testing Approach
- Generate mock data immediately after DB setup (Phase 1)
- Create test scripts (bash/curl) for each endpoint
- Test in browser as you build frontend
- Verify mobile responsive at common breakpoints

### If You Get Stuck
1. Check the detailed plan - it has granular steps
2. Make a reasonable decision and move forward
3. Document the decision in code comments
4. Continue - don't wait for human input

## Quick Start Commands

```bash
# Initialize project
mkdir -p command-center
cd command-center
go mod init github.com/jikku/command-center

# Project structure (see Phase 0 in plan)
mkdir -p cmd/server internal/{config,database,handlers,models,notifier} web/{static/{css,js,img},templates} migrations

# After phases complete, test server
go run cmd/server/main.go
# Should start on :4698

# Test tracking
curl -X POST http://localhost:4698/track \
  -H "Content-Type: application/json" \
  -d '{"h":"test.com","p":"/","e":"pageview","t":["app"]}'

# View dashboard
open http://localhost:4698
```

## Critical Reminders

1. **WAL Mode**: Enable in SQLite connection: `PRAGMA journal_mode=WAL`
2. **CORS**: Add middleware for development (allow all origins)
3. **Port 4698**: Hardcode or make configurable via env
4. **Domain Auto-Detection**: Parse `Referer` header, fallback to query param
5. **Tags**: Store as JSON array or comma-separated string in DB
6. **Tabler**: Use CDN or download dist files - no npm/build process
7. **PWA Icons**: Create simple placeholder SVGs initially
8. **Binary Build**: `GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o cc-server ./cmd/server`

## Resume After Crash

If the session crashes mid-build:

1. Check `koder/plans/01_initial-build.md` for last completed phase
2. Review git log for last commit
3. Test what's working: `go run cmd/server/main.go`
4. Continue from next unchecked phase
5. Don't rebuild completed phases

## Success Criteria

By end of session:
- [ ] All 23 phases complete with commits
- [ ] Server runs on :4698
- [ ] Dashboard loads and displays data
- [ ] Tracking endpoints work (test with curl)
- [ ] Frontend is responsive
- [ ] PWA installable
- [ ] Binary builds for Linux x64
- [ ] README.md has deployment instructions

## Start Now

1. Read `koder/plans/01_initial-build.md` carefully
2. Begin Phase 0 (scaffolding)
3. Work through all 23 phases sequentially
4. Update plan checkboxes as you go
5. Commit after each phase
6. Test continuously
7. Don't stop until complete

**GO BUILD COMMAND CENTER. The developer is counting on you.** 🚀
