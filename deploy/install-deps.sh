#!/bin/bash
set -eo pipefail

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

echo "Updating package repositories..."
apt-get update -q

echo "Installing common dependencies..."
apt-get install -q -y \
    curl \
    wget \
    git \
    build-essential \
    openssl \
    ca-certificates \
    gnupg \
    lsb-release \
    > /dev/null

echo "Installing Nginx..."
apt-get install -q -y nginx > /dev/null

echo "Installing Node.js and npm..."
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
    apt-get install -q -y nodejs > /dev/null
fi

# Verify installations
echo ""
echo "=== Installed versions ==="
echo "Node.js: $(node --version)"
echo "npm: $(npm --version)"
echo "Nginx: $(nginx -v 2>&1)"
echo ""
echo "System dependencies installed successfully"