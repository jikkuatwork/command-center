# Command Center - Detailed Build Plan

## Project Overview
- **Name**: Command Center (CC)
- **Domain**: https://cc.toolbomber.com
- **Port**: 4698
- **Stack**: Go + SQLite + Tabler
- **Target**: x64 Linux via SSH

---

## Phase 0: Project Scaffolding (Commit #0)

**Duration**: 5-10 minutes

### Tasks:
- [x] Create project structure
  ```
  command-center/
  ├── cmd/
  │   └── server/
  │       └── main.go
  ├── internal/
  │   ├── config/
  │   │   └── config.go
  │   ├── database/
  │   │   └── db.go
  │   ├── handlers/
  │   │   ├── track.go
  │   │   ├── redirect.go
  │   │   ├── pixel.go
  │   │   ├── webhook.go
  │   │   └── api.go
  │   ├── models/
  │   │   └── models.go
  │   └── notifier/
  │       └── ntfy.go
  ├── web/
  │   ├── static/
  │   │   ├── css/
  │   │   ├── js/
  │   │   └── img/
  │   └── templates/
  │       └── index.html
  ├── migrations/
  │   └── 001_initial.sql
  ├── go.mod
  ├── go.sum
  ├── Makefile
  ├── .env.example
  └── README.md
  ```

- [x] Initialize Go module: `go mod init github.com/yourusername/command-center`
- [x] Create `.env.example` with:
  ```
  PORT=4698
  DB_PATH=./cc.db
  NTFY_TOPIC=your-topic
  NTFY_URL=https://ntfy.sh
  ENV=development
  ```
- [x] Create basic Makefile with build/run/test targets
- [x] Create README with project description

**Commit**: `feat: initial project scaffolding`

---

## Phase 1: Database Layer (Commit #1)

**Duration**: 15-20 minutes

### Tasks:
- [x] Create SQLite schema in `migrations/001_initial.sql`:
  ```sql
  CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,
    tags TEXT, -- JSON array or comma-separated
    source_type TEXT NOT NULL, -- web/pixel/redirect/webhook
    event_type TEXT NOT NULL, -- pageview/click/redirect/webhook
    path TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip_address TEXT,
    query_params TEXT, -- JSON
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_domain (domain),
    INDEX idx_tags (tags),
    INDEX idx_created_at (created_at),
    INDEX idx_source_type (source_type)
  );

  CREATE TABLE IF NOT EXISTS redirects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    destination TEXT NOT NULL,
    tags TEXT,
    click_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_slug (slug)
  );

  CREATE TABLE IF NOT EXISTS webhooks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    endpoint TEXT UNIQUE NOT NULL,
    secret TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER,
    notification_type TEXT NOT NULL,
    message TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id)
  );
  ```

- [x] Implement `internal/database/db.go`:
  - Initialize SQLite connection with WAL mode
  - Run migrations on startup
  - Connection pool configuration
  - Helper functions: `GetDB()`, `Close()`

- [x] Implement `internal/models/models.go`:
  - Event struct
  - Redirect struct
  - Webhook struct
  - Notification struct
  - Helper methods for JSON marshaling tags

- [x] Create mock data generator for testing:
  - Insert 100 sample events across different domains/tags
  - Insert 10 sample redirects
  - Insert 5 sample webhooks

**Commit**: `feat: database layer with SQLite and models`

---

## Phase 2: Configuration & Core Server (Commit #2)

**Duration**: 10-15 minutes

### Tasks:
- [x] Implement `internal/config/config.go`:
  - Load from environment variables
  - Fallback to defaults
  - Validation
  - Config struct with all settings

- [x] Implement `cmd/server/main.go`:
  - Load configuration
  - Initialize database
  - Setup HTTP router (use `gorilla/mux` or `chi`)
  - Graceful shutdown
  - CORS middleware for development
  - Logging middleware
  - Recovery middleware

