#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/config.env"

check_root

log_info "开始安装系统依赖..."

# 更新包列表
apt-get update

# 安装基础工具
apt-get install -y curl wget git build-essential pkg-config

# 安装nginx
apt-get install -y nginx

# 安装SQLite
apt-get install -y sqlite3

# # 安装Node.js (用于前端构建)
# curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
# apt-get install -y nodejs

# 安装各语言运行时


NODE_VERSION="v22.14.0"
export RUSTUP_DIST_SERVER=https://mirrors.ustc.edu.cn/rust-static
export RUSTUP_UPDATE_ROOT=https://mirrors.ustc.edu.cn/rust-static/rustup

install_nodejs() {
    log_info "安装Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
}

install_nodejs_manually() {
    log_info "手动安装Node.js..."
    wget https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.xz
    tar -xJf node-${NODE_VERSION}-linux-x64.tar.xz
    rm node-${NODE_VERSION}-linux-x64.tar.xz
    sudo mv node-${NODE_VERSION}-linux-x64 /usr/local/node-${NODE_VERSION}
    sudo ln -sfn /usr/local/node-${NODE_VERSION} /usr/local/node
    sudo ln -sf /usr/local/node/bin/node /usr/bin/node
    sudo ln -sf /usr/local/node/bin/npm /usr/bin/npm
    sudo ln -sf /usr/local/node/bin/npx /usr/bin/npx
}

install_rust() {
    log_info "安装Rust工具链..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source $HOME/.cargo/env
}

install_rust1() {
    log_info "安装Rust工具链..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

    . "$HOME/.cargo/env"

    CONFIG_CONTENT='[source.crates-io]
    replace-with = "ustc"

    [source.ustc]
    registry = "https://mirrors.ustc.edu.cn/crates.io-index"'

    mkdir -p ~/.cargo

    if [[ ! -f ~/.cargo/config.toml ]]; then
        echo "$CONFIG_CONTENT" > ~/.cargo/config.toml
        log_info "Registry configuration written to ~/.cargo/config.toml"
    else
        log_info "${HOME}/.cargo/config.toml already exists. Skipping."
    fi
}

install_go() {
    log_info "安装Go..."
    wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    rm go1.21.0.linux-amd64.tar.gz
}

install_python() {
    log_info "安装Python工具..."
    apt-get install -y python3 python3-pip python3-venv
}

install_kotlin() {
    log_info "安装Kotlin/Java环境..."
    apt-get install -y openjdk-17-jdk maven
}

install_rust
install_go
install_python
install_kotlin

# 安装envsubst
apt-get install -y gettext-base

log_success "所有依赖安装完成"
