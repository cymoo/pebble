#!/bin/bash
# 部署前端脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

check_not_root

FRONTEND_SRC="${PROJECT_ROOT}/frontend"
FRONTEND_DEST="${DEPLOY_ROOT}/web"

log_info "开始部署前端..."

# 检查源目录
if [ ! -d "$FRONTEND_SRC" ]; then
    log_error "前端目录不存在: $FRONTEND_SRC"
    exit 1
fi

# 进入前端目录
cd "$FRONTEND_SRC"

# 安装依赖
# if [ ! -d "node_modules" ]; then
#     log_info "安装前端依赖..."
#     npm install
# else
#     log_info "前端依赖已安装"
# fi
log_info "安装前端依赖..."
npx yarn install --frozen-lockfile --silent

# 构建前端
log_info "构建前端..."
# npm run build
VITE_MEMO_URL=$MEMO_URL VITE_BLOG_URL=$BLOG_URL npx vite build --logLevel error

# 检查构建输出
if [ ! -d "dist" ]; then
    log_error "构建失败: dist目录不存在"
    exit 1
fi

# 备份旧的前端文件
# if [ -d "$FRONTEND_DEST/build" ]; then
#     log_info "备份旧的前端文件..."
#     BACKUP_NAME="web-backup-$(date +%Y%m%d-%H%M%S)"
#     sudo mv "$FRONTEND_DEST/build" "$DEPLOY_ROOT/backups/$BACKUP_NAME"
# fi

# 复制构建文件
log_info "复制构建文件到: $FRONTEND_DEST"
sudo mkdir -p "$FRONTEND_DEST/build"
sudo cp -r dist/* "$FRONTEND_DEST/build/"

# # 复制静态资源(如果存在)
# if [ -d "$FRONTEND_SRC/public" ]; then
#     log_info "复制静态资源..."
#     sudo mkdir -p "$FRONTEND_DEST/static"
#     sudo cp -r "$FRONTEND_SRC/public"/* "$FRONTEND_DEST/static/" 2>/dev/null || true
# fi

# 设置权限
log_info "设置权限..."
sudo chown -R "$APP_USER:$APP_USER" "$FRONTEND_DEST"
sudo chmod -R 755 "$FRONTEND_DEST"

# 重载Nginx(如果已配置)
if sudo systemctl is-active --quiet nginx && [ -f "/etc/nginx/sites-enabled/mote.conf" ]; then
    log_info "重载Nginx配置..."
    sudo nginx -t && sudo systemctl reload nginx
fi

log_success "前端部署完成!"
log_info "前端文件位置: $FRONTEND_DEST/build"
