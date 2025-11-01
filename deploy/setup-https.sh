#!/bin/bash

set -eo pipefail

SERVER_NAME="${SERVER_NAME:?The SERVER_NAME environment variable must be set}"
EMAIL="${EMAIL:?The EMAIL environment variable must be set}"

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
source "${SCRIPT_DIR}/config.env"

WEBROOT_PATH="/var/www/certbot"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Ensure webroot directory exists
ensure_webroot() {
    log_info "Ensuring webroot directory exists..."
    mkdir -p "$WEBROOT_PATH"
    chmod -R 755 "$WEBROOT_PATH"
    log_info "Webroot directory ready: $WEBROOT_PATH"
}

# Install dependencies
install_dependencies() {
    log_info "Installing system dependencies..."
    apt-get update -q
    apt-get install -q -y nginx certbot python3-certbot-nginx openssl > /dev/null || {
        log_error "Failed to install dependencies" >&2
        exit 1
    }
}

# Generate DH parameters
generate_dhparam() {
    local dh_file="/etc/nginx/ssl/dhparam.pem"
    mkdir -p /etc/nginx/ssl

    if [[ ! -f "$dh_file" ]]; then
        log_info "Generating Diffie-Hellman parameters (this may take 1-3 minutes)..."
        openssl dhparam -out "$dh_file" 2048
        chmod 600 "$dh_file"
        log_info "DH parameters generated: $dh_file"
    else
        log_info "Existing DH parameter file detected, skipping generation"
    fi
}

# Setup temporary Nginx config for certificate validation
setup_temp_nginx_config() {
    log_info "Setting up temporary Nginx configuration..."

    # Create temporary Nginx config
    cat > /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $SERVER_NAME;

    location /.well-known/acme-challenge/ {
        root $WEBROOT_PATH;
    }

    location / {
        return 444;
    }
}
EOF

    # Test and reload Nginx
    if nginx -t 2>/dev/null; then
        systemctl reload nginx
        log_info "Temporary Nginx configuration loaded"
    else
        log_error "Nginx configuration test failed"
        exit 1
    fi
}

# Remove temporary Nginx config
cleanup_temp_config() {
    log_info "Removing temporary Nginx configuration..."
    rm -f /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf
    nginx -t 2>/dev/null && systemctl reload nginx
}

# Main process
main() {
    ensure_webroot
    install_dependencies
    generate_dhparam
    setup_temp_nginx_config

    # Request certificate using webroot mode
    log_info "Requesting SSL certificate for ${SERVER_NAME} using webroot mode..."
    if certbot certonly --webroot \
        -w "$WEBROOT_PATH" \
        -d "$SERVER_NAME" \
        --email "$EMAIL" \
        --agree-tos \
        --non-interactive \
        --keep-until-expiring; then
        log_success "SSL certificate obtained successfully"
    else
        log_error "Failed to obtain SSL certificate"
        cleanup_temp_config
        exit 1
    fi

    # Clean up temporary config
    cleanup_temp_config

    log_info "Certificate location: /etc/letsencrypt/live/${SERVER_NAME}/"
}

main "$@"
