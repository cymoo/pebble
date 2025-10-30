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

# 验证语言选择
case "$BACKEND_LANG" in
    rust|go|python|kotlin)
        log_info "部署 $BACKEND_LANG 后端..."
        ;;
    *)
        log_error "不支持的语言: $BACKEND_LANG"
        log_error "支持的语言: rust, go, python, kotlin"
        exit 1
        ;;
esac

# 定义源目录和目标目录
case "$BACKEND_LANG" in
    rust)
        SRC_DIR="${PROJECT_ROOT}/api-rs"
        DEST_DIR="${DEPLOY_ROOT}/api/rust"
        ;;
    go)
        SRC_DIR="${PROJECT_ROOT}/api-go"
        DEST_DIR="${DEPLOY_ROOT}/api/go"
        ;;
    python)
        SRC_DIR="${PROJECT_ROOT}/api-py"
        DEST_DIR="${DEPLOY_ROOT}/api/python"
        ;;
    kotlin)
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
    rust)
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
        go build -o mote .
        BINARY_PATH="mote"
        ;;

    python)
        # Python - 不需要编译
        log_info "准备Python环境..."
        BINARY_PATH=""
        ;;

    kotlin)
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
if sudo systemctl is-active --quiet mote; then
    log_info "停止现有服务..."
    sudo systemctl stop mote
fi

# 备份旧版本
if [ -d "$DEST_DIR" ]; then
    log_info "备份旧版本..."
    BACKUP_NAME="${BACKEND_LANG}-backup-$(date +%Y%m%d-%H%M%S)"
    sudo mv "$DEST_DIR" "$DEPLOY_ROOT/backups/$BACKUP_NAME"
fi

# 创建目标目录
sudo mkdir -p "$DEST_DIR"

# 复制文件
log_info "复制文件到: $DEST_DIR"
case "$BACKEND_LANG" in
    rust|go)
        # 复制二进制文件
        sudo cp "$BINARY_PATH" "$DEST_DIR/mote"
        sudo chmod +x "$DEST_DIR/mote"

        # 复制静态资源
        if [ -d "static" ]; then
            sudo cp -r static "$DEST_DIR/"
        fi
        ;;

    python)
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

    kotlin)
        # 复制JAR文件
        JAR_FILE=$(ls target/mote-*.jar | head -n 1)
        sudo cp "$JAR_FILE" "$DEST_DIR/mote.jar"

        # 复制资源文件
        if [ -d "src/main/resources" ]; then
            sudo cp -r src/main/resources "$DEST_DIR/"
        fi
        ;;
esac

# 复制.env文件
if [ -f ".env" ]; then
    log_info "复制配置文件..."
    sudo cp .env "$DEST_DIR/"

    # 更新环境变量
    sudo sed -i "s|DATABASE_URL=.*|DATABASE_URL=${DB_PATH}|g" "$DEST_DIR/.env" || true
    sudo sed -i "s|UPLOAD_DIR=.*|UPLOAD_DIR=${UPLOADS_DIR}|g" "$DEST_DIR/.env" || true
    sudo sed -i "s|PORT=.*|PORT=${BACKEND_PORT}|g" "$DEST_DIR/.env" || true
fi

# 设置权限
log_info "设置权限..."
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
log_info "配置Nginx..."
bash "${SCRIPT_DIR}/setup-nginx.sh"

# 启动服务
log_info "启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable mote
sudo systemctl start mote

# 等待服务启动
sleep 2

# 检查服务状态
if sudo systemctl is-active --quiet mote; then
    log_success "$BACKEND_LANG 后端部署成功!"
    log_info "服务状态:"
    sudo systemctl status mote --no-pager | head -n 10
    log_info "查看日志: make logs"
else
    log_error "服务启动失败!"
    sudo systemctl status mote --no-pager
    exit 1
fi

# 启动Nginx
if ! sudo systemctl is-active --quiet nginx; then
    log_info "启动Nginx..."
    sudo systemctl enable nginx
    sudo systemctl start nginx
fi

log_success "部署完成!"
log_info "后端位置: $DEST_DIR"
log_info "当前后端: $BACKEND_LANG"
log_info "访问地址: http://localhost"
