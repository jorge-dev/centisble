# Simple Makefile for a Go project

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

all: ## Build the application and test it
	build test

build: ## Build the application
	@echo "Building..."
	@go build -o main cmd/api/main.go

run: ## Run the application
	@go run cmd/api/main.go

d-run: ## Run the application in a Docker container
	@docker compose up --build -d

d-run-db: ## Run the database in a Docker container
	@docker compose up psql_centsible -d

d-down: ## Stop the application running in a Docker container
	@docker compose down

d-build: ## Build the application in a Docker container
	@docker compose build

d-clean: ## Clean the application running in a Docker container
	@docker compose down --rmi all --volumes --remove-orphans

format: ## Format the code
	@echo "Formatting..."
	@go fmt ./...

test: ## Run the tests
	@echo "Testing..."
	@go test ./... -v


itest: ## Run the integration tests
	@echo "Running integration tests..."
	@go test ./internal/database -v

clean: ## Clean the application
	@echo "Cleaning..."
	@rm -f main
	@echo "Cleaned."

watch: ## Watch the application for changes
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: help all build run test clean watch docker-run docker-down itest
