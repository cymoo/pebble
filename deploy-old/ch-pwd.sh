#!/bin/bash

set -eo pipefail

if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as the root user or with sudo" >&2
   exit 1
fi

PASSWORD_FILE="/etc/pebble/secure"

# Generate a 32-character random password
password=$(tr -dc 'a-zA-Z0-9_-' < /dev/urandom | fold -w 32 | head -n 1 || true)

echo "The generated random password is: $password"

# Create parent directory with sudo
sudo mkdir -p "$(dirname "${PASSWORD_FILE}")"

# Write password to file securely
echo "PEBBLE_PASSWORD=${password}" | sudo tee "${PASSWORD_FILE}" > /dev/null || {
    echo "Error: Failed to write password to file"
    exit 1
}

# Set secure file permissions (rw- --- ---)
sudo chmod 600 "${PASSWORD_FILE}"

echo "Success: Password securely stored in ${PASSWORD_FILE}"

echo "Restarting service..."
sudo systemctl restart pebble.service
