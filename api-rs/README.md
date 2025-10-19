# Rust API

The backend built with Rust, Axum and SQLite.

## Getting Started

To begin with this project:

### Run Redis

```bash
redis-server
```

### Configure the Application

It is preferably configured via environment variables, supporting multiple environment profiles.

- Base File: The application always loads `.env` as the base configuration.
- Environment-Specific Files (Optional):
  - `.env.dev` for debug mode or `.env.prod` for release mode
- Local Overrides (Optional):
  - A `.env.local` file will override any conflicting variables from the above files.

**Priority Order**:  
`.env` → `.env.dev` or `.env.prod` → `.env.local` (highest priority).

### Starting the Application

```bash
PEBBLE_PASSWORD=xxx cargo run
```

NOTE: The `PEBBLE_PASSWORD` variable is used for login. Ensure it is complex and securely stored in production.

### Auto Reloading

To start the server and auto-reload on code changes:

```bash
cargo install cargo-watch
cargo watch -w src -x run
```

### Database helper

SQLx offers a command-line tool for creating and managing databases as well as migrations.

```bash
cargo install sqlx-cli --no-default-features --features sqlite
```

If the database file does not exist, run the following commands to prepare the sqlite database for use:

```bash
sqlx database create
sqlx migrate run
```

For more usage details about sqlx, please refer to: <https://github.com/launchbadge/sqlx/tree/main/sqlx-cli>

### Test

```bash
RUST_TEST_THREADS=1 cargo test
```

## Deployment

Check the [deploy](../deploy) directory for a production deployment example, 
covering service management, password handling, HTTPS, and Nginx configuration, achieved with only two commands.
