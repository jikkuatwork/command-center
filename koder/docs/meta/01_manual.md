## Deployment Instructions (Manual - After Build)

**On your server (SSH):**

```bash
# 1. Create directory
sudo mkdir -p /opt/command-center
cd /opt/command-center

# 2. Upload release (from local machine)
# scp command-center-v0.1.0.tar.gz user@server:/opt/command-center/

# 3. Extract
tar -xzf command-center-v0.1.0.tar.gz

# 4. Create .env file
cp .env.example .env
nano .env  # Edit with your settings

# 5. Make binary executable
chmod +x cc-server

# 6. Create systemd service
sudo nano /etc/systemd/system/command-center.service
# (paste service file content from Phase 20)

# 7. Start service
sudo systemctl daemon-reload
sudo systemctl enable command-center
sudo systemctl start command-center
sudo systemctl status command-center

# 8. Configure nginx (if not already done)
sudo nano /etc/nginx/sites-available/cc.toolbomber.com
# (paste nginx config from Phase 20)
sudo ln -s /etc/nginx/sites-available/cc.toolbomber.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 9. Test
curl http://localhost:4698/api/stats
# Should return JSON

# 10. Visit https://cc.toolbomber.com
```
