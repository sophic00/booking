.PHONY: generate test build run tidy lint

generate:
	sqlc generate

tidy:
	go mod tidy

build: generate
	go build -o bin/server cmd/server/main.go

run: generate
	go run cmd/server/main.go

test: generate
	go test -v -race ./...

lint:
	golangci-lint run ./...
