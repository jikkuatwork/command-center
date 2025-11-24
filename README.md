# fazt.sh v0.4.0

A unified analytics, monitoring, tracking platform, and **Personal Cloud** with static hosting and serverless JavaScript functions.

## Features

### Analytics & Tracking
- **Universal Tracking Endpoint** - Auto-detects domains and tracks pageviews, clicks, and events
- **Tracking Pixel** - 1x1 transparent GIF for email/image tracking
- **Redirect Service** - URL shortening with click tracking
- **Webhook Receiver** - Accept webhook events from external services with HMAC validation
- **Real-time Dashboard** - Interactive charts, filtering, and live updates

### Personal Cloud (NEW in v0.3.0)
- **Static Site Hosting** - Deploy static websites via CLI or API (Surge-like)
- **Serverless JavaScript** - Run JavaScript functions with `main.js`
- **Key-Value Store** - Persistent data storage for serverless apps
- **Environment Variables** - Secure secrets management per site
- **WebSocket Support** - Real-time communication with broadcast channels
- **API Key Management** - Generate deploy tokens for CLI access

## Quick Start

### Prerequisites

- Go 1.20+ (with CGO support for SQLite)
- Linux/macOS or Windows with WSL

### Installation

```bash
# Clone the repository
git clone https://github.com/jikkuatwork/fazt.sh.git
cd fazt.sh

# Build the server
go build -o fazt ./cmd/server

# Initialize configuration (first-time setup)
./fazt server init --username admin --password secret123 --domain https://fazt.example.com

# Start the server
./fazt server start
```

The server starts on **port 4698**. Access the dashboard at `http://localhost:4698`

### CLI Commands

fazt.sh uses a server/client structure:

```bash
# Server Management
./fazt server init --username admin --password secret123 --domain https://fazt.example.com
./fazt server status
./fazt server set-credentials --username admin --password newsecret
./fazt server set-config --domain https://new.example.com --port 8080 --env production
./fazt server start [--port 8080] [--config /path/to/config.json)]
./fazt server stop

# Client/Deployment
./fazt client set-auth-token --token <TOKEN>
./fazt client deploy --path ./my-site --domain my-app
./fazt deploy --path ./my-site --domain my-app  # Shortcut alias

# Help
./fazt --help              # Main help
./fazt server --help       # Server commands
./fazt client --help       # Client commands
```

**Note**: Flags must come before positional arguments. The `~/.config/fazt/` directory is created automatically with secure permissions.

## Personal Cloud Usage

### 1. Create an Authentication Token

Visit `http://localhost:4698/hosting` (now available in the sidebar) and create a new authentication token. Then configure it via CLI:

```bash
./fazt client set-auth-token --token <YOUR_TOKEN>
```

The authentication token will be saved to `~/.config/fazt/config.json` and used automatically for deployments.

### 2. Deploy a Static Site

```bash
# In your website directory
cd my-website/
fazt client deploy --path . --domain my-site

# Or deploy any directory
fazt client deploy --path /path/to/site --domain my-app
```

Your site is now live at `http://my-site.localhost:4698`

### 3. Deploy to Remote Server

```bash
# Deploy to a remote fazt.sh instance
fazt client deploy --path ./build --domain my-app --server https://fazt.sh
```

### 4. Create a Serverless App

Add a `main.js` file to enable serverless:

```javascript
// main.js - Simple API endpoint
const name = req.query.split('=')[1] || 'World';
res.json({ message: `Hello, ${name}!` });
```

Deploy and access:
```bash
fazt client deploy --path . --domain my-app
curl http://my-app.localhost:4698/?name=Claude
# {"message":"Hello, Claude!"}
```

### 5. Use the Key-Value Store

```javascript
// main.js - Counter app with persistent storage
let count = db.get("visits") || 0;
count++;
db.set("visits", count);
res.json({ visits: count });
```

Available `db` methods:
- `db.get(key)` - Retrieve a value (auto-parses JSON)
- `db.set(key, value)` - Store a value (auto-stringifies)
- `db.delete(key)` - Remove a key

### 6. Use Environment Variables

Set secrets in the Hosting UI, then access in JavaScript:

