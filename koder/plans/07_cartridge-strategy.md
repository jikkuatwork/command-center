# Cartridge Architecture Strategy: The "Unikernel-Lite" Transformation

**Status:** Completed ✅
**Target Version:** v0.5.0
**Goal:** Transform `fazt` into a stateless single-binary application where all persistent state (files, certs, data) lives in `fazt.db`.

## Philosophy
- **The Binary is the Computer:** Updating the OS means replacing the binary.
- **The Database is the Filesystem:** All user content lives in SQLite.
- **Zero Dependencies:** No external Nginx, no `~/.local/share` files.

---

## Phase 1: The Safety Net (Test Suite)
**Objective:** Build a robust test suite to guarantee no regressions when we swap the storage engine.

### Tasks
1.  **Integration Tests (`internal/hosting`)**:
    -   [x] `TestDeploySite`: Deploy a ZIP, verify files exist.
    -   [x] `TestServeSite`: Request files, verify content/headers.
    -   [x] `TestServerless`: Deploy `main.js`, verify execution.
2.  **E2E Test Script (`test_e2e_hosting.sh`)**:
    -   [x] Spin up a test server (random port).
    -   [x] Use `curl` to create an account/token.
    -   [x] Deploy a static site (HTML + CSS).
    -   [x] Deploy a serverless app (Counter).
    -   [x] `curl` the subdomains and verify output.
    -   [x] Verify `events` table entries.

---

## Phase 2: Schema Migration (VFS & Certs)
**Objective:** Prepare the database to act as our filesystem.

### Tasks
1.  **Create `migrations/004_vfs.sql`**:
    - [x] Create `files` table (VFS).
    - [x] Create `certificates` table (CertMagic).
2.  **Update `internal/database/db.go`**:
    - [x] Register new migration.
    - [x] Embed migrations via `embed.FS`.

---

## Phase 3: The Virtual Filesystem (Implementation)
**Objective:** Abstract file operations so they read/write from SQL instead of Disk.

### Tasks
1.  **Create `internal/hosting/vfs.go`**:
    -   [x] `type FileSystem interface { ... }`
    -   [x] `type SQLFileSystem struct { db *sql.DB }`
    -   [x] Implement: `WriteFile`, `ReadFile`, `DeleteSite`.
2.  **Refactor `internal/hosting/deploy.go`**:
    -   [x] Remove `os.Mkdir` / `os.Create`.
    -   [x] Stream ZIP content directly into `files` table.
3.  **Refactor `internal/hosting/runtime.go`**:
    -   [x] Change `os.ReadFile("main.js")` to `vfs.ReadFile(...)`.

---

## Phase 4: Serving from VFS (The Switch)
**Objective:** Serve HTTP traffic directly from the database.

### Tasks
1.  **Create `VFSHandler` in `internal/hosting/handler.go`**:
    -   [x] Logic: Query `files` -> Set Content-Type/ETag -> Stream BLOB.
2.  **Update `cmd/server/main.go`**:
    -   [x] Replace `http.FileServer` with `hosting.VFSHandler`.

---

## Phase 5: CertMagic Integration (HTTPS)
**Objective:** Native HTTPS without Nginx.

### Tasks
1.  **Create `internal/certstore/store.go`**:
    -   [x] Implement `certmagic.Storage` interface backed by `certificates` table.
2.  **Update `cmd/server/main.go`**:
    -   [x] Add `--https` flag.
    -   [x] Use `certmagic.HTTPS(mux, domains...)` with SQL storage.

---

## Phase 6: Cleanup & Polish
**Objective:** Remove legacy code and ensure production readiness.

### Tasks
1.  **Delete `~/.config/fazt/sites/`**: No longer needed.
2.  **Update `status` command**: Report VFS usage.
3.  **Documentation**: Updated diagrams and guides.
4.  **Static Binary**: Switched to `modernc.org/sqlite` (Pure Go) for CGO-free builds.
5.  **Embedded Assets**: Web templates and static files are now inside the binary.

---

## Status Update (v0.5.0)
The Cartridge Architecture is fully implemented. The application is now a single static binary that contains the entire runtime, database schema, and UI assets. It can be deployed to any Linux server with a single command.