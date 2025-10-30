#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

log_info "生成密码文件..."

PASSWORD_FILE="${CONFIG_DIR}/.password"

# 生成随机密码（12位）
RANDOM_PASSWORD=$(openssl rand -base64 12 | tr -d '/+' | cut -c1-12)

# 写入密码文件
echo "${RANDOM_PASSWORD}" > "${PASSWORD_FILE}"

# 设置严格的权限
chown ${DEPLOY_USER}:${DEPLOY_USER} "${PASSWORD_FILE}"
chmod 600 "${PASSWORD_FILE}"

log_success "密码文件已生成: ${PASSWORD_FILE}"
log_info "生成的密码: ${RANDOM_PASSWORD}"
log_warning "请妥善保存此密码！"
