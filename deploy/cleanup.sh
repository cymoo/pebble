#!/bin/bash
set -eo pipefail

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

echo "=== Cleaning up Pebble deployment ==="

# Stop and disable backend service
if systemctl is-active --quiet pebble-backend; then
    echo "Stopping pebble-backend service..."
    systemctl stop pebble-backend
fi

if systemctl is-enabled --quiet pebble-backend 2>/dev/null; then
    echo "Disabling pebble-backend service..."
    systemctl disable pebble-backend
fi

# Remove systemd service file
if [[ -f /etc/systemd/system/pebble-backend.service ]]; then
    echo "Removing systemd service file..."
    rm -f /etc/systemd/system/pebble-backend.service
    systemctl daemon-reload
fi

# Remove deployment directory
if [[ -d /opt/pebble ]]; then
    echo "Removing deployment directory: /opt/pebble"
    rm -rf /opt/pebble
fi

# Remove web root
if [[ -d /var/www/pebble ]]; then
    echo "Removing web root: /var/www/pebble"
    rm -rf /var/www/pebble
fi

# Remove nginx configuration (optional - ask for confirmation)
read -p "Remove Nginx configuration and SSL certificates? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing Nginx configuration..."
    rm -f /etc/nginx/conf.d/*pebble*.conf
    rm -f /etc/nginx/conf.d/*-temp.conf
    
    read -p "Also revoke SSL certificates? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Note: You'll need to manually revoke certificates with:"
        echo "  certbot revoke --cert-name <domain>"
        echo "  certbot delete --cert-name <domain>"
    fi
    
    nginx -t && systemctl reload nginx
fi

echo ""
echo "=== Cleanup completed ==="
echo "Services stopped and deployment files removed"