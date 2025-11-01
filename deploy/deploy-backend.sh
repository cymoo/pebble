#!/bin/bash
# 部署后端脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

check_root

# 获取后端语言参数
BACKEND_LANG="$1"

if [ -z "$BACKEND_LANG" ]; then
    log_error "请指定后端语言: rust, go, python, kotlin"
    exit 1
fi

# # 验证语言选择
# case "$BACKEND_LANG" in
#     rust|go|python|kotlin)
#         log_info "部署 $BACKEND_LANG 后端..."
#         ;;
#     *)
#         log_error "不支持的语言: $BACKEND_LANG"
#         log_error "支持的语言: rust, go, python, kotlin"
#         exit 1
#         ;;
# esac

# 定义源目录和目标目录
case "$BACKEND_LANG" in
    rust|rs)
        SRC_DIR="${PROJECT_ROOT}/api-rs"
        DEST_DIR="${DEPLOY_ROOT}/api/rust"
        ;;
    go)
        SRC_DIR="${PROJECT_ROOT}/api-go"
        DEST_DIR="${DEPLOY_ROOT}/api/go"
        ;;
    python|py)
        SRC_DIR="${PROJECT_ROOT}/api-py"
        DEST_DIR="${DEPLOY_ROOT}/api/python"
        ;;
    kotlin|kt)
        SRC_DIR="${PROJECT_ROOT}/api-kt"
        DEST_DIR="${DEPLOY_ROOT}/api/kotlin"
        ;;
esac

# 检查源目录
if [ ! -d "$SRC_DIR" ]; then
    log_error "后端源目录不存在: $SRC_DIR"
    exit 1
fi

# 构建后端
log_info "构建 $BACKEND_LANG 后端..."
cd "$SRC_DIR"

case "$BACKEND_LANG" in
    rust|rs)
        # Rust构建
        if [ ! -f "$HOME/.cargo/env" ]; then
            log_error "Rust未安装，请先运行 'make install'"
            exit 1
        fi
        source "$HOME/.cargo/env"
        cargo build --release
        BINARY_PATH="target/release/mote"
        ;;

    go)
        # Go构建
        if ! check_command go; then
            log_error "Go未安装，请先运行 'make install'"
            exit 1
        fi
        export PATH=$PATH:/usr/local/go/bin
        go build -o mote ./cmd/server
        BINARY_PATH="mote"
        ;;

    python|py)
        # Python - 不需要编译
        log_info "准备Python环境..."
        BINARY_PATH=""
        ;;

    kotlin|kt)
        # Kotlin构建
        if ! check_command mvn; then
            log_error "Maven未安装，请先运行 'make install'"
            exit 1
        fi
        mvn clean package -DskipTests
        BINARY_PATH="target/mote-*.jar"
        ;;
esac

# 停止现有服务
if sudo systemctl is-active --quiet ${APP_NAME}; then
    log_info "停止现有服务..."
    sudo systemctl stop ${APP_NAME}
fi

# # 备份旧版本
# if [ -d "$DEST_DIR" ]; then
#     log_info "备份旧版本..."
#     BACKUP_NAME="${BACKEND_LANG}-backup-$(date +%Y%m%d-%H%M%S)"
#     sudo mv "$DEST_DIR" "$DEPLOY_ROOT/backups/$BACKUP_NAME"
# fi

# 创建目标目录
sudo mkdir -p "$DEST_DIR"

# 复制文件
log_info "复制文件到: $DEST_DIR"
case "$BACKEND_LANG" in
    rust|rs|go)
        # 复制二进制文件
        sudo mv "$BINARY_PATH" "$DEST_DIR/"
        sudo chmod +x "$DEST_DIR/mote"

        # 复制静态资源
        if [ -d "static" ]; then
            # sudo cp -r static "$DEST_DIR/"
            sudo cp -r static "${WEB_DIR}/"
        fi
        ;;

    python|py)
        # 复制所有Python文件
        sudo cp -r . "$DEST_DIR/"

        # 创建虚拟环境
        log_info "创建Python虚拟环境..."
        sudo -u "$APP_USER" python3 -m venv "$DEST_DIR/.venv"

        # 安装依赖
        log_info "安装Python依赖..."
        sudo -u "$APP_USER" "$DEST_DIR/.venv/bin/pip" install --upgrade pip
        sudo -u "$APP_USER" "$DEST_DIR/.venv/bin/pip" install -r "$DEST_DIR/requirements.txt"
        sudo -u "$APP_USER" "$DEST_DIR/.venv/bin/pip" install gunicorn
        ;;

    kotlin|kt)
        # 复制JAR文件
        JAR_FILE=$(ls target/mote-*.jar | head -n 1)
        sudo mv "$JAR_FILE" "$DEST_DIR/mote.jar"

        # TODO: 需要复制资源文件吗？
        if [ -d "src/main/resources" ]; then
            sudo cp -r src/main/resources "$DEST_DIR/"
        fi
        ;;
    *)
        log_error "不支持的语言: $BACKEND_LANG"
        exit 1
        ;;
