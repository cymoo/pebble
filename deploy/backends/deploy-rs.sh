#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT_DIR")")
SOURCE_DIR="$PROJECT_ROOT/api-rs"
DEPLOY_DIR="/opt/pebble/backend"
API_PORT="${API_PORT:-8000}"

# Validate source
if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "Error: Rust backend source not found: $SOURCE_DIR" >&2
    exit 1
fi

# Install Rust if needed
if ! command -v cargo &> /dev/null; then
    echo "Installing Rust..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
    source "$HOME/.cargo/env"
    echo 'source $HOME/.cargo/env' >> /etc/profile
fi

echo "Rust version: $(rustc --version)"
echo "Cargo version: $(cargo --version)"

# Build
cd "$SOURCE_DIR"
echo "Building Rust binary (this may take a few minutes)..."
cargo build --release

# Verify binary
if [[ ! -f "target/release/pebble-api" ]] && [[ ! -f "target/release/api-rs" ]]; then
    # Try to find any binary in release
    BINARY=$(find target/release -maxdepth 1 -type f -executable | grep -v '\.d$' | head -n 1)
    if [[ -z "$BINARY" ]]; then
        echo "Error: Built binary not found in target/release/" >&2
        exit 1
    fi
else
    if [[ -f "target/release/pebble-api" ]]; then
        BINARY="target/release/pebble-api"
    else
        BINARY="target/release/api-rs"
    fi
fi

echo "Found binary: $BINARY"

# Deploy
mkdir -p "$DEPLOY_DIR"
cp "$BINARY" "$DEPLOY_DIR/pebble-api"
chmod +x "$DEPLOY_DIR/pebble-api"

# Copy config files if they exist
if [[ -f "config.toml" ]]; then
    cp config.toml "$DEPLOY_DIR/"
fi

# Create start script
cat > "$DEPLOY_DIR/start.sh" <<EOF
#!/bin/bash
export PORT=${API_PORT}
exec $DEPLOY_DIR/pebble-api
EOF

chmod +x "$DEPLOY_DIR/start.sh"
chown -R www-data:www-data "$DEPLOY_DIR"

echo "Rust backend deployed to $DEPLOY_DIR"