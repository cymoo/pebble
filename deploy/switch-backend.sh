#!/bin/bash
# 切换后端语言

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

check_root

BACKEND_LANG="$1"

if [ -z "$BACKEND_LANG" ]; then
    log_error "请指定后端语言: rust, go, python, kotlin"
    exit 1
fi

# 验证语言选择
case "$BACKEND_LANG" in
    rust|go|python|kotlin)
        ;;
    *)
        log_error "不支持的语言: $BACKEND_LANG"
        exit 1
        ;;
esac

BACKEND_DIR="$DEPLOY_ROOT/api/$BACKEND_LANG"

# 检查后端是否存在
if [ ! -d "$BACKEND_DIR" ]; then
    log_error "后端不存在: $BACKEND_DIR"
    log_error "请先部署该后端: make deploy-backend LANG=$BACKEND_LANG"
    exit 1
fi

log_info "切换到 $BACKEND_LANG 后端..."

# 停止服务
if sudo systemctl is-active --quiet mote; then
    log_info "停止服务..."
    sudo systemctl stop mote
fi

# 更新软链接
log_info "更新软链接..."
sudo rm -f "$DEPLOY_ROOT/api/current"
sudo ln -s "$BACKEND_DIR" "$DEPLOY_ROOT/api/current"

# 重新配置systemd
log_info "更新systemd配置..."
bash "${SCRIPT_DIR}/setup-systemd.sh" "$BACKEND_LANG"

# 重启服务
log_info "启动服务..."
sudo systemctl daemon-reload
sudo systemctl start mote

# 等待服务启动
sleep 2

# 检查服务状态
if sudo systemctl is-active --quiet mote; then
    log_success "成功切换到 $BACKEND_LANG 后端!"
    sudo systemctl status mote --no-pager | head -n 10
else
    log_error "服务启动失败!"
    sudo systemctl status mote --no-pager
    exit 1
fi
