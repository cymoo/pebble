#!/bin/bash

set -eo pipefail

# Check root privileges
if (( EUID != 0 )); then
    echo >&2 "Error: This script requires root privileges. Use sudo."
    exit 1
fi

# Configuration parameters
SERVICE_NAME="mote.service"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}"
DIRECTORIES=(
    "/etc/mote"
    "/var/www/mote/releases"
)

# Service management functions
stop_service_if_active() {
    if systemctl is-active --quiet "$SERVICE_NAME" &>/dev/null; then
        echo "Stopping ${SERVICE_NAME}..."
        systemctl stop "$SERVICE_NAME"
    fi
}

disable_service_if_enabled() {
    if systemctl is-enabled --quiet "$SERVICE_NAME" &>/dev/null; then
        echo "Disabling ${SERVICE_NAME}..."
        systemctl disable "$SERVICE_NAME"
    fi
}

# File system cleanup functions
remove_service_file() {
    if [[ -f "$SERVICE_FILE" ]]; then
        echo "Removing service file: ${SERVICE_FILE}"
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload
        systemctl reset-failed
    fi
}

clean_directories() {
    for dir in "${DIRECTORIES[@]}"; do
        if [[ -d "$dir" ]]; then
            echo "Removing directory: ${dir}"
            rm -rf "$dir"
        fi
    done
}

stop_service_if_active
disable_service_if_enabled
remove_service_file
clean_directories
rm -f /var/www/mote/current

echo "Cleanup completed successfully."
