#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT_DIR")")
SOURCE_DIR="$PROJECT_ROOT/api-go"

# Load configuration
source "${SCRIPT_DIR}/deploy.conf"

# Validate source
if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "Error: Go backend source not found: $SOURCE_DIR" >&2
    exit 1
fi

# Install Go if needed
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    GO_VERSION="1.21.5"
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
fi

echo "Go version: $(go version)"

# Build
cd "$SOURCE_DIR"
echo "Building Go binary..."
go build -o pebble-api .

# Deploy
mkdir -p "$BACKEND_DIR"
cp pebble-api "$BACKEND_DIR/"
chmod +x "$BACKEND_DIR/pebble-api"

# Copy any config files if they exist
if [[ -f "config.yaml" ]]; then
    cp config.yaml "$BACKEND_DIR/"
fi

# Create start script
cat > "$BACKEND_DIR/start.sh" <<EOF
#!/bin/bash
export PORT=${API_PORT}
exec $BACKEND_DIR/pebble-api
EOF

chmod +x "$BACKEND_DIR/start.sh"
chown -R www-data:www-data "$BACKEND_DIR"

# Mark backend type
echo "go" > "$BACKEND_DIR/.backend_type"

echo "Go backend deployed to $BACKEND_DIR"