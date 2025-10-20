# Pebble 部署文档

这是一个完整的部署系统，支持前端（Vite + React）和四种后端实现（Go、Python、Kotlin、Rust）的自动化部署。支持自动备份、一键回滚和 Docker 容器化部署。

## 目录结构

```
.
├── Makefile                      # 主部署入口
├── Dockerfile                    # Docker 镜像构建文件
├── docker-compose.yml            # Docker Compose 配置
├── docker-build.sh              # Docker 构建脚本
├── docker-run.sh                # Docker 运行脚本
├── .dockerignore                # Docker 忽略文件
├── scripts/
│   ├── install-deps.sh          # 安装系统依赖
│   ├── deploy-frontend.sh       # 部署前端
│   ├── deploy-backend.sh        # 部署后端（调度器）
│   ├── setup-nginx.sh           # 配置 Nginx 和 SSL
│   ├── backup.sh                # 创建备份
│   ├── list-backups.sh          # 列出备份
│   ├── rollback.sh              # 回滚部署
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

# 创建备份
make backup

# 列出所有备份
make list-backups

# 回滚到最新备份
make rollback

# 回滚到指定备份
make rollback BACKUP_ID=20250120_143022

# 清理部署
make clean

# 显示帮助
make help
```

## 备份和回滚功能

### 自动备份

部署脚本会在每次部署前自动创建备份，默认保留最近 5 个备份。可以通过环境变量修改：

```bash
# 保留 10 个备份
make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go MAX_BACKUPS=10
```

### 手动备份

```bash
# 创建手动备份
make backup

# 指定保留数量
make backup MAX_BACKUPS=10
```

### 查看备份列表

```bash
make list-backups
```

输出示例：
```
ID               Created              Size      Backend   Files
===============================================================================
20250120_143022  2025-01-20 14:30:22  45M       go        FBS *
20250120_120000  2025-01-20 12:00:00  43M       python    FBS
20250119_180000  2025-01-19 18:00:00  44M       go        FBS

Legend: F=Frontend, B=Backend, S=Service, * = Latest backup
```

### 回滚部署

```bash
# 回滚到最新备份
make rollback

# 回滚到指定备份
make rollback BACKUP_ID=20250120_143022
```

回滚会：
1. 创建当前状态的安全备份
2. 停止后端服务
3. 恢复前端、后端和服务配置
4. 重启服务

### 备份存储位置

- 备份目录: `/opt/pebble/backups/`
- 每个备份包含:
  - `frontend.tar.gz` - 前端文件
  - `backend.tar.gz` - 后端文件
  - `pebble-backend.service` - Systemd 服务配置
  - `metadata.txt` - 备份元数据

## Docker 部署

### 快速开始

```bash
# 1. 构建镜像
chmod +x docker-build.sh
./docker-build.sh

# 2. 运行容器
chmod +x docker-run.sh
./docker-run.sh

# 或使用 docker-compose
docker-compose up -d
```

### Docker 镜像特性

- **多阶段构建**: 分离构建和运行环境
- **最小化镜像**: 基于 Alpine Linux，最终镜像 < 30MB
- **非 root 用户**: 以普通用户运行，提高安全性
- **健康检查**: 内置健康检查端点
- **静态编译**: Go 二进制静态链接，无外部依赖

### 镜像大小优化

Dockerfile 使用以下技术减小镜像大小：

1. **多阶段构建**: 构建工具不包含在最终镜像
2. **Alpine 基础镜像**: 最小的 Linux 发行版
3. **静态编译**: CGO_ENABLED=0，无需 libc
4. **编译优化**: `-ldflags='-w -s'` 移除调试信息
5. **单一二进制**: 无需运行时依赖

预期镜像大小：
- Frontend build layer: ~200MB (不包含在最终镜像)
- Backend build layer: ~400MB (不包含在最终镜像)
- **最终镜像: ~20-30MB**

### 自定义构建

```bash
# 指定镜像名称和标签
IMAGE_NAME=myapp IMAGE_TAG=v1.0.0 ./docker-build.sh

# 手动构建
docker build -t pebble:latest .

# 查看镜像大小
docker images pebble:latest
```

### 运行容器

```bash
# 基本运行
docker run -d -p 8000:8000 --name pebble pebble:latest

# 自定义端口
docker run -d -p 9000:8000 --name pebble pebble:latest

# 挂载配置文件
docker run -d -p 8000:8000 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  --name pebble pebble:latest

# 查看日志
docker logs -f pebble

# 进入容器
docker exec -it pebble sh
```

### Docker Compose 部署

```bash
# 启动服务
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build
```

### 生产环境部署

```bash
# 1. 构建镜像
docker build -t your-registry.com/pebble:v1.0.0 .

# 2. 推送到镜像仓库
docker push your-registry.com/pebble:v1.0.0

# 3. 在生产服务器拉取并运行
docker pull your-registry.com/pebble:v1.0.0
docker run -d \
  --name pebble \
  --restart unless-stopped \
  -p 8000:8000 \
  -e GIN_MODE=release \
  your-registry.com/pebble:v1.0.0
```

### 与 Nginx 结合使用

使用 Docker Compose 的 nginx 服务作为反向代理：

```yaml
# docker-compose.yml 已包含 nginx 配置
services:
  pebble:
    # ... pebble 配置
  
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
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

### 备份失败

```bash
# 检查磁盘空间
df -h

# 检查备份目录权限
ls -la /opt/pebble/backups/

# 手动创建备份
sudo /path/to/scripts/backup.sh
```

### 回滚失败

```bash
# 检查备份完整性
tar -tzf /opt/pebble/backups/BACKUP_ID/backend.tar.gz
tar -tzf /opt/pebble/backups/BACKUP_ID/frontend.tar.gz

# 手动恢复
sudo tar -xzf /opt/pebble/backups/BACKUP_ID/backend.tar.gz -C /opt/pebble/
sudo systemctl restart pebble-backend
```

### Docker 容器问题

```bash
# 查看容器日志
docker logs pebble-app

# 检查容器状态
docker ps -a | grep pebble

# 进入容器调试
docker exec -it pebble-app sh

# 检查健康状态
docker inspect --format='{{.State.Health.Status}}' pebble-app

# 重建容器
docker-compose down
docker-compose up -d --build
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