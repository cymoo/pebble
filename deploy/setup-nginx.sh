#!/bin/bash

set -eo pipefail

# SERVER_NAME="${SERVER_NAME:?The SERVER_NAME environment variable must be set}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

export SERVER_NAME="$1"

if [ -z "$SERVER_NAME" ]; then
    log_error "请指定服务器域名"
    exit 1
fi

NGINX_TEMPLATE="${SCRIPT_DIR}/nginx.template"

NGINX_CONFIG="${CONFIG_DIR}/nginx/${APP_NAME}.conf"
NGINX_ENABLED="/etc/nginx/sites-enabled/${APP_NAME}.conf"
NGINX_AVAILABLE="/etc/nginx/sites-available/${APP_NAME}.conf"

# Generate Nginx configuration
log_info "Generating Nginx configuration..."

# 检查模板文件
if [ ! -f "$NGINX_TEMPLATE" ]; then
    log_error "Nginx模板文件不存在: $NGINX_TEMPLATE"
    exit 1
fi

envsubst '$WEB_DIR $UPLOADS_DIR $SERVER_NAME $BACKEND_ADDR $BACKEND_PORT $MEMO_URL $BLOG_URL' < "$NGINX_TEMPLATE" | sudo tee "$NGINX_CONFIG" > /dev/null

# 复制到nginx配置目录
sudo cp "$NGINX_CONFIG" "$NGINX_AVAILABLE"

# 创建软链接
if [ -L "$NGINX_ENABLED" ]; then
    sudo rm "$NGINX_ENABLED"
fi
sudo ln -s "$NGINX_AVAILABLE" "$NGINX_ENABLED"

# 删除默认配置
if [ -f "/etc/nginx/sites-enabled/default" ]; then
    log_info "移除Nginx默认配置..."
    sudo rm -f /etc/nginx/sites-enabled/default
fi

# 测试配置
log_info "测试Nginx配置..."
if sudo nginx -t; then
    log_success "Nginx配置有效"

    # 重载Nginx
    if sudo systemctl is-active --quiet nginx; then
        log_info "重载Nginx..."
        sudo systemctl reload nginx
    else
        log_info "启动Nginx..."
        sudo systemctl enable nginx
        sudo systemctl start nginx
    fi
else
    log_error "Nginx配置测试失败!"
    exit 1
fi

log_success "Nginx配置完成!"
log_info "配置文件: $NGINX_CONFIG"
log_info "软链接: $NGINX_ENABLED"


# Clean up temporary files for certbot validation if existed
if [ -f /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf ]; then
    log_info "Removing temporary Nginx configuration for certbot validation..."
    rm -f /etc/nginx/conf.d/"${SERVER_NAME}"-temp.conf
fi

# nginx -t && systemctl reload nginx

# log_success -e "\n=== Nginx configuration completed successfully ==="
