# Configuration Guide

## Overview

Command Center v0.3.0 uses a flexible JSON-based configuration system with support for CLI flags, environment variables, and a clean subcommand-based CLI interface.

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

**Note**: Never set `password_hash` manually. Use `set-credentials` subcommand to update credentials.

#### Ntfy Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ntfy.topic` | string | `""` | ntfy.sh topic for notifications |
| `ntfy.url` | string | `"https://ntfy.sh"` | ntfy.sh server URL |

#### API Key Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `api_key.token` | string | `""` | Generated API token for deployments |
| `api_key.name` | string | `""` | Name/description of the API key |

## CLI Commands

Command Center v0.3.0 uses a subcommand-based interface:

```bash
cc-server <command> [flags] [arguments]
```

### Available Commands

| Command | Description |
|---------|-------------|
| `set-credentials` | Set up authentication credentials |
| `start` | Start the Command Center server |
| `stop` | Stop a running Command Center server |
| `deploy` | Deploy a directory to a site |
| `--help, -h` | Show help |
| `--version` | Show version and exit |

### Command-Specific Flags

#### start command
| Flag | Type | Description |
|------|------|-------------|
| `--config <path>` | string | Path to config file |
| `--db <path>` | string | Database file path (overrides config) |
| `--port <port>` | string | Server port (overrides config) |

#### deploy command
| Flag | Type | Description |
|------|------|-------------|
| `--path <directory>` | string | Directory to deploy (required) |
| `--domain <subdomain>` | string | Domain/subdomain for the site (required) |
| `--server <url>` | string | Command Center server URL |

#### set-credentials command
| Flag | Type | Description |
|------|------|-------------|
| `--username <user>` | string | Username for authentication |
| `--password <pass>` | string | Password for authentication |

### Examples

#### Basic Usage

```bash
# Set up authentication (recommended first step)
./cc-server set-credentials --username admin --password secret123

# Start the server
./cc-server start

# Start with custom config file
./cc-server start --config /path/to/config.json

# Start on custom port
./cc-server start --port 8080

# Start with custom database
./cc-server start --db /path/to/data.db

# Stop the server
./cc-server stop

# Deploy a site
./cc-server deploy --path ./my-site --domain my-app

# Deploy to remote server
./cc-server deploy --path ./build --domain app --server https://cc.example.com
```

#### Getting Help

```bash
# General help
./cc-server --help

# Command-specific help
./cc-server start --help
./cc-server deploy --help
./cc-server set-credentials --help
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

### Using set-credentials Command

The easiest way to create a config is:

```bash
./cc-server set-credentials --username admin --password secret123
```

This will:
1. Create `~/.config/cc/` directory if needed (with secure 0700 permissions)
2. Generate `config.json` with your credentials
3. Hash password with bcrypt (cost factor 12)
4. Enable authentication
5. Exit (doesn't start server)

Then start the server:

```bash
./cc-server start
```

### Manual Creation

1. Copy `config.example.json`:
   ```bash
   cp config.example.json ~/.config/cc/config.json
   ```

2. Generate password hash:
   ```bash
   ./cc-server set-credentials --username admin --password temp
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

**Solution**: Create config using `set-credentials` command or copy from `config.example.json`.

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

**Solution**: Either disable auth or set credentials with `set-credentials` command.

## Best Practices

1. **Use environment-specific configs** for different deployments
2. **Don't commit `config.json`** to version control
3. **Use secure file permissions** (automatic in v0.2.0)
4. **Regular backups** of config and database
5. **Document custom configs** for team members
6. **Use subcommands** instead of flag-based operations
7. **Validate configs** before deploying
8. **Keep example config updated** when adding fields
9. **Use the new CLI** for cleaner, more predictable operations

## Migration from v0.2.x to v0.3.0

The CLI interface has changed from flag-based to subcommand-based:

### Old CLI (v0.2.x)
```bash
# Set credentials
./cc-server --username admin --password secret123

# Start server
./cc-server

# Deploy (explicit flags)
./cc-server deploy --path . --domain my-site
```

### New CLI (v0.3.0)
```bash
# Set credentials
./cc-server set-credentials --username admin --password secret123

# Start server
./cc-server start

# Deploy (explicit flags)
./cc-server deploy --path . --domain my-site
```

### Key Changes
1. **Subcommand structure** - Clear separation of concerns
2. **Flag-based arguments** - All parameters use explicit flags (no positional arguments)
3. **Removed `--env` flag** - Environment configs not supported
4. **Consolidated config** - All settings in `~/.config/cc/config.json`
5. **Better error handling** - Clear help and error messages

### Migration Steps
1. Continue using existing config files (compatible)
2. Update deployment scripts to use new `deploy` command
3. Update service files to use `start` command
4. Update documentation with new CLI examples
