#!/bin/bash

set -eo pipefail

# ==================== Configuration Section ====================
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WWW_ROOT="/var/www/mote"
CURRENT_LINK="${WWW_ROOT}/current"
RELEASES_DIR="${WWW_ROOT}/releases"
MAX_KEEP=5
BUILD_TEMP=$(mktemp -d)

PASSWORD_FILE="/etc/mote/secure"
DATABASE_FILE="${WWW_ROOT}/app.db"

export MEMO_URL="/memo"
export BLOG_URL="/shared"

export DATABASE_URL=sqlite://${DATABASE_FILE}
export REDIS_URL=redis://localhost
export UPLOAD_DIR="${WWW_ROOT}/uploads"
export RUST_LOG=warn
export API_IP="127.0.0.1"
export API_PORT=8000

# ==================== Initialization Checks ====================
command -v npm &> /dev/null || { echo >&2 "npm required"; exit 1; }
command -v cargo &> /dev/null || { echo >&2 "cargo required"; exit 1; }

if [ ! -d "${RELEASES_DIR}" ]; then
    sudo mkdir -p "${RELEASES_DIR}"
    sudo chown "$(whoami)": "${WWW_ROOT}"
    sudo chown "$(whoami)": "${RELEASES_DIR}"
fi

# Generate the login password when it does not exist.
if [ ! -f "${PASSWORD_FILE}" ]; then
    # Generate a 32-character random password
    password=$(tr -dc 'a-zA-Z0-9_-' < /dev/urandom | fold -w 32 | head -n 1 || true)

    echo "No password found, the newly generated password is: $password"

    # Create parent directory
    sudo mkdir -p "$(dirname "${PASSWORD_FILE}")"

    # Write password to file securely
    echo "MOTE_PASSWORD=${password}" | sudo tee "${PASSWORD_FILE}" > /dev/null || {
        echo "Error: Failed to write password to file"
        exit 1
    }

    # Set secure file permissions (rw- --- ---)
    sudo chmod 600 "${PASSWORD_FILE}"

    echo "Success: Password securely stored in ${PASSWORD_FILE}"
fi

# ==================== Build Phase ====================
echo "[1/4] Building frontend..."
cd "${PROJECT_ROOT}/frontend"
npx yarn install --frozen-lockfile --silent
VITE_MEMO_URL=$MEMO_URL VITE_BLOG_URL=$BLOG_URL npx vite build --logLevel error

echo "[2/4] Building backend..."
cd "${PROJECT_ROOT}/api-rs"

if [ ! -f $DATABASE_FILE ]; then
    if ! command -v sqlx &> /dev/null; then
        echo "Installing sqlx-cli..."
        cargo install sqlx-cli --no-default-features --features sqlite
    fi

    echo "Creating database..."
    sqlx database create
    sqlx migrate run
    chmod 600 $DATABASE_FILE
fi

cargo build --release

# ==================== Prepare Artifacts ====================
echo "[3/4] Preparing release.."
mkdir -p "${BUILD_TEMP}/web-dist" "${BUILD_TEMP}/api-dist"
cp -a "${PROJECT_ROOT}/frontend/dist/." "${BUILD_TEMP}/web-dist/"
cp -a "${PROJECT_ROOT}/api-rs/target/release/mote" \
     "${PROJECT_ROOT}/api-rs/.env" \
     "${PROJECT_ROOT}/api-rs/templates" \
     "${PROJECT_ROOT}/api-rs/static" \
     "${PROJECT_ROOT}/api-rs/migrations" \
     "${BUILD_TEMP}/api-dist/"

# ==================== Deployment Phase ====================
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
RELEASE_DIR="${RELEASES_DIR}/${TIMESTAMP}"
mv -f "${BUILD_TEMP}" "${RELEASE_DIR}"
chmod 755 "$RELEASE_DIR"

OLD_RELEASE=$(readlink -f "${CURRENT_LINK}" 2>/dev/null || true)

# Atomically switch symlink
ln -sfn "${RELEASE_DIR}" "${CURRENT_LINK}.tmp"
mv -fT "${CURRENT_LINK}.tmp" "${CURRENT_LINK}"

# ==================== Service Management ====================
SERVICE_FILE="/etc/systemd/system/mote.service"

should_restart_api() {
    # First deployment
    [ -z "${OLD_RELEASE}" ] && return 0

    # Compare binary and .env between old and new releases
    ! cmp -s "${OLD_RELEASE}/api-dist/mote" "${RELEASE_DIR}/api-dist/mote" && return 0
    ! cmp -s "${OLD_RELEASE}/api-dist/.env" "${RELEASE_DIR}/api-dist/.env" && return 0

    return 1
}

# Define the new content to be written to the service file
NEW_SERVICE_FILE_CONTENT=$(cat <<EOF
[Unit]
Description=Mote Web Service
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=${CURRENT_LINK}/api-dist
Environment="IP=${API_IP}"
Environment="PORT=${API_PORT}"
Environment="RUST_LOG=${RUST_LOG}"
Environment="REDIS_URL=${REDIS_URL}"
Environment="DATABASE_URL=${DATABASE_URL}"
Environment="UPLOAD_DIR=${UPLOAD_DIR}"
EnvironmentFile=${PASSWORD_FILE}
ExecStart=${CURRENT_LINK}/api-dist/mote
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
)

# Check if the service file needs to be updated
SHOULD_UPDATE_SERVICE_FILE=false
if [ ! -f "${SERVICE_FILE}" ]; then
    SHOULD_UPDATE_SERVICE_FILE=true
elif ! echo "$NEW_SERVICE_FILE_CONTENT" | diff -q "${SERVICE_FILE}" - >/dev/null; then
    SHOULD_UPDATE_SERVICE_FILE=true
fi

# Update the service file if necessary
if $SHOULD_UPDATE_SERVICE_FILE; then
    echo "[4/4] Configuring service.."
    echo "$NEW_SERVICE_FILE_CONTENT" | sudo tee "${SERVICE_FILE}" >/dev/null
    sudo systemctl daemon-reload
fi

# Service state management
if ! systemctl is-active mote.service >/dev/null 2>&1; then
    echo "Starting service..."
    sudo systemctl enable --now mote.service >/dev/null
elif $SHOULD_UPDATE_SERVICE_FILE; then
    echo "Service configuration Changes detected, restarting service..."
    sudo systemctl restart mote.service
elif should_restart_api; then
    echo "Api Changes detected, restarting service..."
    sudo systemctl restart mote.service
else
    echo "No need to restart service"
fi

# ==================== Cleanup ====================
# Keep the latest 5 releases
find "${RELEASES_DIR}" -maxdepth 1 -type d -printf '%T@ %p\n' |
sort -rn | awk -v max=$((MAX_KEEP + 1)) 'NR > max {print $2}' |
xargs -r rm -rf

echo "Deployment completed!"
