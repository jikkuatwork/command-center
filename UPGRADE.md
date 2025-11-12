# Upgrade Guide: v0.1.0 → v0.2.0

## Overview

Command Center v0.2.0 introduces authentication and JSON-based configuration. This guide helps you upgrade from v0.1.0 safely.

## What's New in v0.2.0

### Major Changes

✅ **JSON Configuration System**
- Replace environment variables with `config.json`
- CLI flags for easy management
- Environment-specific configs

✅ **Authentication & Security**
- Username/password with bcrypt hashing
- Session management with secure cookies
- Rate limiting against brute-force attacks
- Audit logging for security events

✅ **Enhanced CLI**
- `--version`, `--help`, `--verbose`, `--quiet` flags
- Better startup messages
- Improved error handling

✅ **Database Migrations**
- Migration tracking system
- Automatic backups (keeps last 5)
- Safe schema updates

### Breaking Changes

⚠️ **Configuration Method**
- Environment variables still work but deprecated
- New default config location: `~/.config/cc/config.json`
- New default database location: `~/.config/cc/data.db`

⚠️ **Dashboard Access**
- Dashboard now protected by authentication (when enabled)
- Tracking endpoints remain public

## Quick Upgrade Path

### For Simple Deployments

If you're running v0.1.0 with default settings:

```bash
# 1. Stop the server
systemctl stop command-center  # or kill the process

# 2. Backup your database
cp cc.db cc.db.backup

# 3. Download v0.2.0
wget https://github.com/jikkuatwork/command-center/releases/download/v0.2.0/command-center-v0.2.0-linux-amd64.tar.gz
tar -xzf command-center-v0.2.0-linux-amd64.tar.gz

# 4. Set up authentication (optional but recommended)
./cc-server --username admin --password your-secure-password

# 5. Start the server
./cc-server
```

### For Production Deployments

See detailed steps below.

## Detailed Upgrade Steps

### Step 1: Backup Everything

```bash
# Backup database
cp cc.db cc.db.backup.$(date +%Y%m%d)

# Backup environment file (if using)
cp .env .env.backup

# Note your current environment variables
env | grep -E '(PORT|DB_PATH|NTFY)' > env_backup.txt
```

### Step 2: Stop the Current Server

```bash
# If using systemd
sudo systemctl stop command-center

# Or find and kill the process
pkill -f cc-server

# Verify it's stopped
ps aux | grep cc-server
```

### Step 3: Download v0.2.0

```bash
# Download release
wget https://github.com/jikkuatwork/command-center/releases/download/v0.2.0/command-center-v0.2.0-linux-amd64.tar.gz

# Extract
tar -xzf command-center-v0.2.0-linux-amd64.tar.gz

# Make executable
chmod +x cc-server

# Verify version
./cc-server --version
```

### Step 4: Migrate Configuration

#### Option A: Keep Using Environment Variables (Quick)

Your existing `.env` file or environment variables will still work:

```bash
# Just start the server - it will use existing env vars
./cc-server
```

**Note**: While this works, we recommend migrating to JSON config for better management.

#### Option B: Migrate to JSON Config (Recommended)

Create a config file based on your current environment variables:

```bash
# Create config directory
mkdir -p ~/.config/cc

# Option 1: Let the server create it with auth
./cc-server --username admin --password your-password

# Option 2: Create manually from example
cp config.example.json ~/.config/cc/config.json
# Then edit ~/.config/cc/config.json
```

**Migrate Your Settings**:

If you had these environment variables:
```bash
PORT=4698
DB_PATH=./cc.db
NTFY_TOPIC=my-topic
NTFY_URL=https://ntfy.sh
```

Your `config.json` should be:
```json
{
  "server": {
    "port": "4698",
    "domain": "https://cc.toolbomber.com",
    "env": "production"
  },
  "database": {
    "path": "./cc.db"
  },
  "auth": {
    "enabled": true,
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "ntfy": {
    "topic": "my-topic",
    "url": "https://ntfy.sh"
  }
}
```

### Step 5: Move Database (Optional)

The new default location is `~/.config/cc/data.db`. To use the new location:

```bash
# Create directory
mkdir -p ~/.config/cc

# Copy database
cp cc.db ~/.config/cc/data.db

# Update config to use new path
# (or use --db flag)
```

**Or keep using the old location** by setting in config:
```json
{
  "database": {
    "path": "./cc.db"
  }
}
```

### Step 6: Set Up Authentication

#### Enable Authentication (Recommended for Production)

```bash
./cc-server --username admin --password your-secure-password
```

This creates/updates the config with:
- Username in plain text
- Password hashed with bcrypt
- Auth enabled

#### Keep Authentication Disabled (Development Only)

Set in config:
```json
{
  "auth": {
    "enabled": false
  }
}
```

