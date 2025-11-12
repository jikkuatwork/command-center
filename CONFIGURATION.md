# Configuration Guide

## Overview

Command Center v0.2.0 uses a flexible JSON-based configuration system with support for CLI flags, environment variables, and multiple config profiles.

## Configuration Priority

Configuration values are loaded in the following order (highest priority first):

1. **CLI Flags** (highest priority)
2. **JSON Config File**
3. **Environment Variables** (backward compatibility)
4. **Built-in Defaults** (lowest priority)

## Default Locations

- **Config File**: `~/.config/cc/config.json`
- **Database**: `~/.config/cc/data.db`
- **Backups**: `~/.config/cc/backups/`

## Configuration File Format

### Complete Example

```json
{
  "server": {
    "port": "4698",
    "domain": "https://cc.toolbomber.com",
    "env": "production"
  },
  "database": {
    "path": "~/.config/cc/data.db"
  },
  "auth": {
    "enabled": true,
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "ntfy": {
    "topic": "your-topic",
    "url": "https://ntfy.sh"
  }
}
```

### Configuration Fields

#### Server Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `server.port` | string | `"4698"` | Port to listen on |
| `server.domain` | string | `"https://cc.toolbomber.com"` | Public domain for the server |
| `server.env` | string | `"development"` | Environment: `development` or `production` |

#### Database Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `database.path` | string | `"~/.config/cc/data.db"` | Path to SQLite database file |

#### Authentication Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auth.enabled` | boolean | `false` | Enable/disable authentication |
| `auth.username` | string | `""` | Username for login |
| `auth.password_hash` | string | `""` | bcrypt hash of password |

**Note**: Never set `password_hash` manually. Use `--username` and `--password` flags to update credentials.

#### Ntfy Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ntfy.topic` | string | `""` | ntfy.sh topic for notifications |
| `ntfy.url` | string | `"https://ntfy.sh"` | ntfy.sh server URL |

## CLI Flags

### All Available Flags

```bash
cc-server [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--config <path>` | string | Path to config file |
| `--env <environment>` | string | Load environment-specific config |
| `--db <path>` | string | Database file path (overrides config) |
| `--port <port>` | string | Server port (overrides config) |
| `--username <user>` | string | Set/update username |
| `--password <pass>` | string | Set/update password |
| `--version` | flag | Show version and exit |
| `--help, -h` | flag | Show help |
| `--verbose` | flag | Enable verbose logging |
| `--quiet` | flag | Quiet mode (errors only) |

### Examples

#### Basic Usage

```bash
# Start with default config
./cc-server

# Start with custom config file
./cc-server --config /path/to/config.json

# Start on custom port
./cc-server --port 8080

# Start with custom database
./cc-server --db /path/to/data.db
```

#### Authentication Setup

```bash
# Set up authentication (creates/updates config)
./cc-server --username admin --password mysecurepassword

# Then start normally
./cc-server
```

#### Environment-Specific Configs

```bash
# Development (loads config.development.json)
./cc-server --env development

# Production (loads config.production.json)
./cc-server --env production
```

## Environment Variables

For backward compatibility with v0.1.0, the following environment variables are supported:

| Variable | Description | Config Equivalent |
|----------|-------------|-------------------|
| `PORT` | Server port | `server.port` |
| `DB_PATH` | Database path | `database.path` |
| `ENV` | Environment | `server.env` |
| `NTFY_TOPIC` | Ntfy topic | `ntfy.topic` |
| `NTFY_URL` | Ntfy URL | `ntfy.url` |

**Note**: Environment variables have lower priority than config files and CLI flags.

## Environment-Specific Configs

### Development Config

Create `config.development.json`:

```json
{
  "server": {
    "port": "4698",
    "domain": "http://localhost:4698",
    "env": "development"
  },
  "database": {
    "path": "./cc-dev.db"
  },
  "auth": {
    "enabled": false,
    "username": "",
    "password_hash": ""
  },
  "ntfy": {
    "topic": "",
    "url": "https://ntfy.sh"
  }
}
```

**Features**:
- Auth disabled for easy testing
- Local database file
- HTTP (not HTTPS)
- CORS enabled
- Verbose logging

### Production Config

Create `config.production.json`:

```json
{
  "server": {
    "port": "4698",
    "domain": "https://cc.toolbomber.com",
    "env": "production"
  },
  "database": {
    "path": "~/.config/cc/data.db"
  },
  "auth": {
    "enabled": true,
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "ntfy": {
    "topic": "production-alerts",
    "url": "https://ntfy.sh"
  }
}
```

**Features**:
- Auth required
- Secure defaults
- HSTS enabled
- Minimal logging
- Proper database location

## Creating Configs

### Using --username and --password

The easiest way to create a config is:

```bash
./cc-server --username admin --password yourpassword --config ~/.config/cc/config.json
```

This will:
1. Create `~/.config/cc/` directory if needed
2. Generate `config.json` with your credentials
3. Hash password with bcrypt
4. Enable authentication
5. Exit (doesn't start server)

Then start normally:

```bash
./cc-server
```

### Manual Creation

1. Copy `config.example.json`:
   ```bash
   cp config.example.json ~/.config/cc/config.json
   ```

2. Generate password hash:
   ```bash
   ./cc-server --username admin --password temp
   ```

3. Copy the hash from the created config

4. Edit your config file with the hash

## Configuration Validation

The server validates configuration on startup:

- **Port**: Must be 1-65535
- **Environment**: Must be `development` or `production`
- **Database Path**: Must be writable
- **Auth**: If enabled, username and password_hash required

Validation errors prevent server startup with helpful error messages.

## File Permissions

The server automatically sets secure permissions:

- Config file: `0600` (owner read/write only)
- Database: `0600` (owner read/write only)
- Backup directory: `0700` (owner access only)

## Troubleshooting

### Config Not Found

```
Failed to load config: no such file or directory
```

**Solution**: Create config using `--username` and `--password` flags or copy from `config.example.json`.

### Invalid Port

```
Invalid port: must be 1-65535
```

**Solution**: Set valid port in config or use `--port` flag with valid value.

### Permission Denied

```
Failed to write config file: permission denied
```

**Solution**: Ensure `~/.config/cc/` directory exists and you have write permissions.

### Auth Enabled But No Credentials

```
Auth enabled but username is empty
```

**Solution**: Either disable auth or set credentials with `--username` and `--password`.

## Best Practices

1. **Use environment-specific configs** for different deployments
2. **Don't commit `config.json`** to version control
3. **Use secure file permissions** (automatic in v0.2.0)
4. **Regular backups** of config and database
5. **Document custom configs** for team members
6. **Use `--env` flag** instead of manual config switching
7. **Validate configs** with `--config` flag before deploying
8. **Keep example config updated** when adding fields

## Migration from v0.1.0

If you're upgrading from v0.1.0 (environment variable based config):

1. Your existing environment variables will still work
2. Create a config file for better management:
   ```bash
   ./cc-server --username admin --password yourpass
   ```
3. Gradually migrate to config file
4. Remove environment variables when ready

See [UPGRADE.md](UPGRADE.md) for detailed migration guide.
