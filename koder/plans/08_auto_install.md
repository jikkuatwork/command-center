# Plan: Auto-Provisioning (Server Install)

**Status:** Planned
**Target Version:** v0.6.0
**Goal:** Make `fazt` capable of installing itself as a system service on a fresh Linux server.

## Overview
The `fazt server install` command will automate the setup of the application on a production server (specifically targeting Ubuntu/Debian/Systemd based systems). It aims to reduce the deployment manual from ~10 steps to 1 step.

## Proposed Command
```bash
sudo ./fazt server install \
  --domain https://example.com \
  --email admin@example.com \
  --user fazt \
  --https
```

## Internal Logic

1.  **Pre-flight Checks**:
    *   Check for `root` privileges (required for systemd/user creation).
    *   Check for `systemd`.
    *   Check if `fazt` user exists (if not, create it).

2.  **Binary Installation**:
    *   Copy `os.Executable()` to `/usr/local/bin/fazt`.
    *   `chmod +x /usr/local/bin/fazt`.

3.  **Permissions (Non-Root Binding)**:
    *   Execute `setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/fazt`.
    *   *Why?* Allows binding to :80/:443 without running as root user.

4.  **Configuration**:
    *   Run equivalent of `server init` logic for the target user (`/home/fazt/.config/fazt/`).
    *   Set HTTPS config if requested.

5.  **Service Definition**:
    *   Generate unit file `/etc/systemd/system/fazt.service`:
        ```ini
        [Unit]
        Description=Fazt PaaS
        After=network.target

        [Service]
        Type=simple
        User=fazt
        ExecStart=/usr/local/bin/fazt server start
        Restart=always
        LimitNOFILE=4096

        [Install]
        WantedBy=multi-user.target
        ```

6.  **Finalize**:
    *   `systemctl daemon-reload`.
    *   `systemctl enable fazt`.
    *   `systemctl start fazt`.
    *   Check status and print URL.

## Security Considerations
-   **Least Privilege**: Service runs as `fazt`, not `root`.
-   **File Permissions**: Config/DB locked to `fazt` user (0600).

## Roadmap
1.  Implement `internal/provision/systemd.go` (Service generation).
2.  Implement `internal/provision/user.go` (User creation).
3.  Add `server install` subcommand in `cmd/server/main.go`.

```