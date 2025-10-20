#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
FRONTEND_SRC="$PROJECT_ROOT/pebble"
WWW_ROOT="/var/www/pebble"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Validate source directory
if [[ ! -d "$FRONTEND_SRC" ]]; then
    echo "Error: Frontend source directory not found: $FRONTEND_SRC" >&2
    exit 1
fi

echo "Building frontend from: $FRONTEND_SRC"

# Install dependencies
cd "$FRONTEND_SRC"
echo "Installing npm dependencies..."
npm ci --production=false

# Build frontend
echo "Building production bundle..."
npm run build

# Deploy to web root
echo "Deploying to: $WWW_ROOT"
mkdir -p "$WWW_ROOT"
rm -rf "${WWW_ROOT:?}"/*

# Copy build output (adjust based on your vite config)
if [[ -d "$FRONTEND_SRC/dist" ]]; then
    cp -r "$FRONTEND_SRC/dist"/* "$WWW_ROOT/"
elif [[ -d "$FRONTEND_SRC/build" ]]; then
    cp -r "$FRONTEND_SRC/build"/* "$WWW_ROOT/"
else
    echo "Error: Build output directory not found (expected dist/ or build/)" >&2
    exit 1
fi

# Set permissions
chown -R www-data:www-data "$WWW_ROOT"
chmod -R 755 "$WWW_ROOT"

echo "Frontend deployed successfully to $WWW_ROOT"