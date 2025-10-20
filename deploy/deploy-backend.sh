#!/bin/bash
set -eo pipefail

BACKEND="${BACKEND:?BACKEND environment variable must be set (go|py|kt|rs)}"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
DEPLOY_DIR="/opt/pebble"
API_PORT="${API_PORT:-8000}"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Stop existing service
if systemctl is-active --quiet pebble-backend; then
    echo "Stopping existing backend service..."
    systemctl stop pebble-backend
fi

# Deploy based on backend choice
case "$BACKEND" in
    go)
        echo "Deploying Go backend..."
        "$SCRIPT_DIR/backends/deploy-go.sh"
        ;;
    py)
        echo "Deploying Python backend..."
        "$SCRIPT_DIR/backends/deploy-py.sh"
        ;;
    kt)
        echo "Deploying Kotlin backend..."
        "$SCRIPT_DIR/backends/deploy-kt.sh"
        ;;
    rs)
        echo "Deploying Rust backend..."
        "$SCRIPT_DIR/backends/deploy-rs.sh"
        ;;
    *)
        echo "Error: Invalid backend: $BACKEND (must be go|py|kt|rs)" >&2
        exit 1
        ;;
esac

# Create systemd service
echo "Creating systemd service..."
cat > /etc/systemd/system/pebble-backend.service <<EOF
[Unit]
Description=Pebble Backend Service ($BACKEND)
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=$DEPLOY_DIR/backend
Environment="PORT=$API_PORT"
ExecStart=$DEPLOY_DIR/backend/start.sh
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
systemctl daemon-reload
systemctl enable pebble-backend
systemctl start pebble-backend

# Wait for service to start
echo "Waiting for service to start..."
sleep 2

# Check service status
if systemctl is-active --quiet pebble-backend; then
    echo "Backend service started successfully"
    echo "Checking health..."
    if curl -f http://localhost:$API_PORT/health &>/dev/null || \
       curl -f http://localhost:$API_PORT/api/health &>/dev/null; then
        echo "Health check passed"
    else
        echo "Warning: Health check endpoint not responding (this may be normal if no health endpoint exists)"
    fi
else
    echo "Error: Backend service failed to start" >&2
    journalctl -u pebble-backend -n 20 --no-pager
    exit 1
fi

echo "Backend ($BACKEND) deployed successfully on port $API_PORT"