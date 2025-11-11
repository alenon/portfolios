.PHONY: help install migrate-up migrate-down migrate-create build run test docker-up docker-down docker-dev

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install: ## Install dependencies
	go mod download

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the backend server
	go run cmd/api/main.go

test: ## Run all tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker-up: ## Start production Docker containers
	docker-compose up -d

docker-down: ## Stop production Docker containers
	docker-compose down

docker-dev: ## Start development Docker containers with hot-reload
	docker-compose -f docker-compose.dev.yml up

docker-logs: ## View Docker container logs
	docker-compose logs -f

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

lint: ## Run linters
	golangci-lint run

format: ## Format code
	go fmt ./...
