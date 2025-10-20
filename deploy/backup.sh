#!/bin/bash
set -eo pipefail

# Load configuration
source "${SCRIPT_DIR}/deploy.conf"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="${BACKUP_ROOT}/${TIMESTAMP}"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Check if there's anything to backup
if [[ ! -d "$FRONTEND_DIR" ]] && [[ ! -d "$BACKEND_DIR" ]]; then
    echo "Error: No deployment found to backup" >&2
    exit 1
fi

echo "Creating backup: $TIMESTAMP"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup metadata
cat > "$BACKUP_DIR/metadata.txt" <<EOF
Backup Created: $(date)
Hostname: $(hostname)
Backend Type: $(if [[ -f "$BACKEND_DIR/.backend_type" ]]; then cat "$BACKEND_DIR/.backend_type"; else echo "unknown"; fi)
EOF

# Backup frontend
if [[ -d "$FRONTEND_DIR" ]]; then
    echo "Backing up frontend..."
    tar -czf "$BACKUP_DIR/frontend.tar.gz" -C "$(dirname "$FRONTEND_DIR")" "$(basename "$FRONTEND_DIR")" 2>/dev/null || {
        echo "Warning: Frontend backup failed" >&2
    }
fi

# Backup backend
if [[ -d "$BACKEND_DIR" ]]; then
    echo "Backing up backend..."
    tar -czf "$BACKUP_DIR/backend.tar.gz" -C "$(dirname "$BACKEND_DIR")" "$(basename "$BACKEND_DIR")" 2>/dev/null || {
        echo "Warning: Backend backup failed" >&2
    }
fi

# Backup systemd service
if [[ -f /etc/systemd/system/pebble-backend.service ]]; then
    echo "Backing up systemd service..."
    cp /etc/systemd/system/pebble-backend.service "$BACKUP_DIR/"
fi

# Calculate backup size
BACKUP_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
echo "Backup size: $BACKUP_SIZE"

# Create symlink to latest backup
ln -sfn "$BACKUP_DIR" "${BACKUP_ROOT}/latest"

# Cleanup old backups
echo "Cleaning up old backups (keeping last $MAX_BACKUPS)..."
cd "$BACKUP_ROOT"
ls -t | grep -E '^[0-9]{8}_[0-9]{6}$' | tail -n +$((MAX_BACKUPS + 1)) | while read -r old_backup; do
    echo "Removing old backup: $old_backup"
    rm -rf "$old_backup"
done

echo ""
echo "=== Backup completed successfully ==="
echo "Backup ID: $TIMESTAMP"
echo "Location: $BACKUP_DIR"
echo "Size: $BACKUP_SIZE"
echo ""
echo "To rollback to this backup, run:"
echo "  make rollback BACKUP_ID=$TIMESTAMP"