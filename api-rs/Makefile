.PHONY: run build live test clean help db-create db-migrate db-reset redis format lint check

# Default password for development (override with MOTE_PASSWORD env var)
MOTE_PASSWORD ?= foobar

# Build the application in debug mode
build:
	cargo build

# Build the application in release mode
build-release:
	cargo build --release

# Run the application
run:
	MOTE_PASSWORD=$(MOTE_PASSWORD) cargo run

# Run the application in release mode
run-release:
	MOTE_PASSWORD=$(MOTE_PASSWORD) cargo run --release

# Run with live reloading (requires cargo-watch)
live:
	@if ! command -v cargo-watch > /dev/null; then \
		echo "Installing cargo-watch..."; \
		cargo install cargo-watch; \
	fi
	MOTE_PASSWORD=$(MOTE_PASSWORD) cargo watch -w src -x run

# Run tests
test:
	RUST_TEST_THREADS=1 cargo test

# Run tests with output
test-verbose:
	RUST_TEST_THREADS=1 cargo test -- --nocapture

# Create database
db-create:
	@if ! command -v sqlx > /dev/null; then \
		echo "Installing sqlx-cli..."; \
		cargo install sqlx-cli --no-default-features --features sqlite; \
	fi
	sqlx database create

# Run database migrations
db-migrate:
	@if ! command -v sqlx > /dev/null; then \
		echo "Installing sqlx-cli..."; \
		cargo install sqlx-cli --no-default-features --features sqlite; \
	fi
	sqlx migrate run

# Reset database (drop, create, and migrate)
db-reset:
	@if ! command -v sqlx > /dev/null; then \
		echo "Installing sqlx-cli..."; \
		cargo install sqlx-cli --no-default-features --features sqlite; \
	fi
	sqlx database drop -y || true
	sqlx database create
	sqlx migrate run

# Start Redis server
redis:
	redis-server

# Format code
format:
	cargo fmt

# Check formatting without making changes
format-check:
	cargo fmt -- --check

# Run clippy linter
lint:
	cargo clippy -- -D warnings

# Check code without building
check:
	cargo check

# Clean build artifacts
clean:
	cargo clean
	rm -rf target/

# Update dependencies
update:
	cargo update

# Check for outdated dependencies
outdated:
	@if ! command -v cargo-outdated > /dev/null; then \
		echo "Installing cargo-outdated..."; \
		cargo install cargo-outdated; \
	fi
	cargo outdated

# Install development tools
install-tools:
	cargo install cargo-watch
	cargo install sqlx-cli --no-default-features --features sqlite
	cargo install cargo-outdated

help:
	@echo "Makefile commands:"
	@echo "  run            - Run the application in debug mode"
	@echo "  run-release    - Run the application in release mode"
	@echo "  build          - Build the application in debug mode"
	@echo "  build-release  - Build the application in release mode"
	@echo "  live           - Run with auto-reload on code changes"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with output"
	@echo "  db-create      - Create the SQLite database"
	@echo "  db-migrate     - Run database migrations"
	@echo "  db-reset       - Reset database (drop, create, migrate)"
	@echo "  redis          - Start Redis server"
	@echo "  format         - Format code with rustfmt"
	@echo "  format-check   - Check code formatting"
	@echo "  lint           - Run clippy linter"
	@echo "  check          - Check code without building"
	@echo "  clean          - Clean build artifacts"
	@echo "  update         - Update dependencies"
	@echo "  outdated       - Check for outdated dependencies"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help message"
