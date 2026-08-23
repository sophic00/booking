.PHONY: generate test build run tidy lint migrate-up migrate-down migrate-status migrate-force

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

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-status:
	go run cmd/migrate/main.go status

migrate-force:
	go run cmd/migrate/main.go force $(v)
