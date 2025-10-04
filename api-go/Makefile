.PHONY: run build migrate test

run:
	go run cmd/server/main.go

build:
	go build -o bin/api cmd/server/main.go

migrate:
	go run cmd/server/main.go -migrate

test:
	go test -v ./...

clean:
	rm -rf bin/ data/