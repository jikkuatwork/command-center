# fazt.sh

A unified analytics, monitoring, tracking platform, and **Personal Cloud** with static hosting and serverless JavaScript functions.

**Now a completely self-contained "Cartridge" Application.**

## Features

### Personal Cloud (PaaS)
- **Single Binary & Single DB** - The entire platform runs from `fazt` executable and `data.db`.
- **Zero Dependencies** - No Nginx required. Native automatic HTTPS via Let's Encrypt (CertMagic).
- **Virtual Filesystem (VFS)** - Sites are stored in the SQLite database, not on disk.
- **Static Site Hosting** - Deploy static websites via CLI.
- **Serverless JavaScript** - Run JavaScript functions with `main.js` (loaded from DB).
- **Key-Value Store** - Persistent data storage for serverless apps.
- **WebSocket Support** - Real-time communication.

### Analytics & Tracking
- **Universal Tracking Endpoint** - Auto-detects domains and tracks pageviews/events.
- **Real-time Dashboard** - Interactive charts and live updates.

## Quick Start

### Prerequisites
- Go 1.20+ (for building)
- Linux/macOS/Windows

### Installation

```bash
# Build the server
go build -o fazt ./cmd/server

# Initialize configuration
./fazt server init \
  --username admin \
  --password secret123 \
  --domain https://your-domain.com \
  --env production

# Start the server (HTTPS is disabled by default)
./fazt server start
```

### Enabling HTTPS (Production)

To enable automatic HTTPS (Let's Encrypt), update your `config.json` (usually in `~/.config/fazt/`):

```json
{
  "https": {
    "enabled": true,
    "email": "you@example.com",
    "staging": false
  }
}
```

Then restart the server. It will bind to ports 80 and 443 automatically.

## CLI Commands

```bash
# Server Management
./fazt server init ...       # First time setup
./fazt server start          # Start server
./fazt server status         # Check status
./fazt server set-config     # Update settings

# Deployment
./fazt client set-auth-token --token <TOKEN>
./fazt client deploy --path ./my-site --domain my-app
```

## "Cartridge" Architecture

**fazt** follows a "Cartridge" architecture:
- **State**: All state (Users, Analytics, Sites, Files, SSL Certs) lives in a single SQLite file (`data.db`).
- **Stateless Binary**: The `fazt` binary contains all logic. Updating is as simple as replacing the binary.
- **Backup/Restore**: Just copy `data.db`.

## License
MIT License