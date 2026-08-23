.PHONY: generate backend-build backend-run backend-test tidy migrate-up migrate-down migrate-status migrate-force frontend-install frontend-dev frontend-build frontend-lint

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

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-lint:
	cd frontend && npm run lint
