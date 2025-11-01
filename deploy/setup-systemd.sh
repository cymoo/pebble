#!/bin/bash
# 配置systemd服务

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

BACKEND_LANG="$1"

if [ -z "$BACKEND_LANG" ]; then
    log_error "请指定后端语言"
    exit 1
fi

SERVICE_FILE="$DEPLOY_ROOT/config/systemd/${APP_NAME}.service"
SYSTEMD_PATH="/etc/systemd/system/${APP_NAME}.service"

log_info "生成systemd服务文件..."

# 根据语言生成不同的ExecStart
case "$BACKEND_LANG" in
    rust|rs|go)
        EXEC_START="$DEPLOY_ROOT/api/current/mote"
        WORKING_DIR="$DEPLOY_ROOT/api/current"
        ;;
    python|py)
        EXEC_START="$DEPLOY_ROOT/api/current/.venv/bin/gunicorn -k gevent -b ${API_ADDR}:$API_PORT wsgi:app"
        WORKING_DIR="$DEPLOY_ROOT/api/current"
        ;;
    kotlin|kt)
        EXEC_START="/usr/bin/java -jar $DEPLOY_ROOT/api/current/mote.jar"
        WORKING_DIR="$DEPLOY_ROOT/api/current"
        ;;
    *)
        log_error "不支持的语言: $BACKEND_LANG"
        exit 1
        ;;
esac

# 生成服务文件
sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description="$(capitalize "$APP_NAME") Application ($(capitalize "$BACKEND_LANG") backend)"
After=network.target

[Service]
Type=simple
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$WORKING_DIR
EnvironmentFile=$WORKING_DIR/.env
EnvironmentFile=$SECRET_FILE
ExecStart=$EXEC_START
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${APP_NAME}

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DEPLOY_ROOT/data $DEPLOY_ROOT/uploads

[Install]
WantedBy=multi-user.target
EOF

# 创建软链接
log_info "创建systemd软链接..."
sudo ln -sf "$SERVICE_FILE" "$SYSTEMD_PATH"

log_success "Systemd服务配置完成!"
log_info "服务文件: $SERVICE_FILE"
log_info "软链接: $SYSTEMD_PATH"

log_info "启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable ${APP_NAME}
sudo systemctl start ${APP_NAME}
