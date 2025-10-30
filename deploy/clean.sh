#!/bin/bash
# 清理部署文件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

CLEAN_ALL=false

# 检查参数
if [ "$1" == "--all" ]; then
    CLEAN_ALL=true
fi

log_warn "开始清理..."

# 停止服务
if sudo systemctl is-active --quiet mote; then
    log_info "停止服务..."
    sudo systemctl stop mote
    sudo systemctl disable mote
fi

# 移除systemd服务文件
if [ -f "/etc/systemd/system/mote.service" ]; then
    log_info "移除systemd服务..."
    sudo rm -f /etc/systemd/system/mote.service
    sudo systemctl daemon-reload
fi

# 移除Nginx配置
if [ -f "/etc/nginx/sites-enabled/mote.conf" ]; then
    log_info "移除Nginx配置..."
    sudo rm -f /etc/nginx/sites-enabled/mote.conf
    sudo rm -f /etc/nginx/sites-available/mote.conf
    sudo systemctl reload nginx 2>/dev/null || true
fi

# 清理后端文件
log_info "清理后端文件..."
sudo rm -rf "$DEPLOY_ROOT/api/rust"
sudo rm -rf "$DEPLOY_ROOT/api/go"
sudo rm -rf "$DEPLOY_ROOT/api/python"
sudo rm -rf "$DEPLOY_ROOT/api/kotlin"
sudo rm -f "$DEPLOY_ROOT/api/current"

# 清理前端文件
log_info "清理前端文件..."
sudo rm -rf "$DEPLOY_ROOT/web/build"
sudo rm -rf "$DEPLOY_ROOT/web/static"

# 清理配置文件
log_info "清理配置文件..."
sudo rm -rf "$DEPLOY_ROOT/config/nginx"
sudo rm -rf "$DEPLOY_ROOT/config/systemd"

if [ "$CLEAN_ALL" == true ]; then
    log_warn "执行完全清理(包括数据和上传文件)..."

    # 备份数据库和上传文件
    if [ -f "$DB_PATH" ] || [ -d "$UPLOADS_DIR" ]; then
        BACKUP_NAME="full-backup-$(date +%Y%m%d-%H%M%S)"
        BACKUP_PATH="$DEPLOY_ROOT/backups/$BACKUP_NAME"

        log_info "备份数据到: $BACKUP_PATH"
        sudo mkdir -p "$BACKUP_PATH"

        [ -f "$DB_PATH" ] && sudo cp "$DB_PATH" "$BACKUP_PATH/"
        [ -d "$UPLOADS_DIR" ] && sudo cp -r "$UPLOADS_DIR" "$BACKUP_PATH/"

        log_success "数据已备份到: $BACKUP_PATH"
    fi

    # 清理数据
    log_info "清理数据库..."
    sudo rm -rf "$DEPLOY_ROOT/data"

    log_info "清理上传文件..."
    sudo rm -rf "$DEPLOY_ROOT/uploads"

    log_info "清理密码文件..."
    sudo rm -f "$DEPLOY_ROOT/config/.password"

    # 询问是否删除整个部署目录
    read -p "是否删除整个部署目录 $DEPLOY_ROOT? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "删除部署目录..."
        sudo rm -rf "$DEPLOY_ROOT"
        log_success "已完全清理!"
    fi
else
    log_info "保留数据、上传文件和密码"
    log_info "如需完全清理,请使用: make clean-all"
fi

log_success "清理完成!"

if [ -d "$DEPLOY_ROOT/backups" ]; then
    BACKUP_COUNT=$(sudo ls -1 "$DEPLOY_ROOT/backups" 2>/dev/null | wc -l)
    if [ "$BACKUP_COUNT" -gt 0 ]; then
        log_info "现有备份: $BACKUP_COUNT 个"
        log_info "备份位置: $DEPLOY_ROOT/backups"
    fi
fi