esac

# If static directory exists, set permissions
if [ -d "$FRONTEND_DEST/static" ]; then
    sudo chown -R "$APP_USER:$APP_USER" "$FRONTEND_DEST/static"
    sudo chmod -R 755 "$FRONTEND_DEST/static"
fi

log_info "生成环境配置文件..."

BASE_ENV_TEMPLATE="""# 基础配置
UPLOAD_PATH=${UPLOADS_DIR}
HTTP_PORT=${API_PORT}
HTTP_IP=${API_ADDR}
LOG_REQUESTS=false
"""

# 根据语言添加特定的环境变量
case "$BACKEND_LANG" in
    rust|rs)
        LANGUAGE_SPECIFIC="""
RUST_LOG=info
DATABASE_URL=sqlite://${DB_PATH}
"""
        ;;
    go)
        LANGUAGE_SPECIFIC="""
APP_ENV=prod
DATABASE_URL=${DB_PATH}
"""
        ;;
    python|py)
        LANGUAGE_SPECIFIC="""
FLASK_ENV=production
DATABASE_URL=sqlite:///${DB_PATH}
"
        ;;
    kotlin|kt)
        LANGUAGE_SPECIFIC="""
SPRING_PROFILES_ACTIVE=prod
DATABASE_URL=sqlite:${DB_PATH}
"""
        ;;
    *)
        log_error "不支持的语言: $BACKEND_LANG"
        exit 1
        ;;
esac

# 组合并生成环境配置文件
ENV_CONTENT="$BASE_ENV_TEMPLATE
$LANGUAGE_SPECIFIC"

# 确保目标目录存在
sudo mkdir -p "$DEST_DIR"

# 写入环境文件
sudo tee "${DEST_DIR}/.env" > /dev/null <<EOF
$ENV_CONTENT
EOF


# # 复制.env文件
# if [ -f ".env" ]; then
#     log_info "复制配置文件..."
#     sudo cp .env "$DEST_DIR/"

#     # 更新环境变量
#     sudo sed -i "s|DATABASE_URL=.*|DATABASE_URL=${DB_PATH}|g" "$DEST_DIR/.env" || true
#     sudo sed -i "s|UPLOAD_DIR=.*|UPLOAD_DIR=${UPLOADS_DIR}|g" "$DEST_DIR/.env" || true
#     sudo sed -i "s|PORT=.*|PORT=${API_PORT}|g" "$DEST_DIR/.env" || true
# fi

# 设置权限
log_info "设置权限..."
ensure_user $APP_USER
sudo chown -R "$APP_USER:$APP_USER" "$DEST_DIR"
sudo chmod -R 755 "$DEST_DIR"
[ -f "$DEST_DIR/.env" ] && sudo chmod 600 "$DEST_DIR/.env"

# 更新软链接
log_info "更新当前后端软链接..."
sudo rm -f "$DEPLOY_ROOT/api/current"
sudo ln -s "$DEST_DIR" "$DEPLOY_ROOT/api/current"

# 配置systemd服务
log_info "配置systemd服务..."
bash "${SCRIPT_DIR}/setup-systemd.sh" "$BACKEND_LANG"

# 配置Nginx
# log_info "配置Nginx..."
# bash "${SCRIPT_DIR}/setup-nginx.sh"

# 启动服务
# log_info "启动服务..."
# sudo systemctl daemon-reload
# sudo systemctl enable ${APP_NAME}
# sudo systemctl start ${APP_NAME}

# 等待服务启动
sleep 2

# 检查服务状态
if sudo systemctl is-active --quiet ${APP_NAME}; then
    log_success "$BACKEND_LANG 后端部署成功!"
    log_info "服务状态:"
    sudo systemctl status ${APP_NAME} --no-pager | head -n 10
else
    log_error "服务启动失败!"
    sudo systemctl status ${APP_NAME} --no-pager
    exit 1
fi

# # 启动Nginx
# if ! sudo systemctl is-active --quiet nginx; then
#     log_info "启动Nginx..."
#     sudo systemctl enable nginx
#     sudo systemctl start nginx
# fi

log_success "部署完成!"
# log_info "后端位置: $DEST_DIR"
# log_info "当前后端: $BACKEND_LANG"
# log_info "访问地址: http://localhost"
