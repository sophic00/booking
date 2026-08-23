.PHONY: generate backend-build backend-run backend-test tidy migrate-up migrate-down migrate-status migrate-force

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

migrate-up:
	cd backend && go run cmd/migrate/main.go up

migrate-down:
	cd backend && go run cmd/migrate/main.go down

migrate-status:
	cd backend && go run cmd/migrate/main.go status

migrate-force:
	cd backend && go run cmd/migrate/main.go force $(v)