- [x] Setup routing structure (no handlers yet):
  ```go
  // API routes
  r.HandleFunc("/track", trackHandler).Methods("POST", "OPTIONS")
  r.HandleFunc("/pixel.gif", pixelHandler).Methods("GET")
  r.HandleFunc("/r/{slug}", redirectHandler).Methods("GET")
  r.HandleFunc("/webhook/{endpoint}", webhookHandler).Methods("POST")
  
  // Dashboard API routes
  r.HandleFunc("/api/stats", statsHandler).Methods("GET")
  r.HandleFunc("/api/events", eventsHandler).Methods("GET")
  r.HandleFunc("/api/redirects", redirectsHandler).Methods("GET")
  r.HandleFunc("/api/domains", domainsHandler).Methods("GET")
  r.HandleFunc("/api/tags", tagsHandler).Methods("GET")
  
  // Admin routes
  r.HandleFunc("/api/redirects", createRedirectHandler).Methods("POST")
  r.HandleFunc("/api/webhooks", webhooksHandler).Methods("GET", "POST")
  
  // Static files & dashboard
  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
  r.HandleFunc("/", dashboardHandler).Methods("GET")
  ```

- [x] Test server starts and responds on :4698

**Commit**: `feat: core server setup with routing and middleware`

---

## Phase 3: Tracking Endpoint (Commit #3)

**Duration**: 20-25 minutes

### Tasks:
- [x] Implement `internal/handlers/track.go`:
  - Parse JSON body
  - Extract domain (explicit > hostname from referrer > "unknown")
  - Parse tags (comma-separated string to array)
  - Extract IP, User-Agent, Referrer
  - Validate required fields
  - Insert into events table
  - Return 204 No Content on success
  - Handle OPTIONS for CORS preflight

- [x] Add request validation:
  - Max body size (10KB)
  - Required fields check
  - Sanitize inputs

- [x] Create test script `test_track.sh`:
  ```bash
  #!/bin/bash
  # Test pageview
  curl -X POST http://localhost:4698/track \
    -H "Content-Type: application/json" \
    -d '{"h":"test.com","p":"/page1","e":"pageview","t":["app","test"]}'
  
  # Test with explicit domain
  curl -X POST http://localhost:4698/track \
    -H "Content-Type: application/json" \
    -d '{"d":"custom-domain","p":"/","e":"click","t":["campaign-123"]}'
  
  # Test with query params
  curl -X POST http://localhost:4698/track \
    -H "Content-Type: application/json" \
    -d '{"h":"blog.com","p":"/post","e":"pageview","t":["blog"],"q":{"ref":"twitter"}}'
  ```

- [x] Test tracking endpoint with mock data
- [x] Verify data in SQLite

**Commit**: `feat: tracking endpoint with validation`

---

## Phase 4: Pixel & Redirect Handlers (Commit #4)

**Duration**: 20 minutes

### Tasks:
- [x] Implement `internal/handlers/pixel.go`:
  - Parse query params (domain, tags, source)
  - Extract referrer, IP, User-Agent from request
  - Log event with source_type="pixel"
  - Return 1x1 transparent GIF
  - Set cache headers (no-cache)

- [x] Implement `internal/handlers/redirect.go`:
  - Extract slug from URL
  - Lookup redirect in database
  - Log event with source_type="redirect"
  - Parse tags from query string if present
  - Increment click_count
  - Return 302 redirect
  - Handle 404 for invalid slugs

- [x] Create test scripts:
  ```bash
  # Test pixel
  curl "http://localhost:4698/pixel.gif?domain=newsletter&tags=dec,email"
  
  # Test redirect (setup test redirect first)
  curl -I "http://localhost:4698/r/test123?tags=reddit,promo"
  ```

- [x] Add helper function to create test redirects via database

**Commit**: `feat: pixel tracking and redirect handler`

---

## Phase 5: Webhook Handler (Commit #5)

**Duration**: 15 minutes

### Tasks:
- [x] Implement `internal/handlers/webhook.go`:
  - Validate webhook endpoint exists and is active
  - Verify secret if configured (HMAC SHA256)
  - Parse JSON payload (flexible structure)
  - Log event with source_type="webhook"
  - Extract useful fields (event type, source, etc.)
  - Return 200 OK with confirmation JSON

- [x] Create webhook registration helpers
- [x] Mock webhook sender for testing:
  ```bash
  # Test webhook
  curl -X POST http://localhost:4698/webhook/deployment \
    -H "Content-Type: application/json" \
    -d '{"event":"deploy","project":"my-site","status":"success"}'
  ```

- [x] Test webhook logging

**Commit**: `feat: webhook handler with secret validation`

---

## Phase 6: Dashboard API Endpoints (Commit #6)

**Duration**: 30-40 minutes

