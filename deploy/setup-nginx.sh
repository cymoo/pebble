#!/bin/bash
set -eo pipefail

DOMAIN="${DOMAIN:?The DOMAIN environment variable must be set}"
EMAIL="${EMAIL:?The EMAIL environment variable must be set}"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
NGINX_TEMPLATE="${SCRIPT_DIR}/nginx.template"

export WWW_ROOT=/var/www/pebble
export SERVER_NAME="$DOMAIN"
export API_PORT=8000

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "Error: This script must be run as root" >&2
    exit 1
fi

# Install dependencies
install_dependencies() {
    echo "Installing Nginx and Certbot..."
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
    if ! crontab -l 2>/dev/null | grep -qF "certbot renew"; then
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
    
    # Generate temporary configuration for cert validation
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
    if [[ -f "$NGINX_TEMPLATE" ]]; then
        # shellcheck disable=SC2016
        envsubst '$WWW_ROOT $SERVER_NAME $API_PORT' < "$NGINX_TEMPLATE" > /etc/nginx/conf.d/"${DOMAIN}".conf
    else
        # Create default configuration if template doesn't exist
        cat > /etc/nginx/conf.d/"${DOMAIN}".conf <<'ENDCONF'
# HTTP redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${SERVER_NAME};
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${SERVER_NAME};

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/${SERVER_NAME}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${SERVER_NAME}/privkey.pem;
    ssl_dhparam /etc/nginx/ssl/dhparam.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-CHACHA20-POLY1305;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Root directory for frontend
    root ${WWW_ROOT};
    index index.html;

    # Frontend routing
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://localhost:${API_PORT}/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 90s;
    }

    # Static assets caching
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
ENDCONF
        # Substitute variables
        sed -i "s|\${SERVER_NAME}|${SERVER_NAME}|g" /etc/nginx/conf.d/"${DOMAIN}".conf
        sed -i "s|\${WWW_ROOT}|${WWW_ROOT}|g" /etc/nginx/conf.d/"${DOMAIN}".conf
        sed -i "s|\${API_PORT}|${API_PORT}|g" /etc/nginx/conf.d/"${DOMAIN}".conf
    fi
    
    # Clean up and reload
    rm -f /etc/nginx/conf.d/"${DOMAIN}"-temp.conf
    nginx -t && systemctl reload nginx
    
    # Configure automatic renewal
    setup_renewal
    
    echo -e "\n=== HTTPS configuration completed successfully ==="
    echo "Visit: https://${DOMAIN}"
}

main "$@"