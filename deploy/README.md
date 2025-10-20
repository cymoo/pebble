# Pebble 部署文档

这是一个完整的部署系统，支持前端（Vite + React）和四种后端实现（Go、Python、Kotlin、Rust）的自动化部署。

## 目录结构

```
.
├── Makefile                      # 主部署入口
├── scripts/
│   ├── install-deps.sh          # 安装系统依赖
│   ├── deploy-frontend.sh       # 部署前端
│   ├── deploy-backend.sh        # 部署后端（调度器）
│   ├── setup-nginx.sh           # 配置 Nginx 和 SSL
│   ├── cleanup.sh               # 清理部署
│   ├── nginx.template           # Nginx 配置模板（可选）
│   └── backends/
│       ├── deploy-go.sh         # Go 后端部署
│       ├── deploy-py.sh         # Python 后端部署
│       ├── deploy-kt.sh         # Kotlin 后端部署
│       └── deploy-rs.sh         # Rust 后端部署
├── pebble/                      # 前端源码
├── api-go/                      # Go 后端源码
├── api-py/                      # Python 后端源码
├── api-kt/                      # Kotlin 后端源码
└── api-rs/                      # Rust 后端源码
```

## 系统要求

- Ubuntu 20.04+ / Debian 11+
- Root 权限或 sudo 访问
- 域名已正确解析到服务器 IP
- 端口 80、443 和 8000 可访问

## 快速开始

### 1. 准备脚本目录

```bash
# 创建 scripts 目录结构
mkdir -p scripts/backends
chmod +x scripts/*.sh
chmod +x scripts/backends/*.sh
```

### 2. 完整部署

选择一个后端进行部署：

```bash
# 部署 Go 后端
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go

# 部署 Python 后端
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=py

# 部署 Kotlin 后端
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=kt

# 部署 Rust 后端
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=rs
```

### 3. 单独部署前端或后端

```bash
# 只部署前端
make deploy-frontend

# 只部署后端（需要先设置 BACKEND 变量）
make deploy-backend BACKEND=go

# 切换到不同的后端
make deploy-backend BACKEND=py
```

## 部署流程说明

### 完整部署会执行以下步骤：

1. **安装依赖** (`install-deps.sh`)
   - 更新系统包
   - 安装 Nginx、Node.js、构建工具

2. **部署前端** (`deploy-frontend.sh`)
   - 在 `pebble/` 目录执行 `npm ci` 安装依赖
   - 执行 `npm run build` 构建生产版本
   - 将构建结果复制到 `/var/www/pebble/`

3. **部署后端** (`deploy-backend.sh` + 对应的后端脚本)
   - 安装所需的运行时环境（Go/Python/Java/Rust）
   - 构建/编译后端代码
   - 部署到 `/opt/pebble/backend/`
   - 创建 systemd 服务并启动

4. **配置 Nginx** (`setup-nginx.sh`)
   - 生成 DH 参数
   - 申请 Let's Encrypt SSL 证书
   - 配置 HTTPS 和反向代理
   - 设置自动续期任务

## 各后端的特殊说明

### Go 后端
- 自动下载并安装 Go 1.21.5
- 构建单个可执行文件
- 启动最快，占用资源少

### Python 后端
- 创建虚拟环境
- 使用 Gunicorn 4 个 worker 进程
- 需要 `requirements.txt` 文件
- Flask 应用需要命名为 `app.py` 并导出 `app` 对象

### Kotlin 后端
- 安装 OpenJDK 17 和 Maven
- 使用 Maven 构建 JAR 包
- 支持 Spring Boot 配置文件

### Rust 后端
- 安装最新稳定版 Rust
- 编译时间较长（首次可能需要 5-10 分钟）
- 生成高度优化的二进制文件

## 配置文件

### Nginx 配置模板

如果你有自定义的 Nginx 配置，创建 `scripts/nginx.template` 文件，可用变量：
- `${WWW_ROOT}` - 前端文件路径 (`/var/www/pebble`)
- `${SERVER_NAME}` - 域名
- `${API_PORT}` - 后端端口（默认 8000）

### 环境变量

可以在部署时设置以下环境变量：

```bash
# 修改 API 端口
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go API_PORT=9000
```

## 常用命令

```bash
# 查看服务状态
make status

# 查看日志
make logs

# 重启服务
make restart

# 清理部署
make clean

# 显示帮助
make help
```

## 部署后的目录结构

```
/opt/pebble/backend/          # 后端部署目录
├── pebble-api               # 可执行文件（Go/Rust）或
├── pebble-api.jar          # JAR 文件（Kotlin）或
├── venv/                   # Python 虚拟环境
├── start.sh                # 启动脚本
└── ...                     # 其他配置文件

/var/www/pebble/            # 前端文件
├── index.html
├── assets/
└── ...

/etc/systemd/system/
└── pebble-backend.service  # Systemd 服务文件

/etc/nginx/conf.d/
└── example.com.conf        # Nginx 配置

/etc/letsencrypt/live/
└── example.com/            # SSL 证书
```

## 日志位置

- **后端日志**: `journalctl -u pebble-backend -f`
- **Nginx 访问日志**: `/var/log/nginx/access.log`
- **Nginx 错误日志**: `/var/log/nginx/error.log`

## 故障排查

### 后端服务无法启动

```bash
# 查看详细日志
sudo journalctl -u pebble-backend -n 100 --no-pager

# 检查端口是否被占用
sudo netstat -tlnp | grep 8000

# 手动测试启动
cd /opt/pebble/backend
sudo -u www-data ./start.sh
```

### SSL 证书申请失败

```bash
# 确认域名解析正确
dig example.com

# 检查防火墙
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 手动申请证书（调试模式）
sudo certbot certonly --webroot -w /var/www/certbot -d example.com --dry-run
```

### 前端 404 错误

```bash
# 检查文件权限
ls -la /var/www/pebble/

# 确认 Nginx 配置
sudo nginx -t
sudo cat /etc/nginx/conf.d/example.com.conf

# 重新部署前端
make deploy-frontend
```

## 更新部署

### 更新前端

```bash
cd pebble/
git pull
cd ..
make deploy-frontend
```

### 更新后端

```bash
cd api-go/  # 或其他后端目录
git pull
cd ..
make deploy-backend BACKEND=go
```

### 切换后端实现

```bash
# 从 Go 切换到 Python
make deploy-backend BACKEND=py

# 服务会自动停止旧后端并启动新后端
```

## 安全建议

1. **防火墙配置**
   ```bash
   sudo ufw enable
   sudo ufw allow 22/tcp
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

2. **定期更新系统**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

3. **监控日志**
   ```bash
   # 设置日志轮转
   sudo logrotate -f /etc/logrotate.d/nginx
   ```

4. **备份证书**
   ```bash
   sudo tar -czf letsencrypt-backup.tar.gz /etc/letsencrypt/
   ```

## 卸载

完全移除部署：

```bash
make clean
# 按提示选择是否删除 Nginx 配置和 SSL 证书
```

## 支持

如遇问题，请检查：
1. 系统日志：`journalctl -xe`
2. 服务状态：`make status`
3. 端口监听：`sudo netstat -tlnp`
4. 磁盘空间：`df -h`