### Tasks:
- [ ] Implement `internal/handlers/api.go` with:

  **GET /api/stats**:
  - Total events (today, week, month, all-time)
  - Events by source_type
  - Top 10 domains
  - Top 10 tags
  - Events timeline (hourly for today, daily for month)
  - Response: JSON

  **GET /api/events**:
  - Query params: domain, tags, source_type, limit, offset, from, to
  - Paginated event list
  - Filtering and sorting
  - Response: JSON array

  **GET /api/domains**:
  - List all unique domains with event counts
  - Sort by count desc
  - Response: JSON array

  **GET /api/tags**:
  - List all unique tags with usage counts
  - Response: JSON array (tag cloud data)

  **GET /api/redirects**:
  - List all redirects with click counts
  - Sort by clicks or created_at
  - Response: JSON array

  **POST /api/redirects**:
  - Create new redirect
  - Body: {slug, destination, tags}
  - Validation: slug uniqueness, valid URL
  - Response: created redirect JSON

  **GET /api/webhooks**:
  - List all configured webhooks
  - Response: JSON array

  **POST /api/webhooks**:
  - Create new webhook endpoint
  - Body: {name, endpoint, secret}
  - Response: created webhook JSON

- [ ] Add pagination helpers
- [ ] Add date range parsing
- [ ] Create comprehensive test script for all API endpoints

**Commit**: `feat: dashboard API endpoints with filtering`

---

## Phase 7: ntfy.sh Integration (Commit #7)

**Duration**: 15 minutes

### Tasks:
- [x] Implement `internal/notifier/ntfy.go`:
  - Send notification function
  - Mock mode for testing (log instead of HTTP call)
  - Error handling and retries
  - Message formatting

- [x] Add notification triggers:
  - Traffic spike detection (10x avg in last hour)
  - New domain detection
  - Webhook events (configurable)
  - Error events

- [x] Create notification rules system (simple config)
- [x] Test with mock ntfy.sh calls (log output)

**Commit**: `feat: ntfy.sh integration with event triggers`

---

## Phase 8: Frontend - Tabler Integration (Commit #8)

**Duration**: 30 minutes

### Tasks:
- [x] Download Tabler from CDN or npm (use dist files):
  - tabler.min.css
  - tabler.min.js
  - tabler-icons.min.css

- [x] Place in `web/static/`:
  ```
  web/static/
  ├── css/
  │   ├── tabler.min.css
  │   ├── tabler-icons.min.css
  │   └── custom.css
  ├── js/
  │   ├── tabler.min.js
  │   └── app.js
  └── img/
      └── logo-placeholder.svg
  ```

- [x] Create `web/templates/index.html`:
  - Base HTML structure
  - Tabler theme setup (light/dark mode toggle)
  - Navigation sidebar with sections:
    - Dashboard (overview)
    - Analytics (detailed views)
    - Redirects
    - Webhooks
    - Settings
  - Empty content area for dynamic loading
  - Mobile-responsive header
  - PWA meta tags

- [x] Create placeholder logo SVG
- [x] Test static file serving
- [x] Verify responsive layout on different viewport sizes

**Commit**: `feat: tabler frontend integration with responsive layout`

---

## Phase 9: Dashboard Overview Page (Commit #9)

**Duration**: 40 minutes

### Tasks:
- [ ] Create `web/static/js/app.js` with:
  - API client functions (fetch wrappers)
  - Router for SPA navigation
  - State management (simple object)
  - Theme toggle (localStorage persistence)

- [ ] Implement Overview dashboard in HTML/JS:
  - **Stats cards** (4 cards in grid):
    - Total events today
    - Total events this week
    - Total unique domains
    - Total redirects clicks
  
  - **Live traffic graph** (Chart.js):
    - Last 24 hours, hourly breakdown
    - Line chart with smooth curves
    - Responsive
  
  - **Top domains** (table):
    - Domain name
    - Event count
    - Percentage
    - Mini sparkline
  
  - **Top tags** (tag cloud):
    - Visual tag sizes based on usage
    - Clickable to filter
  
  - **Recent events** (table):
    - Last 20 events
    - Columns: Time, Domain, Type, Path, Tags
    - Truncate long strings with tooltip
    - Real-time updates (poll every 10s)

- [ ] Add Chart.js via CDN
- [ ] Style with custom CSS for vivid graphs
- [ ] Test with mock data from API

**Commit**: `feat: dashboard overview with live stats and graphs`

---

