# Security Guide

## Overview

Command Center v0.2.0 includes comprehensive security features to protect your tracking data and dashboard access.

## Authentication

### Enabling Authentication

Set up authentication with a simple command:

```bash
./cc-server --username admin --password your-secure-password
```

This creates/updates the config file at `~/.config/cc/config.json` with:
- Username stored in plain text
- Password hashed using bcrypt (cost factor 12)
- Auth automatically enabled

### Password Requirements

**Minimum Requirements:**
- At least 8 characters long

**Recommendations for Strong Passwords:**
- 12+ characters
- Mix of uppercase and lowercase letters
- Include numbers
- Include special characters (!@#$%^&*)

### Changing Passwords

To change your password, run the setup command again:

```bash
./cc-server --username admin --password new-password
```

## Session Management

### Session Security Features

- **HTTPOnly Cookies**: Prevents JavaScript access to session cookies
- **Secure Flag**: Cookies only sent over HTTPS in production
- **SameSite=Strict**: Protection against CSRF attacks
- **24-Hour Expiry**: Sessions expire after 24 hours of inactivity
- **Remember Me**: Option to extend sessions to 7 days
- **Session Refresh**: Activity extends session lifetime automatically

### Session Storage

- Sessions stored in-memory (lost on server restart)
- Session cleanup runs every 5 minutes
- Sessions bound to session ID only (portable across IPs)

## Rate Limiting

### Brute-Force Protection

- **5 failed login attempts** per IP address
- **15-minute lockout** after limit exceeded
- Automatic reset after successful login
- Attempts tracked per IP (supports proxy headers)

### IP Detection

The system checks the following headers in order:
1. `X-Forwarded-For` (first IP in chain)
2. `X-Real-IP`
3. `RemoteAddr` (direct connection)

## Audit Logging

### What's Logged

All security events are logged to the database:
- Login attempts (success/failure)
- Logout events
- Invalid username attempts
- Invalid password attempts

### Audit Log Fields

- Timestamp
- Username
- IP Address
- Action (login, logout, etc.)
- Result (success, failure)
- Details (failure reason)

### Viewing Audit Logs

Audit logs are stored in the SQLite database in the `audit_logs` table. Automatic cleanup removes logs older than 90 days.

## Security Headers

### HTTP Headers Applied

**Always:**
- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `Referrer-Policy: no-referrer` - Protects privacy
- `X-XSS-Protection: 1; mode=block` - XSS protection
- `Permissions-Policy` - Restricts browser features

**Content Security Policy:**
```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net;
style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;
img-src 'self' data: https:;
font-src 'self' data: https://cdn.jsdelivr.net;
connect-src 'self'
```

**Production Only:**
- `Strict-Transport-Security: max-age=31536000; includeSubDomains` (HSTS)

## File Permissions

### Automatic Security

The server automatically sets secure permissions on startup:

- **Config file**: `0600` (owner read/write only)
- **Database files**: `0600` (owner read/write only)
- **Backup directory**: `0700` (owner access only)

### Manual Verification

```bash
ls -la ~/.config/cc/
```

Should show:
```
-rw------- config.json
-rw------- data.db
```

## Public vs Protected Endpoints

### Public Endpoints (No Auth Required)

These endpoints remain accessible without authentication:

- `/track` - Event tracking
- `/pixel.gif` - Tracking pixel
- `/r/*` - Redirect service
- `/webhook/*` - Webhook receiver
- `/static/*` - Static assets
- `/login` - Login page
- `/api/login` - Login API
- `/health` - Health check

### Protected Endpoints (Auth Required)

These require valid authentication:

- `/` - Dashboard
- `/api/stats` - Analytics API
- `/api/events` - Events API
- `/api/redirects` - Redirects management
- `/api/webhooks` - Webhooks management
- `/api/domains` - Domains list
- `/api/tags` - Tags list
- `/api/config` - Configuration API
- `/api/logout` - Logout API
- `/api/auth/status` - Auth status

## Production Deployment

### Security Checklist

- [ ] Enable authentication (`--username` and `--password`)
- [ ] Use strong, unique password (12+ characters)
- [ ] Set `env` to `production` in config
- [ ] Run behind HTTPS reverse proxy (nginx, Caddy)
- [ ] Verify file permissions (0600 for config/db)
- [ ] Regular audit log review
- [ ] Keep server updated
- [ ] Use firewall to restrict access
- [ ] Monitor for brute-force attempts
- [ ] Regular database backups

### Reverse Proxy Configuration

#### Nginx Example

```nginx
server {
    listen 443 ssl http2;
    server_name cc.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:4698;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Reporting Security Issues

If you discover a security vulnerability, please email security@toolbomber.com with:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

**Please do not create public GitHub issues for security vulnerabilities.**

## Security Best Practices

1. **Never commit** `config.json` to version control
2. **Use environment-specific configs** for different deployments
3. **Rotate passwords** periodically
4. **Monitor audit logs** for suspicious activity
5. **Keep backups** of your database
6. **Use HTTPS** in production
7. **Restrict network access** with firewall rules
8. **Update regularly** to get security patches
9. **Use strong passwords** (not default examples)
10. **Disable auth only** in trusted, isolated environments
