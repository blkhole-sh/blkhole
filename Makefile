include .env
export

dev:
	@cd frontend && bun dev & cd backend && air

build:
	@cd backend && go build -o server .
	@cd frontend && bun run build

test:
	@cd backend && go test ./...

clean:
	@cd backend && rm -f server
	@cd frontend && rm -rf .output dist

docker:
	docker-compose up --build

docker-down:
	docker-compose down

.PHONY: dev build test clean docker docker-down