## Phase 10: Analytics Deep Dive Page (Commit #10)

**Duration**: 35 minutes

### Tasks:
- [ ] Create Analytics view with:
  
  **Filters panel**:
  - Date range picker (today, week, month, custom)
  - Domain multi-select dropdown
  - Tags multi-select
  - Source type checkboxes
  - Apply/Reset buttons

  **Charts section**:
  - Events timeline (bar chart, daily/hourly toggle)
  - Source type pie chart
  - Top pages bar chart
  - Top referrers list

  **Data table**:
  - Paginated events list
  - Sortable columns
  - Export to CSV button (client-side)
  - Filter indicators (active filters shown as chips)

- [ ] Implement filter logic in JS
- [ ] Add debounced search
- [ ] URL state persistence (query params)
- [ ] Test filtering combinations

**Commit**: `feat: analytics deep dive with filtering and charts`

---

## Phase 11: Redirects Management (Commit #11)

**Duration**: 25 minutes

### Tasks:
- [ ] Create Redirects page with:
  
  **Create redirect form**:
  - Slug input (validation, URL-safe)
  - Destination URL input
  - Tags input (comma-separated)
  - Preview: show full URL
  - Copy button for generated link
  - Submit creates via API

  **Redirects list table**:
  - Columns: Slug, Destination, Tags, Clicks, Created
  - Click to copy full URL
  - Click count badge
  - Delete button (with confirmation)
  - Sort by clicks/date

  **Click analytics per redirect**:
  - Modal/expandable row
  - Mini chart of clicks over time
  - Referrer breakdown

- [ ] Form validation
- [ ] Error handling
- [ ] Success toasts (Tabler built-in)

**Commit**: `feat: redirect management with creation and analytics`

---

## Phase 12: Webhooks Configuration (Commit #12)

**Duration**: 20 minutes

### Tasks:
- [ ] Create Webhooks page with:
  
  **Webhook list**:
  - Table: Name, Endpoint, Status, Created
  - Toggle active/inactive
  - Show secret (masked, reveal on click)
  - Delete button

  **Create webhook form**:
  - Name input
  - Endpoint path (auto-prefixed with /webhook/)
  - Secret generation button
  - Test webhook button (sends mock payload)

  **Recent webhook events**:
  - Last 50 webhook calls
  - Timestamp, endpoint, payload preview
  - Status (success/fail)

- [ ] Webhook testing from UI
- [ ] Copy endpoint URL + curl example

**Commit**: `feat: webhook configuration and monitoring`

---

## Phase 13: Settings & Preferences (Commit #13)

**Duration**: 20 minutes

### Tasks:
- [ ] Create Settings page with:
  
  **Appearance**:
  - Theme selector (light/dark/auto)
  - Color scheme picker (if implementing themes)
  - Accent color selection

  **Notifications**:
  - ntfy.sh topic configuration
  - Enable/disable specific notification types
  - Test notification button

  **Tracking**:
  - Generate tracking script snippet
  - Generate pixel HTML
  - API key display (if implementing auth)

  **Data management**:
  - Database stats (size, event count)
  - Export all data (JSON)
  - Danger zone: Clear old events (>90 days)

- [ ] Implement client-side settings persistence (localStorage)
- [ ] Settings sync with server (if needed)

**Commit**: `feat: settings page with customization options`

---

## Phase 14: PWA Configuration (Commit #14)

**Duration**: 15 minutes

### Tasks:
- [ ] Create `web/static/manifest.json`:
  ```json
  {
    "name": "Command Center",
    "short_name": "CC",
    "start_url": "/",
    "display": "standalone",
    "background_color": "#ffffff",
    "theme_color": "#206bc4",
    "icons": [
      {
        "src": "/static/img/icon-192.png",
        "sizes": "192x192",
        "type": "image/png"
      },
      {
        "src": "/static/img/icon-512.png",
        "sizes": "512x512",
        "type": "image/png"
      }
    ]
  }
  ```

- [ ] Create placeholder icons (simple CC logo)
- [ ] Add manifest link to HTML
- [ ] Create basic service worker `sw.js`:
  - Cache static assets
  - Offline fallback page
  - Cache API responses (short TTL)

- [ ] Register service worker in app.js
- [ ] Test PWA installability
- [ ] Add "Add to Home Screen" prompt

**Commit**: `feat: PWA support with manifest and service worker`

---

