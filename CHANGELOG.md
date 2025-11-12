# Changelog

All notable changes to Command Center will be documented in this file.

## [0.2.0] - 2024-11-12

### Added

**Authentication & Security**
- Username/password authentication with bcrypt hashing
- Session management with secure cookies
- Rate limiting (5 attempts per 15 min per IP)
- Audit logging for all security events
- Brute-force protection with automatic lockout
- Login page with modern UI
- Session expiry and refresh
- Remember me functionality (7-day sessions)

**Configuration System**
- JSON-based configuration files
- CLI flags: --config, --db, --port, --username, --password, --env
- Environment-specific configs (development/production)
- Simple credential setup: `./cc-server --username admin --password pass`
- Automatic config directory creation
- Config validation on startup
- Backward compatible with environment variables

**Security Enhancements**
- Security headers (CSP, HSTS, X-Frame-Options, etc.)
- CSRF protection via SameSite cookies
- Automatic file permission enforcement (0600 for config/db)
- Session hijacking prevention
- IP-based rate limiting
- Audit trail for all authentication events

**Database Improvements**
- Migration tracking system
- Automatic backups (keeps last 5)
- New default location: ~/.config/cc/data.db
- Audit logs table
- Migration versioning

**CLI Improvements**
- --version flag with build info
- --help flag with comprehensive documentation
- --verbose and --quiet modes
- Beautiful startup banner
- Production warnings
- Better error messages

**Documentation**
- SECURITY.md - Complete security guide
- CONFIGURATION.md - Configuration reference
- UPGRADE.md - v0.1.0 to v0.2.0 migration guide
- Updated README with v0.2.0 features

**Testing**
- Comprehensive authentication flow test script
- End-to-end testing capabilities

### Changed
- Default config location: ~/.config/cc/config.json
- Default database location: ~/.config/cc/data.db
- Dashboard now protected by authentication (when enabled)
- Improved startup messages with visual banner
- Enhanced Makefile with new targets

### Security
- Dashboard requires authentication (configurable)
- Tracking endpoints remain public
- Protected routes: /, /api/* (except /api/login)
- Public routes: /track, /pixel.gif, /r/*, /webhook/*, /static/*, /login, /health
- File permissions automatically enforced
- Secure cookie defaults in production

### Deprecated
- Environment variables (still work but use JSON config instead)

## [0.1.0] - 2025-11-11

### Added
- Initial release of Command Center
- Universal tracking endpoint with domain auto-detection
- 1x1 transparent GIF pixel tracking
- URL redirect service with click tracking
- Webhook receiver with HMAC SHA256 validation
- Real-time dashboard with interactive charts
- Analytics page with filtering (domain, source, search)
- Redirects management interface
- Webhooks configuration interface
- Settings page with integration snippets
- PWA support with service worker
- Client-side tracking script (track.min.js)
- Light/dark theme toggle with persistence
- SQLite database with WAL mode
- ntfy.sh integration for notifications
- RESTful API with 8 endpoints
- Comprehensive test scripts
- Production-ready deployment configuration

### Features
- **Backend**: Go + SQLite with proper indexing
- **Frontend**: Tabler UI with Chart.js visualizations
- **Database**: 4 tables (events, redirects, webhooks, notifications)
- **API**: Complete CRUD operations for all resources
- **Security**: HMAC validation, input sanitization, prepared statements
- **Performance**: Database indexing, service worker caching, auto-refresh

### Documentation
- Complete README with installation instructions
- API endpoint documentation
- Deployment guide (systemd, nginx)
- Usage examples for all tracking methods
- Troubleshooting section

### Testing
- 4 comprehensive test scripts
- All endpoints tested and validated
- Mock data generator for development

---

**Total Commits**: 13
**Lines of Code**: ~5000+
**Build Time**: ~8 hours (autonomous session)
