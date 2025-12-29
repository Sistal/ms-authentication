.PHONY: help build run test docker-build docker-up docker-down clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o main ./cmd/api

run: ## Run the application locally
	go run cmd/api/main.go

test: ## Run tests
	go test -v ./...

docker-build: ## Build Docker image
	docker compose build

docker-up: ## Start all services with Docker Compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

clean: ## Clean build artifacts
	rm -f main
	go clean

deps: ## Download dependencies
	go mod download
	go mod tidy
