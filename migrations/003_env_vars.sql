-- Environment variables for serverless apps
CREATE TABLE IF NOT EXISTS env_vars (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(site_id, name)
);

CREATE INDEX IF NOT EXISTS idx_env_vars_site_id ON env_vars(site_id);