## Phase 15: Tracking Client Script (Commit #15)

**Duration**: 20 minutes

### Tasks:
- [ ] Create `web/static/js/track.min.js`:
  - Self-contained, no dependencies
  - Auto-capture: hostname, path, referrer
  - Accept query params: domain, tags
  - Configurable via `window.CC_CONFIG`
  - Auto-track pageviews on load
  - Expose `window.ccTrack(event, data)` for manual tracking
  - Click tracking (data-cc-track attribute)
  - Form submission tracking
  - Scroll depth tracking (optional)
  - Error handling (silent failures)

- [ ] Example usage documentation:
  ```html
  <!-- Simple pageview tracking -->
  <script src="https://cc.toolbomber.com/static/js/track.min.js"></script>
  
  <!-- With config -->
  <script>
    window.CC_CONFIG = {
      domain: 'my-site',
      tags: ['app', 'production']
    };
  </script>
  <script src="https://cc.toolbomber.com/static/js/track.min.js"></script>
  
  <!-- Manual tracking -->
  <button onclick="ccTrack('button-click', {id: 'signup'})">Sign Up</button>
  
  <!-- Auto-track clicks -->
  <a href="/pricing" data-cc-track>Pricing</a>
  ```

- [ ] Minify script
- [ ] Test on sample HTML pages

**Commit**: `feat: tracking client script with auto-capture`

---

## Phase 16: Dark Mode & Theming (Commit #16)

**Duration**: 25 minutes

### Tasks:
- [ ] Implement theme system in `custom.css`:
  - CSS custom properties for colors
  - Multiple theme variants (default, purple, green, orange)
  - Dark mode variants for each theme
  - Smooth transitions

- [ ] Add theme switcher UI:
  - Dropdown in settings
  - Preview swatches
  - Apply immediately

- [ ] Persist theme preference:
  - localStorage
  - Apply on page load (prevent flash)
  - Respect system preference initially

- [ ] Ensure all charts adapt to theme:
  - Chart.js theme colors
  - Update on theme change

- [ ] Test all themes in light/dark mode
- [ ] Verify accessibility (contrast ratios)

**Commit**: `feat: theming system with multiple color schemes`

---

## Phase 17: Mobile Optimizations (Commit #17)

**Duration**: 20 minutes

### Tasks:
- [ ] Mobile-specific CSS refinements:
  - Larger touch targets (min 44x44px)
  - Simplified table views (stack columns)
  - Bottom navigation for mobile
  - Swipeable cards
  - Pull-to-refresh (if feasible)

- [ ] Optimize chart rendering for mobile:
  - Responsive canvas sizing
  - Simplified tooltips
  - Touch gestures support

- [ ] Mobile menu improvements:
  - Slide-out drawer
  - Gesture close
  - Quick actions at bottom

- [ ] Test on various viewport sizes:
  - iPhone SE (375px)
  - iPhone 14 (390px)
  - iPad (768px)
  - MacBook Air 14" (1512px)

- [ ] Performance optimizations:
  - Lazy load charts
  - Virtual scrolling for long lists
  - Debounced resize handlers

**Commit**: `feat: mobile optimizations and responsive enhancements`

---

## Phase 18: Error Handling & Loading States (Commit #18)

**Duration**: 20 minutes

### Tasks:
- [ ] Add global error handling:
  - API error interceptor
  - User-friendly error messages
  - Retry logic for failed requests
  - Offline detection

- [ ] Implement loading states:
  - Skeleton loaders for tables/cards
  - Progress indicators for long operations
  - Disable buttons during submission
  - Loading overlays

- [ ] Add empty states:
  - No data illustrations
  - Helpful onboarding messages
  - Quick action CTAs

- [ ] Toast notifications:
  - Success confirmations
  - Error alerts
  - Info messages
  - Auto-dismiss timers

- [ ] Form validation feedback:
  - Inline error messages
  - Field highlighting
  - Submit prevention

**Commit**: `feat: comprehensive error handling and loading states`

---

## Phase 19: Performance & Optimization (Commit #19)

**Duration**: 20 minutes

### Tasks:
- [ ] Backend optimizations:
  - Add database indexes (already in schema, verify)
  - Query optimization for aggregations
  - Response compression (gzip)
  - ETag support for static files
  - Connection pooling tuning

- [ ] Frontend optimizations:
  - Minify CSS/JS (build step)
  - Lazy load non-critical JS
  - Defer chart rendering until visible
  - Memoize expensive computations
  - Virtual scrolling for large lists

