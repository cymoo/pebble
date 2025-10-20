#!/bin/bash

set -euo pipefail

NODE_VERSION=v22.14.0

sudo apt-get update -q
sudo apt-get install -q -y lsb-release curl gpg libssl-dev pkg-config nginx

# ==================== Install Rust ====================
echo "Installing Rust..."
export RUSTUP_DIST_SERVER=https://mirrors.ustc.edu.cn/rust-static
export RUSTUP_UPDATE_ROOT=https://mirrors.ustc.edu.cn/rust-static/rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

. "$HOME/.cargo/env"

CONFIG_CONTENT='[source.crates-io]
replace-with = "ustc"

[source.ustc]
registry = "https://mirrors.ustc.edu.cn/crates.io-index"'

mkdir -p ~/.cargo

if [[ ! -f ~/.cargo/config.toml ]]; then
    echo "$CONFIG_CONTENT" > ~/.cargo/config.toml
    echo "Registry configuration written to ~/.cargo/config.toml"
else
    echo "${HOME}/.cargo/config.toml already exists. Skipping."
fi

# ==================== Install Redis ====================
echo "Installing Redis..."
curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
sudo chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list
sudo apt-get update -q
sudo apt-get install -q -y redis
sudo systemctl enable redis-server
sudo systemctl start redis-server

# ==================== Install Node.js ====================
echo "Installing Node.js..."
wget https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.xz
tar -xJf node-${NODE_VERSION}-linux-x64.tar.xz
rm node-${NODE_VERSION}-linux-x64.tar.xz
sudo mv node-${NODE_VERSION}-linux-x64 /usr/local/node-${NODE_VERSION}
sudo ln -sfn /usr/local/node-${NODE_VERSION} /usr/local/node
sudo ln -sf /usr/local/node/bin/node /usr/bin/node
sudo ln -sf /usr/local/node/bin/npm /usr/bin/npm
sudo ln -sf /usr/local/node/bin/npx /usr/bin/npx
