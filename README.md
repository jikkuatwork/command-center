# Command Center (CC) v0.3.0

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
git clone https://github.com/jikkuatwork/command-center.git
cd command-center

# Build the server
go build -o cc-server ./cmd/server

# Run the server
./cc-server
```

The server starts on **port 4698**. Access the dashboard at `http://localhost:4698`

### Setup Authentication

```bash
# Set up admin credentials
./cc-server --username admin --password your-secure-password

# Start the server
./cc-server
```

## Personal Cloud Usage

### 1. Create an API Key

Visit `http://localhost:4698/hosting` and create a new deploy key.

Save the token to `~/.cc-token`:
```bash
echo "your-api-token" > ~/.cc-token
```

### 2. Deploy a Static Site

```bash
# In your website directory
cd my-website/
cc-server deploy my-site

# Or with explicit token
cc-server deploy my-site --token your-api-token
```

Your site is now live at `http://my-site.localhost:4698`

### 3. Create a Serverless App

Add a `main.js` file to enable serverless:

```javascript
// main.js - Simple API endpoint
const name = req.query.split('=')[1] || 'World';
res.json({ message: `Hello, ${name}!` });
```

Deploy and access:
```bash
curl http://my-site.localhost:4698/?name=Claude
# {"message":"Hello, Claude!"}
```

### 4. Use the Key-Value Store

```javascript
// main.js - Counter app with persistent storage
let count = db.get("visits") || 0;
count++;
db.set("visits", count);
res.json({ visits: count });
```

Available `db` methods:
- `db.get(key)` - Retrieve a value
- `db.set(key, value)` - Store a value (auto-serializes objects)
- `db.delete(key)` - Remove a key

### 5. Use Environment Variables

Set secrets in the Hosting UI, then access in JavaScript:

```javascript
// main.js - Use API keys securely
const apiKey = process.env.OPENAI_API_KEY;
res.json({ configured: !!apiKey });
```

### 6. WebSocket Broadcasting

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
Description=Command Center
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/command-center
ExecStart=/opt/command-center/cc-server
Restart=always

[Install]
WantedBy=multi-user.target
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

**Command Center v0.3.0** | Analytics + Personal Cloud | Port 4698
