#!/bin/bash

set -eo pipefail

DOMAIN="${DOMAIN:?The DOMAIN environment variable must be set}"
EMAIL="${EMAIL:?The EMAIL environment variable must be set}"

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx.template"
source "${SCRIPT_DIR}/config.env"

# export WWW_ROOT=/var/www/mote
# export SERVER_NAME="$DOMAIN"
# export API_PORT=8000
# export MEMO_URL="/memo"
# export BLOG_URL="/shared"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Install dependencies
install_dependencies() {
    echo "Installing system dependencies..."
    apt-get update -q
    apt-get install -q -y nginx certbot python3-certbot-nginx openssl > /dev/null || {
        echo "Failed to install dependencies" >&2
        exit 1
  }
}

# Generate DH parameters
generate_dhparam() {
    local dh_file="/etc/nginx/ssl/dhparam.pem"
    mkdir -p /etc/nginx/ssl

    if [[ ! -f "$dh_file" ]]; then
        echo "Generating Diffie-Hellman parameters (this may take 1-3 minutes)..."
        openssl dhparam -out "$dh_file" 2048
        chmod 600 "$dh_file"
        echo "DH parameters generated: $dh_file"
    else
        echo "Existing DH parameter file detected, skipping generation"
    fi
}

# Configure automatic cert renewal
setup_renewal() {
    local cron_time="0 3 * * *"
    local cron_cmd="/usr/bin/certbot renew --quiet --post-hook 'systemctl reload nginx'"

    if ! crontab -l | grep -qF "certbot renew"; then
        (crontab -l 2>/dev/null; echo "${cron_time} ${cron_cmd}") | crontab -
        echo "Configured daily automatic renewal task at 3 AM"
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
    echo "Configuring temporary Nginx server..."
    cat > /etc/nginx/conf.d/"${DOMAIN}"-temp.conf <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

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
    echo "Requesting SSL certificate..."
    certbot certonly --webroot -w /var/www/certbot \
        -d "$DOMAIN" \
        --email "$EMAIL" \
        --agree-tos \
        --non-interactive \
        --keep-until-expiring

    # Generate final configuration
    echo "Generating Nginx configuration..."
    # shellcheck disable=SC2016
    envsubst '$WWW_ROOT $SERVER_NAME $API_PORT $MEMO_URL $BLOG_URL' < "$NGINX_TEMPLATE" > /etc/nginx/conf.d/"${DOMAIN}".conf

    # Clean up and reload
    rm -f /etc/nginx/conf.d/"${DOMAIN}"-temp.conf
    nginx -t && systemctl reload nginx

    # Configure automatic renewal
    setup_renewal

    echo -e "\n=== HTTPS configuration completed successfully ==="
    echo "Visit: https://${DOMAIN}"
}

main "$@"
