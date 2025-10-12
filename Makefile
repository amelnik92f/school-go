.PHONY: help build run test clean install-deps migrate dev

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

migrate: ## Run database migrations
	go run cmd/migrate/main.go

# Scraping targets
scrape-statistics: ## Run the statistics scraper (production command)
	@echo "Running statistics scraper..."
	@go run cmd/scrape-statistics/main.go

scrape-school-details: ## Run the school details scraper (production command)
	@echo "Running school details scraper..."
	@echo "⚠️  This will take several hours to complete"
	@go run cmd/scrape-school-details/main.go

debug-school: ## Debug a single school (Hermann-Ehlers-Gymnasium) in headless mode
	@echo "Running debug scraper for single school..."
	@go run cmd/debug-school/main.go