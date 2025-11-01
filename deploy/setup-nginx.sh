#!/bin/bash

set -eo pipefail

SERVER_NAME="${SERVER_NAME:?The SERVER_NAME environment variable must be set}"

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx.template"
source "${SCRIPT_DIR}/config.env"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Generate Nginx configuration
log_info "Generating Nginx configuration..."
# shellcheck disable=SC2016
envsubst '$WEB_DIR $UPLOADS_DIR $SERVER_NAME $BACKEND_ADDR $BACKEND_PORT $MEMO_URL $BLOG_URL' < "$NGINX_TEMPLATE" > /etc/nginx/conf.d/"${SERVER_NAME}".conf

# Clean up temporary files for certbot validation if existed
if [ -f /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf ]; then
    log_info "Removing temporary Nginx configuration for certbot validation..."
    rm -f /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf
fi

nginx -t && systemctl reload nginx

log_success -e "\n=== Nginx configuration completed successfully ==="
