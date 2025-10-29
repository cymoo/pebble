# 部署指南

本部署使用了 Rust 实现的后端服务，并配置了 HTTPS。

部署完成后：

- `/memo`   ->  笔记首页
- `/shared` ->  博客首页
- `/`       ->  博客首页（重定向）

## 路径结构

```
# 项目部署路径结构
/var/www/mote/
├── app.db                     # SQLite 数据库文件
├── uploads/                   # 上传的图片目录
├── releases/                  # 存放最近5个构建版本
│   └── 2025-03-16_18-21-44/   # 示例构建版本目录（按时间戳命名）
│       ├── api-dist/          # Rust API 构建产物
│       │   ├── .env             # 环境变量
│       │   ├── static/          # 静态资源
│       │   ├── templates/       # HTML 模板
│       │   └── mote           # API 二进制可执行文件
│       └── web-dist/          # 前端构建产物
│           ├── index.html       # 入口 HTML
│           └── assets/          # 编译后的 JS/CSS/图片等资源
└── current -> releases/2025-03-16_18-21-44  # 软链接指向最新版本

# 系统配置文件路径
/etc/
├── mote/
│   └── secure                  # 存放密码
└── systemd/system/
    └── mote.service
```

## 环境要求

- Linux 服务器（我使用了 Ubuntu 24.04，其他 systemd-based 发行版也可，需修改脚本中的包管理器命令）
- sudo 账号
- 域名且已配置好 DNS 解析
- 开放端口：80 和 443

## 部署

1. 安装基本依赖：Rust Toolchain、Node.js 与 Redis，如果已存在，可忽略此步骤

```bash
./init-env.sh
```

2. 构建和运行后端服务（首次运行会随机生成用于登录的的密码），代码有修改时，重复运行此脚本

```bash
./start-service.sh
```

3. 配置 HTTPS（申请 SSL 证书并自动续期）和 Nginx，首次运行一次即可，域名或 Nginx 模板文件有修改时，重复运行此脚本

```bash
sudo DOMAIN=your.domain EMAIL=your@email.addr ./setup-nginx.sh
```

## 其它

1. 修改登录密码：`sudo ./ch-pwd.sh`

2. 停止服务并清理构建历史：`sudo ./del-service.sh`，注意：它**不会**删除数据库文件和上传目录
