# Go API

The backend built with Go, net/http and SQLite.

## Getting Started

To begin with this project:

### Run Redis

```bash
redis-server
```

### Configure the Application

This project uses a `.env` file for all configuration settings.

Uncomment and modify any settings if you need to change from their defaults.


### Starting the Application

```bash
PEBBLE_PASSWORD=xxx go run cmd/server/main.go
```

NOTE: The `PEBBLE_PASSWORD` variable is used for login. Ensure it is complex and securely stored in production.

### Test

```bash
go test -v ./...
```
