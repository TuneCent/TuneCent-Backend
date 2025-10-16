.PHONY: help build run test clean install migrate

# Variables
APP_NAME=tunecent-backend
GO_FILES=$(shell find . -name '*.go' -type f)
MAIN_PATH=./cmd/server/main_complete.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_PATH)

dev: ## Run in development mode with hot reload (requires air)
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

test: ## Run tests
	@echo "Running tests..."
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run $(MAIN_PATH) migrate

db-setup: ## Setup MySQL database
	@echo "Setting up database..."
	mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS tunecent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	@echo "Database created. Running schema..."
	mysql -u root -p tunecent_db < schema.sql

generate-bindings: ## Generate contract bindings (requires abigen)
	@echo "Generating contract bindings..."
	@if ! command -v abigen > /dev/null; then \
		echo "Error: abigen not found. Install it from go-ethereum"; \
		exit 1; \
	fi
	@mkdir -p internal/blockchain/contracts
	abigen --sol=../TuneCent-SmartContract/src/MusicRegistry.sol --pkg=contracts --out=internal/blockchain/contracts/MusicRegistry.go
	abigen --sol=../TuneCent-SmartContract/src/RoyaltyDistributor.sol --pkg=contracts --out=internal/blockchain/contracts/RoyaltyDistributor.go
	abigen --sol=../TuneCent-SmartContract/src/CrowdfundingPool.sol --pkg=contracts --out=internal/blockchain/contracts/CrowdfundingPool.go
	abigen --sol=../TuneCent-SmartContract/src/ReputationScore.sol --pkg=contracts --out=internal/blockchain/contracts/ReputationScore.go

swagger: ## Generate swagger documentation
	@if ! command -v swag > /dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/server/main_complete.go

all: clean install build ## Clean, install deps and build

prod: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/$(APP_NAME) $(MAIN_PATH)

check: fmt vet lint test ## Run all checks

.DEFAULT_GOAL := help
