.PHONY: help build run dev test clean docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the notification worker binary
	@echo "Building notification-worker..."
	@go build -o bin/notification-worker .
	@echo "✅ Build complete: bin/notification-worker"

run: ## Run the notification worker
	@echo "Running notification-worker..."
	@go run main.go

dev: ## Run with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
	@echo "Running in development mode with auto-reload..."
	@air

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "✅ Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t notification-worker:latest .
	@echo "✅ Docker image built: notification-worker:latest"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm --env-file .env notification-worker:latest

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "✅ Lint complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies downloaded"
