# CLAUDE.md - Fazt.sh Development Guide

## Project Overview
**Fazt.sh** is a self-hosted Platform-as-a-Service (PaaS) and Analytics platform.
- **Architecture**: "Cartridge" Model (Single Binary + Single SQLite DB).
- **Core Philosophy**: Zero dependencies, database-as-filesystem (VFS).
- **Language**: Go 1.20+
- **Database**: SQLite (with WAL mode).
- **HTTPS**: Native automatic HTTPS via CertMagic (Let's Encrypt).

## Key Commands

### Build & Run
- **Build**: `go build -o fazt ./cmd/server`
- **Test (All)**: `go test ./...`
- **Test (E2E)**: `./test_e2e_hosting.sh` (Requires built binary)
- **Run (Dev)**: `go run ./cmd/server server start`

### Code Quality
- **Format**: `go fmt ./...`
- **Lint**: `golangci-lint run` (if available)
- **Vet**: `go vet ./...`

## Project Structure
- `cmd/server/`: Main entry point (CLI & Server).
- `internal/`: Core logic.
  - `hosting/`: VFS, Deploy, Runtime (JS), WebSocket.
  - `database/`: DB connection, CertMagic storage, Migrations runner.
  - `config/`: Configuration structs and loading.
  - `handlers/`: HTTP API and Dashboard handlers.
  - `auth/`: Authentication and Session management.
- `migrations/`: SQL schema files (auto-applied on startup).
- `web/`: Embedded static assets and templates.

## Architecture Standards
1.  **VFS First**: All site content MUST be stored in the `files` table. No disk I/O for user content.
2.  **Single DB**: All state (files, certs, logs, users) must live in `data.db`.
3.  **Config**: Prioritize `config.json` > CLI Flags > Env Vars.
4.  **Security**: Auth is required by default. HTTPS is available via config.

## Deployment Strategy
1.  **Update**: Replace the `fazt` binary.
2.  **Backup**: Copy `data.db`.
3.  **Certificates**: Stored in `certificates` table (managed by CertMagic).

## Pending Features
- **Auto-Provisioning**: `fazt server install` command (See `koder/plans/08_auto_install.md`).
