# Configuration Guide

## Overview

fazt.sh v0.3.0 uses a flexible JSON-based configuration system with support for CLI flags, environment variables, and a clean subcommand-based CLI interface.

## Configuration Priority

Configuration values are loaded in the following order (highest priority first):

1. **CLI Flags** (highest priority)
2. **JSON Config File**
3. **Environment Variables** (backward compatibility)
4. **Built-in Defaults** (lowest priority)

## Default Locations

- **Config File**: `~/.config/fazt/config.json`
- **Database**: `~/.config/fazt/data.db`
- **Backups**: `~/.config/fazt/backups/`

## Configuration File Format

### Complete Example

```json
{
  "server": {
    "port": "4698",
    "domain": "https://fazt.sh",
    "env": "production"
  },
  "database": {
    "path": "~/.config/fazt/data.db"
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
| `server.domain` | string | `"https://fazt.sh"` | Public domain for the server |
| `server.env` | string | `"development"` | Environment: `development` or `production` |

#### Database Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `database.path` | string | `"~/.config/fazt/data.db"` | Path to SQLite database file |

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

fazt.sh v0.3.0 uses a subcommand-based interface:

```bash
fazt <command> [flags] [arguments]
```

### Available Commands

| Command | Description |
|---------|-------------|
| `server` | Server management commands |
| `client` | Client/deployment commands |
| `--help, -h` | Show help |
| `--version` | Show version and exit |

### Server Commands

| Command | Description |
|---------|-------------|
| `set-credentials` | Set up authentication credentials |
| `start` | Start the fazt.sh server |
| `stop` | Stop a running fazt.sh server |

### Client Commands

| Command | Description |
|---------|-------------|
| `set-auth-token` | Set authentication token for deployments |
| `deploy` | Deploy a directory to a site |

### Command-Specific Flags

#### server start command
| Flag | Type | Description |
|------|------|-------------|
| `--config <path>` | string | Path to config file |
| `--db <path>` | string | Database file path (overrides config) |
| `--port <port>` | string | Server port (overrides config) |

#### client deploy command
| Flag | Type | Description |
|------|------|-------------|
| `--path <directory>` | string | Directory to deploy (required) |
| `--domain <subdomain>` | string | Domain/subdomain for the site (required) |
| `--server <url>` | string | fazt.sh server URL |

#### server set-credentials command
| Flag | Type | Description |
|------|------|-------------|
| `--username <user>` | string | Username for authentication |
| `--password <pass>` | string | Password for authentication |

#### client set-auth-token command
| Flag | Type | Description |
|------|------|-------------|
| `--token <TOKEN>` | string | Authentication token (required) |

### Examples

#### Basic Usage

```bash
# Set up authentication (recommended first step)
./fazt server set-credentials --username admin --password secret123

# Set authentication token (after generating in web interface)
./fazt client set-auth-token --token <YOUR_TOKEN>

# Start the server
./fazt server start

# Start with custom config file
./fazt server start --config /path/to/config.json

# Start on custom port
./fazt server start --port 8080

# Start with custom database
./fazt server start --db /path/to/data.db

# Stop the server
./fazt server stop

# Deploy a site
./fazt client deploy --path ./my-site --domain my-app

# Deploy to remote server
./fazt client deploy --path ./build --domain app --server https://fazt.sh
```

#### Getting Help

```bash
# General help
./fazt --help

# Category-specific help
./fazt server --help
./fazt client --help

# Command-specific help
./fazt server start --help
./fazt client deploy --help
./fazt server set-credentials --help
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
    "path": "./fazt-dev.db"
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
    "domain": "https://fazt.sh",
    "env": "production"
  },
  "database": {
    "path": "~/.config/fazt/data.db"
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
./fazt server set-credentials --username admin --password secret123
```

This will:
1. Create `~/.config/fazt/` directory if needed (with secure 0700 permissions)
2. Generate `config.json` with your credentials
3. Hash password with bcrypt (cost factor 12)
4. Enable authentication
5. Exit (doesn't start server)

Then start the server:

```bash
./fazt server start
```

### Manual Creation

1. Copy `config.example.json`:
   ```bash
   cp config.example.json ~/.config/fazt/config.json
   ```

2. Generate password hash:
   ```bash
   ./fazt server set-credentials --username admin --password temp
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

**Solution**: Ensure `~/.config/fazt/` directory exists and you have write permissions.

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
./fazt server set-credentials --username admin --password secret123

# Start server
./fazt server start

# Deploy (server/client structure)
./fazt client deploy --path . --domain my-site
```

### Key Changes
1. **Server/Client structure** - Clear separation between server management and client operations
2. **Hierarchical commands** - Organized by functional area (server vs client)
3. **Flag-based arguments** - All parameters use explicit flags (no positional arguments)
4. **Removed `--env` flag** - Environment configs not supported
5. **Consolidated config** - All settings in `~/.config/fazt/config.json`
6. **Better error handling** - Clear help and error messages
7. **Improved discoverability** - Help is categorized and context-aware

### Migration Steps
1. Continue using existing config files (compatible)
2. Update deployment scripts to use `client deploy` command
3. Update service files to use `server start` command
4. Update documentation with new CLI examples
