.PHONY: help install migrate-up migrate-down migrate-create build build-cli run test docker-up docker-down docker-dev install-cli

# CLI build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.Version=$(VERSION)' \
           -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.BuildDate=$(BUILD_DATE)' \
           -X 'github.com/lenon/portfolios/cmd/portfolios/cmd.GitCommit=$(GIT_COMMIT)'

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

build-cli: ## Build the CLI tool
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/portfolios ./cmd/portfolios
	@echo "CLI built: bin/portfolios (version: $(VERSION))"

install-cli: build-cli ## Install CLI to /usr/local/bin
	@echo "Installing CLI to /usr/local/bin..."
	sudo cp bin/portfolios /usr/local/bin/portfolios
	@echo "Installed! Run 'portfolios --help' to get started"

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
