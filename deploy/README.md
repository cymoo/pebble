# Deployment Guide

This deployment uses the Rust Axum as the backend and HTTPS is configured.

After deployment:

- `/memo`   ->  Memo homepage
- `/shared` ->  Blog homepage
- `/`       ->  Blog homepage (redirect)

## Directory Structure

```
# Deployment directory structure
/var/www/mote/
├── app.db                     # SQLite database file
├── uploads/                   # Directory for uploaded images
├── releases/                  # Stores the last 5 build versions
│   └── 2025-03-16_18-21-44/   # Example of build version directory
│       ├── api-dist/          # Rust API build artifacts
│       │   ├── .env             # Environment variables
│       │   ├── static/          # Static resources
│       │   ├── templates/       # HTML templates
│       │   └── mote           # API binary
│       └── web-dist/          # Frontend build artifacts
│           ├── index.html       # Entry HTML
│           └── assets/          # Compiled JS/CSS/images and other resources
└── current -> releases/2025-03-16_18-21-44  # Symlink pointing to the latest version

# System configuration file paths
/etc/
├── mote/
│   └── secure                  # Stores passwords
└── systemd/system/
    └── mote.service
```

## Environment Requirements

- Linux server (I used Ubuntu 24.04, other systemd-based versions or distributions are also OK, but you may need to modify the package manager commands in the scripts)
- Account with sudo privileges
- Domain name with DNS resolution configured
- Ports: 80 and 443

## Deployment

1. Install basic dependencies: Rust Toolchain, Node.js, and Redis. If they already exist, you can skip this step.

```bash
./init-env.sh
```

2. Build and run the backend service (a random password for login will be generated on the first run). Repeat this script if the code has been modified.

```bash
./start-service.sh
```

3. Configure HTTPS (apply for SSL certificate and set up auto-renewal) and Nginx. Run this script once for the first time. Repeat if the domain or Nginx template file is modified.

```bash
sudo DOMAIN=your.domain EMAIL=your@email.addr ./setup-nginx.sh
```

## Others

1. Change login password: `sudo ./ch-pwd.sh`

2. Stop the service and clean up build history: `sudo ./del-service.sh`. Note: It **will not** delete the database file and upload directory.
