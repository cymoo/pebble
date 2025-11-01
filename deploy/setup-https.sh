
#!/bin/bash

set -eo pipefail

SERVER_NAME="${SERVER_NAME:?The SERVER_NAME environment variable must be set}"
EMAIL="${EMAIL:?The EMAIL environment variable must be set}"

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx.template"
source "${SCRIPT_DIR}/config.env"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

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

# Configure automatic cert renewal
setup_renewal() {
    local cron_time="0 3 * * *"
    local cron_cmd="/usr/bin/certbot renew --quiet --post-hook 'systemctl reload nginx'"

    if ! crontab -l | grep -qF "certbot renew"; then
        (crontab -l 2>/dev/null; echo "${cron_time} ${cron_cmd}") | crontab -
        log_success "Configured daily automatic renewal task at 3 AM"
    fi
}

# Main process
main() {
    install_dependencies
    generate_dhparam

    # Create certificate validation directory
    mkdir -p /var/www/certbot
    chmod -R 755 /var/www/certbot

    # Generate temporary configuration
    log_info "Configuring temporary Nginx server..."
    cat > /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $SERVER_NAME;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 444;
    }
}
EOF

    nginx -t && systemctl reload nginx

    # Request certificate
    log_info "Requesting SSL certificate..."
    certbot certonly --webroot -w /var/www/certbot \
        -d "$SERVER_NAME" \
        --email "$EMAIL" \
        --agree-tos \
        --non-interactive \
        --keep-until-expiring

    # Configure automatic renewal
    setup_renewal

    log_success -e "\n=== HTTPS configuration completed successfully ==="
    log_success "Visit: https://${SERVER_NAME}"
}

main "$@"
