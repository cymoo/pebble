#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT_DIR")")
SOURCE_DIR="$PROJECT_ROOT/api-py"

# Load configuration
source "${SCRIPT_DIR}/deploy.conf"

VENV_DIR="$BACKEND_DIR/venv"

# Validate source
if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "Error: Python backend source not found: $SOURCE_DIR" >&2
    exit 1
fi

# Install Python dependencies
if ! command -v python3 &> /dev/null; then
    echo "Installing Python..."
    apt-get install -q -y python3 python3-pip python3-venv > /dev/null
fi

echo "Python version: $(python3 --version)"

# Create deployment directory
mkdir -p "$BACKEND_DIR"

# Copy source files
echo "Copying source files..."
rsync -a --delete \
    --exclude '__pycache__' \
    --exclude '*.pyc' \
    --exclude '.pytest_cache' \
    --exclude 'venv' \
    "$SOURCE_DIR/" "$BACKEND_DIR/"

# Create virtual environment
echo "Setting up virtual environment..."
python3 -m venv "$VENV_DIR"

# Install dependencies
echo "Installing Python dependencies..."
"$VENV_DIR/bin/pip" install --upgrade pip > /dev/null
"$VENV_DIR/bin/pip" install -r "$BACKEND_DIR/requirements.txt" > /dev/null

# Create start script with gunicorn
cat > "$BACKEND_DIR/start.sh" <<EOF
#!/bin/bash
export PORT=${API_PORT}
cd $BACKEND_DIR
exec $VENV_DIR/bin/gunicorn \
    --bind 0.0.0.0:\${PORT} \
    --workers 4 \
    --worker-class sync \
    --timeout 30 \
    --access-logfile - \
    --error-logfile - \
    app:app
EOF

chmod +x "$BACKEND_DIR/start.sh"
chown -R www-data:www-data "$BACKEND_DIR"

# Mark backend type
echo "python" > "$BACKEND_DIR/.backend_type"

echo "Python backend deployed to $BACKEND_DIR"