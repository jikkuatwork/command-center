# Cartridge Architecture Strategy: The "Unikernel-Lite" Transformation

**Status:** Planning
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
    -   `TestDeploySite`: Deploy a ZIP, verify files exist.
    -   `TestServeSite`: Request files, verify content/headers.
    -   `TestServerless`: Deploy `main.js`, verify execution.
2.  **E2E Test Script (`test_e2e_hosting.sh`)**:
    -   Spin up a test server (random port).
    -   Use `curl` to create an account/token.
    -   Deploy a static site (HTML + CSS).
    -   Deploy a serverless app (Counter).
    -   `curl` the subdomains and verify output.
    -   Verify `events` table entries.

---

## Phase 2: Schema Migration (VFS & Certs)
**Objective:** Prepare the database to act as our filesystem.

### Tasks
1.  **Create `migrations/004_vfs.sql`**:
    ```sql
    -- The Virtual Filesystem
    CREATE TABLE files (
        site_id TEXT NOT NULL,
        path TEXT NOT NULL,         -- e.g. "/index.html"
        content BLOB,
        size_bytes INTEGER NOT NULL,
        mime_type TEXT,
        hash TEXT NOT NULL,         -- SHA256 for ETag
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (site_id, path)
    );
    CREATE INDEX idx_files_site_path ON files(site_id, path);

    -- CertMagic Storage
    CREATE TABLE certificates (
        key TEXT PRIMARY KEY,       -- CertMagic key
        value BLOB,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    ```
2.  **Update `internal/database/db.go`**: Register new migration.

---

## Phase 3: The Virtual Filesystem (Implementation)
**Objective:** Abstract file operations so they read/write from SQL instead of Disk.

### Tasks
1.  **Create `internal/hosting/vfs.go`**:
    -   `type FileSystem interface { ... }`
    -   `type SQLFileSystem struct { db *sql.DB }`
    -   Implement: `WriteFile(siteID, path, content)`, `ReadFile(siteID, path)`, `DeleteSite(siteID)`.
2.  **Refactor `internal/hosting/deploy.go`**:
    -   Remove `os.Mkdir` / `os.Create`.
    -   Stream ZIP content directly into `files` table.
    -   Calculate SHA256 hash during upload.
3.  **Refactor `internal/hosting/runtime.go`**:
    -   Change `os.ReadFile("main.js")` to `vfs.ReadFile(...)`.

---

## Phase 4: Serving from VFS (The Switch)
**Objective:** Serve HTTP traffic directly from the database.

### Tasks
1.  **Create `VFSHandler` in `internal/hosting/handler.go`**:
    -   Input: `siteID`, `path`.
    -   Logic:
        -   Query `files` table.
        -   If not found -> 404.
        -   Set `Content-Type` (from DB).
        -   Set `ETag` (from DB hash).
        -   Handle `If-None-Match` (return 304).
        -   Stream BLOB to response.
2.  **Update `cmd/server/main.go`**:
    -   Replace `http.FileServer` with `hosting.VFSHandler`.

---

## Phase 5: CertMagic Integration (HTTPS)
**Objective:** Native HTTPS without Nginx.

### Tasks
1.  **Create `internal/certstore/store.go`**:
    -   Implement `certmagic.Storage` interface backed by `certificates` table.
2.  **Update `cmd/server/main.go`**:
    -   Add `--https` flag (default: false for dev).
    -   If enabled:
        -   Initialize `certmagic.Config` with SQL storage.
        -   Use `certmagic.HTTPS(mux, domains...)`.

---

## Phase 6: Cleanup & Polish
**Objective:** Remove legacy code.

### Tasks
1.  **Delete `~/.config/fazt/sites/`**: No longer needed.
2.  **Update `status` command**: Report VFS usage (SQL count/size) instead of disk usage.
3.  **Documentation**: Update architecture diagrams and backup guides.

---

## Test Plan for Transition
1.  Run **Phase 1 Tests** on v0.4.0 (Baseline).
2.  Implement Phase 2 & 3.
3.  Run **Phase 1 Tests** (Should fail or need adjustment for "Deploy" check).
4.  Implement Phase 4.
5.  Run **Phase 1 Tests** (Should pass, proving VFS behaves like Disk).
