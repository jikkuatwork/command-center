# Hosting on DigitalOcean ($5 Droplet) 🚀

This guide assumes you have a **Domain Name** (e.g., `example.com`) and a **DigitalOcean Account**.

## 1. Create Droplet 💧
*   **Region**: Closest to you (e.g., NYC1, BLR1).
*   **Image**: Ubuntu 24.04 LTS (or latest).
*   **Size**: Basic, Regular, $5/mo (512MB RAM, 1 vCPU).
*   **Auth**: SSH Key (Recommended).
*   **Hostname**: `fazt-server` (or whatever you like).

## 2. DNS Setup (Namecheap) 🌐
Point your domain to your Droplet's IP.

*   **A Record**: `@` -> `YOUR_DROPLET_IP`
*   **A Record**: `*` -> `YOUR_DROPLET_IP` (Wildcard for subdomains)

*Wait a few minutes for propagation.*

## 3. Server Setup (SSH in) 💻

```bash
ssh root@YOUR_DROPLET_IP
```

### Prepare Directory
```bash
# Create user (optional but safer)
# useradd -m -s /bin/bash fazt
# su - fazt

mkdir -p ~/fazt
cd ~/fazt
```

### Upload Binary
From your local machine:
```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o fazt ./cmd/server

# Upload
scp fazt root@YOUR_DROPLET_IP:~/fazt/
```

## 4. Initialize & Configure ⚙️

```bash
# Initialize (Generates config.json & data.db)
./fazt server init \
  --username admin \
  --password SUPER_SECURE_PASS \
  --domain https://example.com \
  --env production

# Enable HTTPS (Let's Encrypt)
# Edit config to enable HTTPS
nano ~/.config/fazt/config.json
```

Change `"https"` section:
```json
"https": {
  "enabled": true,
  "email": "you@example.com",
  "staging": false
}
```

## 5. Systemd Service (Auto-Start) 🔄

Create service file:
`nano /etc/systemd/system/fazt.service`

```ini
[Unit]
Description=Fazt PaaS
After=network.target

[Service]
Type=simple
User=root
# Or 'fazt' if you created a user.
# NOTE: Port 80/443 requires root or `setcap`.
# To run as non-root on low ports:
# sudo setcap CAP_NET_BIND_SERVICE=+eip ~/fazt/fazt

WorkingDirectory=/root/fazt
ExecStart=/root/fazt/fazt server start
Restart=always
RestartSec=5

# Environment variables if needed
# Environment=FAZT_DOMAIN=https://example.com

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
systemctl daemon-reload
systemctl enable fazt
systemctl start fazt
systemctl status fazt
```

## 6. Verify ✅

1.  Visit `https://example.com` -> Should load Dashboard (Login).
2.  Login with your credentials.
3.  Deploy a site locally:
    ```bash
    fazt client deploy --path ./mysite --domain blog --server https://example.com
    ```
4.  Visit `https://blog.example.com` -> Should load your site (with SSL!).

## 7. Firewall (UFW) 🛡️

Secure your server:
```bash
ufw allow OpenSSH
ufw allow 80
ufw allow 443
ufw enable
```

---
**Enjoy your Personal Cloud!** ☁️

```