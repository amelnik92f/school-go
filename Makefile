.PHONY: help build run test clean install-deps migrate dev docker-build docker-up docker-down docker-logs docker-restart

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/schools-be cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

dev: ## Run in development mode with hot reload (requires air: go install github.com/air-verse/air@latest)
	air

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

# Docker commands
docker-build: ## Build Docker image
	docker-compose build

docker-up: ## Start Docker containers in detached mode
	docker-compose up -d

docker-down: ## Stop and remove Docker containers
	docker-compose down

docker-logs: ## Show Docker container logs
	docker-compose logs -f schools-api

docker-restart: ## Restart Docker containers
	docker-compose restart

docker-rebuild: ## Rebuild and restart Docker containers
	docker-compose up -d --build