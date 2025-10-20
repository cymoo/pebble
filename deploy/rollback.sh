#!/bin/bash
set -eo pipefail

BACKUP_ID="${BACKUP_ID:-}"

# Load configuration
source "${SCRIPT_DIR}/deploy.conf"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Determine which backup to use
if [[ -z "$BACKUP_ID" ]]; then
    # Use latest backup
    if [[ ! -L "$BACKUP_ROOT/latest" ]]; then
        echo "Error: No backups found" >&2
        exit 1
    fi
    BACKUP_DIR=$(readlink -f "$BACKUP_ROOT/latest")
    BACKUP_ID=$(basename "$BACKUP_DIR")
    echo "Using latest backup: $BACKUP_ID"
else
    BACKUP_DIR="$BACKUP_ROOT/$BACKUP_ID"
    if [[ ! -d "$BACKUP_DIR" ]]; then
        echo "Error: Backup not found: $BACKUP_ID" >&2
        echo "Run 'make list-backups' to see available backups" >&2
        exit 1
    fi
fi

echo "=== Rolling back to backup: $BACKUP_ID ==="
echo ""

# Show backup info
if [[ -f "$BACKUP_DIR/metadata.txt" ]]; then
    cat "$BACKUP_DIR/metadata.txt"
    echo ""
fi

# Confirm rollback
read -p "Are you sure you want to rollback? This will replace current deployment (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Rollback cancelled"
    exit 0
fi

# Create a backup of current state before rollback
echo "Creating safety backup of current state..."
SAFETY_BACKUP_DIR="${BACKUP_ROOT}/pre-rollback-$(date +%Y%m%d_%H%M%S)"
mkdir -p "$SAFETY_BACKUP_DIR"

if [[ -d "$FRONTEND_DIR" ]]; then
    tar -czf "$SAFETY_BACKUP_DIR/frontend.tar.gz" -C "$(dirname "$FRONTEND_DIR")" "$(basename "$FRONTEND_DIR")" 2>/dev/null || true
fi

if [[ -d "$BACKEND_DIR" ]]; then
    tar -czf "$SAFETY_BACKUP_DIR/backend.tar.gz" -C "$(dirname "$BACKEND_DIR")" "$(basename "$BACKEND_DIR")" 2>/dev/null || true
fi

echo "Safety backup created at: $SAFETY_BACKUP_DIR"
echo ""

# Stop backend service
if systemctl is-active --quiet pebble-backend; then
    echo "Stopping backend service..."
    systemctl stop pebble-backend
fi

# Restore frontend
if [[ -f "$BACKUP_DIR/frontend.tar.gz" ]]; then
    echo "Restoring frontend..."
    rm -rf "$FRONTEND_DIR"
    mkdir -p "$(dirname "$FRONTEND_DIR")"
    tar -xzf "$BACKUP_DIR/frontend.tar.gz" -C "$(dirname "$FRONTEND_DIR")"
    chown -R www-data:www-data "$FRONTEND_DIR"
    echo "Frontend restored"
else
    echo "Warning: No frontend backup found in this backup"
fi

# Restore backend
if [[ -f "$BACKUP_DIR/backend.tar.gz" ]]; then
    echo "Restoring backend..."
    rm -rf "$BACKEND_DIR"
    mkdir -p "$(dirname "$BACKEND_DIR")"
    tar -xzf "$BACKUP_DIR/backend.tar.gz" -C "$(dirname "$BACKEND_DIR")"
    chown -R www-data:www-data "$BACKEND_DIR"
    
    # Ensure start script is executable
    if [[ -f "$BACKEND_DIR/start.sh" ]]; then
        chmod +x "$BACKEND_DIR/start.sh"
    fi
    
    echo "Backend restored"
else
    echo "Warning: No backend backup found in this backup"
fi

# Restore systemd service
if [[ -f "$BACKUP_DIR/pebble-backend.service" ]]; then
    echo "Restoring systemd service..."
    cp "$BACKUP_DIR/pebble-backend.service" /etc/systemd/system/
    systemctl daemon-reload
    echo "Systemd service restored"
fi

# Start backend service
if [[ -f /etc/systemd/system/pebble-backend.service ]]; then
    echo "Starting backend service..."
    systemctl start pebble-backend
    
    # Wait and check
    sleep 2
    if systemctl is-active --quiet pebble-backend; then
        echo "Backend service started successfully"
    else
        echo "Warning: Backend service failed to start" >&2
        echo "Check logs with: journalctl -u pebble-backend -n 50" >&2
    fi
fi

# Reload nginx
if systemctl is-active --quiet nginx; then
    echo "Reloading Nginx..."
    systemctl reload nginx
fi

echo ""
echo "=== Rollback completed ==="
echo "Restored from backup: $BACKUP_ID"
echo "Safety backup location: $SAFETY_BACKUP_DIR"
echo ""
echo "Check service status with: make status"
echo "View logs with: make logs"