# Architecture Specification: The "Fazt" Monolithic PaaS

**Target Audience**: LLM Agent (Go Implementation Specialist)
**Goal**: Build a self-contained, single-binary PaaS using Go and SQLite.
**Philosophy**: "The Database is the Filesystem."

## 1. Executive Summary

**Fazt** is a personal Platform-as-a-Service (PaaS) designed as a single executable. It manages multiple static websites, handles automatic HTTPS, and serves traffic—all from a single Go binary backed by a single SQLite database.

**Core Constraints:**
1.  **Zero External Dependencies**: No Nginx, No Caddy, No Redis, No Filesystem storage.
2.  **Single Artifact State**: The entire state of the world (Users, Sites, Files, Logs, SSL Certificates) must live in `fazt.db`.
3.  **Low Resource Footprint**: Optimized for $5/mo VPS (1GB RAM, 1 vCPU).

---

## 2. Architecture Overview

### 2.1 The "Unikernel-Lite" Model
Instead of a container, the application behaves like a dedicated OS process that handles all IO.

```
[ Internet ]
      |
      v
[ Fazt Binary (Go) ] <------> [ fazt.db (SQLite) ]
      |
      +--- HTTP/HTTPS Listener (Net/HTTP + CertMagic)
      +--- Virtual Filesystem (SQL Blob Store)
      +--- Email Service (SMTP Relay/Receiver)
      +--- Admin API/UI (Embedded)
```

### 2.2 Component Stack
*   **Language**: Go 1.22+
*   **Database**: SQLite3 (with `mattn/go-sqlite3` via CGO or `modernc.org/sqlite` pure Go).
*   **HTTPS**: `caddyserver/certmagic` (Library).
*   **Replication Readiness**: `Litestream` (External sidecar, optimized for WAL mode).

---

## 3. Detailed Design

### 3.1 Database Schema (The "Filesystem")

Since we are replacing the filesystem, the schema is critical.

**Key Tables:**

1.  **`sites`**:
    *   `id` (UUID), `domain` (Text, Unique), `user_id`, `created_at`.
    *   *Purpose*: Maps incoming `Host` header to a specific site bucket.

2.  **`files`**:
    *   `site_id` (FK), `path` (Text, e.g., "/index.html"), `content` (BLOB), `mime_type` (Text), `size` (Int), `hash` (Text, for ETag), `updated_at`.
    *   *Index*: `(site_id, path)` for O(1) file lookups.
    *   *Rationale*: Serving files from BLOBs is faster than small file IO on cheap VPS disks.

3.  **`certificates`** (CertMagic Storage):
    *   `key` (Text, Primary Key), `value` (BLOB), `updated_at`.
    *   *Purpose*: Stores Let's Encrypt keys and certs. Replaces `~/.local/share/caddy`.

4.  **`analytics`**:
    *   `site_id` (FK), `path`, `ip_hash`, `user_agent`, `timestamp`.
    *   *Optimization*: Use Write-Ahead Logging (WAL) to prevent locking during writes.

### 3.2 HTTPS & CertMagic Integration

We replace Caddy with its internal engine, `certmagic`.

*   **Logic**:
    1.  Initialize `certmagic.Config` with `OnDemand: true` (Issues certs when a new domain hits the server).
    2.  **CRITICAL**: Implement `certmagic.Storage` interface backed by the `certificates` SQLite table.
        *   `Store(key, value)` -> `INSERT/REPLACE INTO certificates...`
        *   `Load(key)` -> `SELECT value FROM certificates...`
        *   `Delete(key)` -> `DELETE FROM certificates...`
        *   `List(prefix)` -> `SELECT key FROM certificates WHERE key LIKE prefix%`
        *   `Stat(key)` -> Metadata query.
        *   `Lock/Unlock` -> Use SQLite `transactions` or a dedicated `locks` table to prevent cluster race conditions (though we are single-node, CertMagic requires it).
    3.  Wrap the standard `http.ServeMux` with `certmagic.HTTPS()`.

### 3.3 The Virtual Filesystem (VFS) Router

The HTTP Handler logic:

1.  **Host Matching**: Extract `r.Host` (e.g., `blog.myapp.com`). Query `sites` table.
2.  **Path Matching**: Extract `r.URL.Path` (e.g., `/style.css`).
3.  **File Query**: `SELECT content, mime_type, hash FROM files WHERE site_id=? AND path=?`.
4.  **Response**:
    *   If found: Set `Content-Type`, `ETag` (from hash). Stream BLOB to `w`.
    *   If not found: Serve `404.html` from DB or generic 404.
5.  **Caching**: Use `ETag` matching. If `If-None-Match == hash`, return `304 Not Modified` (saves bandwidth/DB reads).

### 3.4 Email Architecture

*   **Sending (Tx)**:
    *   Do **NOT** implement SMTP protocol directly.
    *   Use an API client (e.g., Postmark/Resend/AWS SES SDK).
    *   Store `outbox` in SQLite for reliability (Worker process reads `outbox` -> sends via API -> marks `sent`).
*   **Receiving (Rx)**:
    *   Listen on Port 25 (if possible) or utilize an Inbound Parse Webhook (e.g., Postmark/SendGrid posts JSON to `https://paas.com/api/hooks/email`).
    *   Store received emails in `inbox` table.

---

## 4. "Litestream Ready" Requirements

Litestream replicates SQLite databases to S3 by reading the Write-Ahead Log (WAL).

**Application Requirements:**
1.  **WAL Mode**: The application MUST run `PRAGMA journal_mode = WAL;` on startup.
2.  **Busy Timeout**: Set `PRAGMA busy_timeout = 5000;` (5s) to handle replication locks gracefully.
3.  **No Checkpoints**: Let Litestream manage checkpoints if possible, or ensure `PRAGMA wal_autocheckpoint` isn't too aggressive.
4.  **File Stability**: The `fazt.db` path must not change.

**Architecture Fit**: Litestream runs as a separate binary (sidecar systemd service). It does not touch the Go code. The Go app just needs to be "polite" with the DB file (WAL mode).

---

## 5. Development Workflow (The "Cartridge" Experience)

1.  **The Artifact**: A single file `fazt.db` contains your entire world.
2.  **Backup/Analyze**: `ssh user@host "sqlite3 fazt.db '.backup download.db'"` (Safe online backup).
3.  **Restore/Swap**:
    *   Upload `local.db` -> `fazt.db`.
    *   Restart service.
    *   *Note*: This resets Analytics/Users to the snapshot time. Ideal for code/config changes, risky for high-volume user data (use Migration scripts for that).

---

## 6. Implementation Roadmap

1.  **Phase 1: The Core**: Go binary + SQLite connection + WAL mode setup.
2.  **Phase 2: SQL Storage**: Implement `certmagic.Storage` interface.
3.  **Phase 3: VFS**: Implement the File Upload API and VFS Router (Serving BLOBs).
4.  **Phase 4: Multi-tenancy**: Host-based routing logic.
5.  **Phase 5: Email**: API integration.

## 7. Rationale

*   **Why CertMagic?**: Removes the need for Caddy/Nginx sidecars. Go handles TLS 1.3/ACME natively.
*   **Why SQLite BLOBs?**: Removes "File Sync" complexity. Backing up the DB backs up the images.
*   **Why Single Binary?**: "SCP to Deploy". No Docker, no `apt-get`, no dependencies.

This architecture achieves the "Unbelievable Achievement": A robust, SSL-secured, multi-tenant Cloud Platform in a single 25MB file + one Database.
