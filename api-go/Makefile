.PHONY: run build live test clean help upgradeable tidy

build:
	go build -o=/tmp/bin/mote ./cmd/server

run: build
	/tmp/bin/mote

live:
	APP_ENV=dev go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "/tmp/bin/mote" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"

test:
	go test -v -race ./...

clean:
	rm -rf bin/

upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest

tidy:
	go mod tidy -v
	go fmt ./...

help:
	@echo "Makefile commands:"
	@echo "  live   - Run the API server with live reloading on file changes"
	@echo "  run    - Run the API server"
	@echo "  build  - Build the API server binary"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean up build artifacts"
	@echo "  upgradeable - List direct dependencies with available upgrades"
	@echo "  tidy   - Tidy modifies and formats .go files