- [ ] Caching strategy:
  - Cache-Control headers
  - Service worker caching rules
  - LocalStorage for recent data
  - Invalidation on updates

- [ ] Load time improvements:
  - Critical CSS inline
  - Preload key resources
  - Font optimization

- [ ] Test performance:
  - Lighthouse audit (target 90+ on all metrics)
  - Bundle size analysis
  - API response times

**Commit**: `perf: optimization pass for speed and efficiency`

---

## Phase 20: Documentation & Deployment Prep (Commit #20)

**Duration**: 25 minutes

### Tasks:
- [ ] Update README.md with:
  - Project description
  - Features list
  - Installation instructions
  - Configuration guide
  - API documentation
  - Usage examples
  - Deployment guide

- [ ] Create deployment artifacts:
  - Build script (compile binary)
  - systemd service file:
    ```ini
    [Unit]
    Description=Command Center
    After=network.target

    [Service]
    Type=simple
    User=your-user
    WorkingDirectory=/opt/command-center
    ExecStart=/opt/command-center/cc-server
    Restart=always
    RestartSec=5

    [Install]
    WantedBy=multi-user.target
    ```
  - nginx config example:
    ```nginx
    server {
        listen 80;
        server_name cc.toolbomber.com;
        
        location / {
            proxy_pass http://localhost:4698;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
    ```

- [ ] Create deployment script `deploy.sh`:
  ```bash
  #!/bin/bash
  # Build binary
  GOOS=linux GOARCH=amd64 go build -o cc-server ./cmd/server
  
  # Create tarball
  tar -czf cc-release.tar.gz cc-server web/ migrations/
  
  # SCP to server (you'll run this manually)
  echo "Run: scp cc-release.tar.gz user@server:/opt/command-center/"
  ```

- [ ] Create `CHANGELOG.md` with v0.1.0 notes
- [ ] Add license file (MIT)
- [ ] Create `.gitignore`

**Commit**: `docs: comprehensive documentation and deployment prep`

---

## Phase 21: Testing & Bug Fixes (Commit #21)

**Duration**: 30-40 minutes

### Tasks:
- [ ] Comprehensive testing:
  - Test all API endpoints with various payloads
  - Test tracking from different "sites"
  - Test redirects with various tags
  - Test webhook calls
  - Test all dashboard views
  - Test filters and pagination
  - Test theme switching
  - Test mobile views
  - Test PWA installation
  - Test offline behavior

- [ ] Edge case testing:
  - Empty database (new install)
  - Large dataset (1000+ events)
  - Invalid inputs
  - Missing fields
  - Concurrent requests
  - Very long strings
  - Special characters in tags/domains

- [ ] Browser compatibility:
  - Chrome/Edge
  - Firefox
  - Safari (mobile & desktop)

- [ ] Fix any bugs discovered
- [ ] Performance profiling
- [ ] Memory leak check

**Commit**: `test: comprehensive testing and bug fixes`

---

## Phase 22: Final Polish (Commit #22)

**Duration**: 20 minutes

### Tasks:
- [ ] Code cleanup:
  - Remove debug logs
  - Remove unused code
  - Consistent formatting
  - Add missing comments
  - Fix TODOs

- [ ] UI polish:
  - Consistent spacing
  - Animation timing
  - Icon consistency
  - Typography refinement
  - Color consistency

- [ ] Accessibility audit:
  - ARIA labels
  - Keyboard navigation
  - Focus indicators
  - Screen reader testing

- [ ] Final checks:
  - All links work
  - All buttons work
  - Forms validate properly
  - Graphs render correctly
  - No console errors

**Commit**: `polish: final refinements and cleanup`

---

## Phase 23: Build & Release (Commit #23)

**Duration**: 15 minutes

### Tasks:
- [ ] Version bump to v0.1.0
- [ ] Build production binary:
  ```bash
  make build
  # or
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o cc-server ./cmd/server
  ```

- [ ] Create release package:
  ```bash
  tar -czf command-center-v0.1.0.tar.gz \
    cc-server \
    web/ \
    migrations/ \
    .env.example \
    README.md
  ```

- [ ] Create GitHub release (if using GitHub)
- [ ] Tag commit: `git tag v0.1.0`

**Commit**: `release: v0.1.0 - initial release`
