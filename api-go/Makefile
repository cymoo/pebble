.PHONY: run build test

run:
	go run cmd/server/main.go

build:
	go build -o bin/api cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/