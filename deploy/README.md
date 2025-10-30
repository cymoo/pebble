# Deploy

## 部署结构

/opt/mote/                         # 部署根目录
├── api/                           # 后端目录
│   ├── go/                        # Go 后端
│   │   ├── mote                   # 二进制文件
│   │   ├── .env                   # 配置文件
│   │   └── static/                # 静态资源(可选)
│   ├── rust/                      # Rust 后端
│   │   ├── mote                   # 二进制文件
│   │   ├── .env                   # 配置文件
│   │   └── static/                # 静态资源(可选)
│   ├── python/                    # Python 后端
│   │   ├── .venv/                 # 虚拟环境
│   │   ├── app/                   # 应用代码
│   │   ├── migrations/            # 数据库迁移
│   │   ├── templates/             # 模板文件
│   │   ├── static/                # 静态资源
│   │   ├── wsgi.py                # WSGI 入口
│   │   ├── requirements.txt       # Python 依赖
│   │   └── .env                   # 配置文件
│   ├── kotlin/                    # Kotlin 后端
│   │   ├── mote.jar               # JAR 文件
│   │   ├── resources/             # 资源文件
│   │   └── .env                   # 配置文件
│   └── current -> python/         # 软链接指向当前使用的后端
│
├── web/                            # 前端目录
│   ├── build/                     # React 构建产物
│   │   ├── index.html
│   │   ├── assets/
│   │   └── ...
│   └── static/                    # 静态资源(可选)
│
├── data/                           # 数据目录 (权限: 700)
│   └── app.db                     # SQLite 数据库
│
├── uploads/                        # 上传文件目录 (权限: 755)
│   └── [用户上传的文件]
│
├── backups/                        # 备份目录
│   ├── backup-20241030-120000.tar.gz
│   ├── web-backup-20241030-110000/
│   └── python-backup-20241029-100000/
│
└── config/                         # 配置目录
    ├── .password                  # 登录密码 (权限: 600)
    ├── nginx/
    │   └── mote.conf              # Nginx 配置
    └── systemd/
        └── mote.service           # Systemd 服务文件

## 系统服务文件

/etc/nginx/
├── sites-available/
│   └── mote.conf -> /opt/mote/config/nginx/mote.conf
└── sites-enabled/
    └── mote.conf -> /etc/nginx/sites-available/mote.conf

/etc/systemd/system/
└── mote.service -> /opt/mote/config/systemd/mote.service

## 日志位置

系统日志:
├── /var/log/nginx/
│   ├── mote-access.log          # Nginx 访问日志
│   └── mote-error.log           # Nginx 错误日志
└── systemd journal
    └── journalctl -u mote       # 后端服务日志