```javascript
// main.js - Use API keys securely
const apiKey = process.env.OPENAI_API_KEY;
res.json({ configured: !!apiKey });
```

### 7. WebSocket Broadcasting

```javascript
// main.js - Broadcast to all connected clients
socket.broadcast("New message: " + req.body);
res.json({ clients: socket.clients() });
```

Clients connect to `ws://my-site.localhost:4698/ws`

## JavaScript Runtime API

### Request Object (`req`)
```javascript
req.method   // "GET", "POST", etc.
req.path     // "/api/users"
req.query    // "id=123&name=test"
req.headers  // { "Content-Type": "application/json" }
req.body     // Request body as string
```

### Response Object (`res`)
```javascript
res.send(html)       // Send HTML response
res.json(object)     // Send JSON response
res.status(code)     // Set status code
res.header(k, v)     // Set response header
```

### Storage (`db`)
```javascript
db.get(key)          // Get value (auto-parses JSON)
db.set(key, value)   // Set value (auto-stringifies)
db.delete(key)       // Delete key
```

### Environment (`process.env`)
```javascript
process.env.API_KEY  // Access environment variables
```

### WebSocket (`socket`)
```javascript
socket.broadcast(msg) // Send to all connected clients
socket.clients()      // Get connected client count
```

### Console (`console`)
```javascript
console.log(...)     // Log to server stdout
```

## Example Apps

### Static Site (`examples/static-site/`)
```
index.html
style.css
script.js
```

### Counter App (`examples/counter-app/`)
```javascript
// main.js
let count = db.get("count") || 0;
count++;
db.set("count", count);

res.send(`
<!DOCTYPE html>
<html>
<head><title>Counter</title></head>
<body>
  <h1>Visit Count: ${count}</h1>
</body>
</html>
`);
```

### Chat App (`examples/chat-app/`)
See `examples/chat-app/` for a complete WebSocket chat implementation.

## Analytics API

### Tracking Website Analytics

```html
<script src="https://your-domain.com/static/js/track.min.js"></script>
```

### Tracking Pixel

```html
<img src="https://your-domain.com/pixel.gif?domain=newsletter&tags=email" style="display:none">
```

### URL Redirects

```
https://your-domain.com/r/your-slug?tags=twitter,promo
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/track` | POST | Track events |
| `/pixel.gif` | GET | Tracking pixel |
| `/r/{slug}` | GET | Redirect with tracking |
| `/webhook/{endpoint}` | POST | Receive webhooks |
| `/api/deploy` | POST | Deploy site (requires API key) |
| `/api/sites` | GET | List hosted sites |
| `/api/keys` | GET/POST/DELETE | Manage API keys |
| `/api/envvars` | GET/POST/DELETE | Manage env vars |
| `/api/deployments` | GET | List deployments |
| `/api/stats` | GET | Dashboard statistics |
| `/api/events` | GET | Paginated events |

## Security

- **100ms execution timeout** for JavaScript
- **Path traversal protection** on static files
- **100MB file size limit** in deployments
- **Site isolation** - KV store and env vars scoped per site
- **bcrypt hashed** API keys
- **HMAC SHA256** webhook validation

## Deployment

### Nginx with Subdomain Routing

```nginx
server {
    listen 80;
    server_name *.your-domain.com your-domain.com;

    location / {
        proxy_pass http://localhost:4698;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Systemd Service

```ini
[Unit]
Description=fazt.sh
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/fazt
ExecStart=/opt/fazt/fazt server start
Restart=always
PIDFile=/opt/fazt/fazt.pid

[Install]
WantedBy=multi-user.target
```

Start/stop commands:
```bash
sudo systemctl start fazt
sudo systemctl stop fazt
sudo systemctl status fazt
```

## Architecture

- **Single Binary** - One executable with embedded static files
- **SQLite Database** - Persistent storage with WAL mode
- **Goja JS Runtime** - Embedded JavaScript engine
- **Gorilla WebSocket** - Real-time communication
- **Host-based Routing** - Subdomains map to sites

## License

MIT License

---

**fazt.sh v0.3.0** | Analytics + Personal Cloud | Port 4698
