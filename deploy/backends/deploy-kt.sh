#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT_DIR")")
SOURCE_DIR="$PROJECT_ROOT/api-kt"
DEPLOY_DIR="/opt/pebble/backend"
API_PORT="${API_PORT:-8000}"

# Validate source
if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "Error: Kotlin backend source not found: $SOURCE_DIR" >&2
    exit 1
fi

# Install Java if needed
if ! command -v java &> /dev/null; then
    echo "Installing Java..."
    apt-get install -q -y openjdk-17-jdk > /dev/null
fi

echo "Java version: $(java -version 2>&1 | head -n 1)"

# Install Maven if needed
if ! command -v mvn &> /dev/null; then
    echo "Installing Maven..."
    apt-get install -q -y maven > /dev/null
fi

echo "Maven version: $(mvn -version | head -n 1)"

# Build with Maven
cd "$SOURCE_DIR"
echo "Building with Maven..."
mvn clean package -DskipTests -q

# Find the built JAR
JAR_FILE=$(find target -name "*.jar" -not -name "*-sources.jar" -not -name "*-javadoc.jar" | head -n 1)

if [[ -z "$JAR_FILE" || ! -f "$JAR_FILE" ]]; then
    echo "Error: Built JAR file not found in target/" >&2
    exit 1
fi

echo "Found JAR: $JAR_FILE"

# Deploy
mkdir -p "$DEPLOY_DIR"
cp "$JAR_FILE" "$DEPLOY_DIR/pebble-api.jar"

# Copy application properties if exists
if [[ -f "src/main/resources/application.properties" ]]; then
    mkdir -p "$DEPLOY_DIR/config"
    cp "src/main/resources/application.properties" "$DEPLOY_DIR/config/"
fi

# Create start script
cat > "$DEPLOY_DIR/start.sh" <<EOF
#!/bin/bash
export PORT=${API_PORT}
export SERVER_PORT=\${PORT}
exec java -jar $DEPLOY_DIR/pebble-api.jar \
    --server.port=\${PORT}
EOF

chmod +x "$DEPLOY_DIR/start.sh"
chown -R www-data:www-data "$DEPLOY_DIR"

echo "Kotlin backend deployed to $DEPLOY_DIR"