#!/bin/bash
# 健康检查脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

check_root

log_info "执行健康检查..."

HEALTH_STATUS=0

# 检查部署目录
log_info "检查部署目录..."
if [ -d "$DEPLOY_ROOT" ]; then
    log_success "部署目录存在: $DEPLOY_ROOT"
else
    log_error "部署目录不存在: $DEPLOY_ROOT"
    HEALTH_STATUS=1
fi

# 检查Nginx
log_info "检查Nginx..."
if sudo systemctl is-active --quiet nginx; then
    log_success "Nginx运行中"

    # 测试Nginx配置
    if sudo nginx -t 2>&1 | grep -q "successful"; then
        log_success "Nginx配置有效"
    else
        log_error "Nginx配置无效"
        HEALTH_STATUS=1
    fi
else
    log_error "Nginx未运行"
    HEALTH_STATUS=1
fi

# 检查后端服务
log_info "检查后端服务..."
if sudo systemctl is-active --quiet ${APP_NAME}; then
    log_success "后端服务运行中"

    # 检查当前后端
    if [ -L "$DEPLOY_ROOT/api/current" ]; then
        CURRENT_BACKEND=$(basename "$(readlink "$DEPLOY_ROOT/api/current")")
        log_info "当前后端: $CURRENT_BACKEND"
    else
        log_warn "未设置当前后端"
    fi
else
    log_error "后端服务未运行"
    HEALTH_STATUS=1
fi

# 检查端口监听
log_info "检查端口监听..."
if sudo netstat -tlnp 2>/dev/null | grep -q ":$API_PORT"; then
    log_success "后端端口 $API_PORT 正在监听"
else
    log_error "后端端口 $API_PORT 未监听"
    HEALTH_STATUS=1
fi

if sudo netstat -tlnp 2>/dev/null | grep -q ":80"; then
    log_success "Nginx端口 80 正在监听"
else
    log_error "Nginx端口 80 未监听"
    HEALTH_STATUS=1
fi

# 检查数据库
log_info "检查数据库..."
if [ -f "$DB_PATH" ]; then
    log_success "数据库存在: $DB_PATH"

    # 检查数据库权限
    if sudo -u "$APP_USER" test -r "$DB_PATH"; then
        log_success "数据库可读"
    else
        log_error "数据库权限错误"
        HEALTH_STATUS=1
    fi
else
    log_warn "数据库文件不存在: $DB_PATH (首次部署正常)"
fi

# 检查上传目录
log_info "检查上传目录..."
if [ -d "$UPLOADS_DIR" ]; then
    log_success "上传目录存在: $UPLOADS_DIR"

    if sudo -u "$APP_USER" test -w "$UPLOADS_DIR"; then
        log_success "上传目录可写"
    else
        log_error "上传目录权限错误"
        HEALTH_STATUS=1
    fi
else
    log_error "上传目录不存在"
    HEALTH_STATUS=1
fi

# 检查前端文件
log_info "检查前端文件..."
if [ -d "$DEPLOY_ROOT/web/build" ]; then
    if [ -f "$DEPLOY_ROOT/web/build/index.html" ]; then
        log_success "前端文件存在"
    else
        log_error "前端index.html不存在"
        HEALTH_STATUS=1
    fi
else
    log_error "前端构建目录不存在"
    HEALTH_STATUS=1
fi

# HTTP健康检查
log_info "执行HTTP健康检查..."
if command -v curl &> /dev/null; then
    # 检查后端
    if curl -sf "http://localhost:$API_PORT/health" > /dev/null 2>&1 || \
       curl -sf "http://localhost:$API_PORT/api/health" > /dev/null 2>&1; then
        log_success "后端HTTP响应正常"
    else
        log_warn "后端HTTP响应失败(可能未实现/health端点)"
    fi

    # 检查前端
    if curl -sf "http://localhost/" > /dev/null 2>&1; then
        log_success "前端HTTP响应正常"
    else
        log_error "前端HTTP响应失败"
        HEALTH_STATUS=1
    fi
else
    log_warn "curl未安装,跳过HTTP检查"
fi

# 磁盘空间检查
log_info "检查磁盘空间..."
DISK_USAGE=$(df -h "$DEPLOY_ROOT" | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 80 ]; then
    log_success "磁盘使用率: ${DISK_USAGE}%"
elif [ "$DISK_USAGE" -lt 90 ]; then
    log_warn "磁盘使用率较高: ${DISK_USAGE}%"
else
    log_error "磁盘空间不足: ${DISK_USAGE}%"
    HEALTH_STATUS=1
fi

# 内存检查
log_info "检查内存使用..."
MEM_AVAILABLE=$(free -m | awk 'NR==2{print $7}')
if [ "$MEM_AVAILABLE" -gt 500 ]; then
    log_success "可用内存: ${MEM_AVAILABLE}MB"
else
    log_warn "可用内存较低: ${MEM_AVAILABLE}MB"
fi

# 总结
echo ""
echo "========================================"
if [ $HEALTH_STATUS -eq 0 ]; then
    log_success "健康检查通过!"
else
    log_error "健康检查发现问题,请查看上述错误信息"
fi
echo "========================================"

exit $HEALTH_STATUS
