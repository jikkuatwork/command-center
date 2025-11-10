# Command Center (CC)

A unified analytics, monitoring, and tracking platform that provides comprehensive web analytics with real-time dashboard capabilities.

## Features

- **Universal Tracking Endpoint** - Auto-detects domains and tracks pageviews, clicks, and events
- **Tracking Pixel** - Traditional 1x1 GIF pixel for web analytics
- **Redirect Service** - URL shortening with built-in click tracking
- **Webhook Receiver** - Accept webhook events from external services
- **Real-time Dashboard** - Interactive charts, filtering, and live updates
- **PWA Support** - Mobile-ready progressive web app
- **Push Notifications** - ntfy.sh integration for alerts

## Technology Stack

- **Backend**: Go (Golang) with SQLite (WAL mode)
- **Frontend**: Tabler (vanilla JavaScript)
- **Charts**: Chart.js
- **Database**: SQLite with proper indexing
- **Notifications**: ntfy.sh

## Project Status

**Phase 0/23 Complete** - Planning phase completed

This project is currently in the planning phase with comprehensive documentation ready for implementation. The full implementation requires 23 development phases.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/yourusername/cc.git
cd cc

# Build the server (planned)
make build

# Run the server (planned)
./cc-server
```

The server will run on port 4698.

## Usage

### Basic Tracking
```html
<!-- Include the tracking script -->
<script src="https://your-domain.com/static/js/track.min.js"></script>

<!-- Use the tracking pixel -->
<img src="https://your-domain.com/pixel" style="display:none;">
```

### Redirect Service
```
https://your-domain.com/r/your-slug
```

### Webhook Receiver
```
POST https://your-domain.com/webhook
```

## Documentation

- [Deployment Guide](./koder/docs/meta/01_manual.md)
- [Development Plan](./koder/plans/01_initial-build.md)
- [Getting Started](./koder/start.md)

## License

[Add your license here]