⚠️ **Warning**: Disabling auth in production makes your dashboard publicly accessible!

### Step 7: Update systemd Service (If Using)

If you're using systemd, update your service file:

**Old (v0.1.0)**:
```ini
[Unit]
Description=Command Center
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/command-center
EnvironmentFile=/opt/command-center/.env
ExecStart=/opt/command-center/cc-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**New (v0.2.0)**:
```ini
[Unit]
Description=Command Center v0.2.0
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/command-center
ExecStart=/opt/command-center/cc-server --config /home/www-data/.config/cc/config.json
Restart=on-failure

# Optional: Set resource limits
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

Reload and restart:
```bash
sudo systemctl daemon-reload
sudo systemctl start command-center
sudo systemctl status command-center
```

### Step 8: Verify Migration

```bash
# Check server starts
./cc-server

# Test health endpoint
curl http://localhost:4698/health

# Test tracking (should still work without auth)
curl -X POST http://localhost:4698/track \
  -H "Content-Type: application/json" \
  -d '{"h":"test.com","p":"/","e":"pageview"}'

# Test login (if auth enabled)
curl -X POST http://localhost:4698/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

# Test dashboard (should require auth if enabled)
curl http://localhost:4698/
```

### Step 9: Update Reverse Proxy (If Using)

If you're using nginx or similar, you may want to update headers:

```nginx
location / {
    proxy_pass http://localhost:4698;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # New: For proper IP detection in rate limiting
    proxy_set_header X-Forwarded-Host $server_name;
}
```

## Rollback Procedure

If you need to rollback to v0.1.0:

```bash
# 1. Stop v0.2.0
./cc-server stop  # or kill process

# 2. Restore old binary
mv cc-server cc-server-v0.2.0
mv cc-server-old cc-server  # your v0.1.0 backup

# 3. Restore database backup
cp cc.db.backup cc.db

# 4. Restore environment variables
# Load from .env or env_backup.txt

# 5. Start v0.1.0
./cc-server
```

## Migration Checklist

Use this checklist to ensure smooth migration:

- [ ] Backup database
- [ ] Backup environment variables/config
- [ ] Download v0.2.0 binary
- [ ] Stop v0.1.0 server
- [ ] Create JSON config OR plan to use env vars
- [ ] Set up authentication (recommended)
- [ ] Move database to new location (optional)
- [ ] Update systemd service file (if using)
- [ ] Test server startup
- [ ] Test tracking endpoints (should be public)
- [ ] Test dashboard access (should require auth if enabled)
- [ ] Test login/logout (if auth enabled)
- [ ] Update monitoring/alerting scripts
- [ ] Update documentation for team
- [ ] Remove old backups after successful migration

## Troubleshooting

### Server Won't Start

**Check logs**:
```bash
./cc-server --verbose
```

**Common issues**:
- Config file syntax error → Validate JSON
- Port already in use → Change port or stop other service
- Permission denied → Check file permissions
- Database locked → Ensure old server is stopped

### Can't Access Dashboard

**With auth enabled**:
- Expected behavior! Navigate to `/login`
- Login with credentials set via `--username` and `--password`

**Without auth**:
- Should redirect to dashboard
- If not, check `auth.enabled` in config

### Tracking Not Working

**Check if public endpoints are accessible**:
```bash
# This should work without auth
curl -X POST http://localhost:4698/track \
  -H "Content-Type: application/json" \
  -d '{"h":"test.com","p":"/","e":"pageview"}'
```

If this fails:
- Server might not be running
- Wrong port
- Firewall blocking

### Rate Limited

**If you're locked out after failed login attempts**:
- Wait 15 minutes for automatic reset
- OR restart the server (in-memory rate limiting)

## Getting Help

If you encounter issues during migration:

1. Check the [Configuration Guide](CONFIGURATION.md)
2. Check the [Security Guide](SECURITY.md)
3. Review server logs with `--verbose` flag
4. Open an issue on GitHub with:
   - v0.1.0 configuration
   - Steps you followed
   - Error messages
   - Server logs

## Post-Migration

After successful migration:

1. **Monitor for a few days** to ensure stability
2. **Update your documentation** with new config locations
3. **Train team members** on new authentication
4. **Set up regular backups** of config and database
5. **Review audit logs** periodically
6. **Plan password rotation** schedule
7. **Remove old backups** after confidence period

## Benefits of Upgrading

After migrating to v0.2.0, you'll have:

✨ **Better Security**
- Protected dashboard access
- Audit trail of all logins
- Rate limiting against attacks
- Secure file permissions

✨ **Easier Management**
- Single config file instead of scattered env vars
- Easy credential updates
- Environment-specific configs
- Better error messages

✨ **More Professional**
- Proper CLI with --help and --version
- Clean startup messages
- Comprehensive documentation
- Production-ready defaults
