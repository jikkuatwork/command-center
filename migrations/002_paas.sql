-- Command Center - PaaS Migration (Personal Cloud)

-- API Keys for deploying sites (for friends/CLI)
CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL, -- e.g. "Friend Bob"
    key_hash TEXT NOT NULL, -- bcrypt hash of the token
    scopes TEXT, -- JSON array e.g. ["deploy:blog", "read:stats"]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- Key-Value store for Serverless Apps
CREATE TABLE IF NOT EXISTS kv_store (
    site_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT, -- JSON or text
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (site_id, key)
);

CREATE INDEX IF NOT EXISTS idx_kv_store_site_id ON kv_store(site_id);

-- Track deployments
CREATE TABLE IF NOT EXISTS deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL,
    size_bytes INTEGER,
    file_count INTEGER,
    deployed_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deployments_site_id ON deployments(site_id);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at);
