-- Migration 004: Virtual Filesystem and CertMagic Storage

-- 1. Files table (The Virtual Filesystem)
CREATE TABLE files (
    site_id TEXT NOT NULL,
    path TEXT NOT NULL,         -- e.g. "index.html" or "css/style.css"
    content BLOB,
    size_bytes INTEGER NOT NULL,
    mime_type TEXT,
    hash TEXT NOT NULL,         -- SHA256 hash for ETag/Deduplication
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (site_id, path)
);

-- Index for fast file lookups during request serving
CREATE INDEX idx_files_lookup ON files(site_id, path);

-- 2. Certificates table (CertMagic Storage)
CREATE TABLE certificates (
    key TEXT PRIMARY KEY,       -- CertMagic storage key
    value BLOB,                 -- Certificate data (PEM, JSON, etc)
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. Site Logs (Observability for Serverless)
CREATE TABLE site_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL,
    level TEXT NOT NULL,        -- 'info', 'warn', 'error'
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_site_logs_site ON site_logs(site_id, created_at DESC);
