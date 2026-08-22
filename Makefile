.PHONY: generate backend-build backend-run backend-test tidy

generate:
	cd backend && sqlc generate

backend-build:
	cd backend && go build -o bin/server cmd/server/main.go

backend-run:
	cd backend && go run cmd/server/main.go

backend-test:
	cd backend && go test -v -race ./...

tidy:
	cd backend && go mod tidy
