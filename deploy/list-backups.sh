#!/bin/bash
set -eo pipefail

# Load configuration
source "${SCRIPT_DIR}/deploy.conf"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Check if backup directory exists
if [[ ! -d "$BACKUP_ROOT" ]]; then
    echo "No backups found"
    exit 0
fi

# List backups
cd "$BACKUP_ROOT"
BACKUPS=$(ls -t | grep -E '^[0-9]{8}_[0-9]{6}$' || true)

if [[ -z "$BACKUPS" ]]; then
    echo "No backups found"
    exit 0
fi

echo "ID               Created              Size      Backend   Files"
echo "==============================================================================="

while IFS= read -r backup_id; do
    if [[ ! -d "$backup_id" ]]; then
        continue
    fi
    
    # Parse timestamp
    year=${backup_id:0:4}
    month=${backup_id:4:2}
    day=${backup_id:6:2}
    hour=${backup_id:9:2}
    minute=${backup_id:11:2}
    second=${backup_id:13:2}
    created="${year}-${month}-${day} ${hour}:${minute}:${second}"
    
    # Get size
    size=$(du -sh "$backup_id" 2>/dev/null | cut -f1)
    
    # Get backend type
    backend="unknown"
    if [[ -f "$backup_id/metadata.txt" ]]; then
        backend=$(grep "Backend Type:" "$backup_id/metadata.txt" | cut -d: -f2 | xargs)
    fi
    
    # Count files
    files=""
    if [[ -f "$backup_id/frontend.tar.gz" ]]; then
        files="${files}F"
    else
        files="${files}-"
    fi
    if [[ -f "$backup_id/backend.tar.gz" ]]; then
        files="${files}B"
    else
        files="${files}-"
    fi
    if [[ -f "$backup_id/pebble-backend.service" ]]; then
        files="${files}S"
    else
        files="${files}-"
    fi
    
    # Mark latest
    marker=""
    if [[ -L "latest" ]] && [[ "$(readlink latest)" == "$backup_id" ]]; then
        marker=" *"
    fi
    
    printf "%-16s %-20s %-9s %-9s %-5s%s\n" "$backup_id" "$created" "$size" "$backend" "$files" "$marker"
done <<< "$BACKUPS"

echo ""
echo "Legend: F=Frontend, B=Backend, S=Service, * = Latest backup"
echo ""
echo "To rollback to a backup, run:"
echo "  make rollback BACKUP_ID=<backup_id>"
echo "Or rollback to the latest:"
echo "  make rollback"