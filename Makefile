.PHONY: help build run test clean docker-up docker-down lint format

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Variables
BINARY_NAME := taskS3
BUILD_DIR := bin
DOCKER_COMPOSE := docker-compose.yml
GO_MODULE := taskS3
MAIN_PACKAGE := ./cmd/app

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "$(YELLOW)TaskS3 - Image Management with S3$(NC)"
	@echo "$(YELLOW)=====================================$(NC)"
	@echo ""
	@echo "$(GREEN)Available commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-25s$(NC) %s\n", $$1, $$2}'

# Development commands
build: ## Build the application
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run: ## Run the application locally
	@echo "$(GREEN)Running $(BINARY_NAME) locally...$(NC)"
	@export S3_ENDPOINT=http://localhost:9000 && \
	 export S3_ACCESS_KEY_ID=minioadmin && \
	 export S3_SECRET_ACCESS_KEY=minioadmin && \
	 go run $(MAIN_PACKAGE)

dev: ## Run with hot reload (requires air)
	@if ! command -v air &> /dev/null; then \
		echo "$(YELLOW)Installing air for hot reload...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s; \
	fi
	@export S3_ENDPOINT=http://localhost:9000 && \
	 export S3_ACCESS_KEY_ID=minioadmin && \
	 export S3_SECRET_ACCESS_KEY=minioadmin && \
	 air

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf uploads/processed/*
	@go clean
	@echo "$(GREEN)Clean complete$(NC)"

# Code quality
lint: ## Run linters
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
	fi
	@golangci-lint run ./...

format: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@if command -v goimports &> /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)Install goimports for better formatting: go install golang.org/x/tools/cmd/goimports@latest$(NC)"; \
	fi

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

# Dependency management
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify

deps-update: ## Update all dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy

# Docker commands
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):latest .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 8080:8080 -p 9000:9000 -p 9001:9001 $(BINARY_NAME):latest

docker-up: ## Start Docker Compose services
	@echo "$(GREEN)Starting Docker Compose services...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started.$(NC)"
	@echo "$(YELLOW)App: http://localhost:8080$(NC)"
	@echo "$(YELLOW)MinIO Console: http://localhost:9001 (minioadmin/minioadmin)$(NC)"

docker-down: ## Stop Docker Compose services
	@echo "$(GREEN)Stopping Docker Compose services...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) down
	@echo "$(GREEN)Services stopped$(NC)"

docker-logs: ## View Docker Compose logs
	@docker-compose -f $(DOCKER_COMPOSE) logs -f

docker-clean: ## Remove Docker containers and images
	@echo "$(GREEN)Cleaning Docker containers and images...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) down -v --rmi all
	@docker system prune -f
	@echo "$(GREEN)Docker cleanup complete$(NC)"

# MinIO setup
minio-up: ## Start MinIO only
	@echo "$(GREEN)Starting MinIO...$(NC)"
	@docker run -d \
		-p 9000:9000 \
		-p 9001:9001 \
		--name minio \
		-e "MINIO_ROOT_USER=minioadmin" \
		-e "MINIO_ROOT_PASSWORD=minioadmin" \
		-v /tmp/minio_data:/data \
		minio/minio server /data --console-address ":9001"
	@echo "$(GREEN)MinIO started$(NC)"
	@echo "$(YELLOW)Console: http://localhost:9001$(NC)"
	@echo "$(YELLOW)Access Key: minioadmin$(NC)"
	@echo "$(YELLOW)Secret Key: minioadmin$(NC)"

minio-down: ## Stop MinIO
	@echo "$(GREEN)Stopping MinIO...$(NC)"
	@docker stop minio 2>/dev/null || true
	@docker rm minio 2>/dev/null || true
	@echo "$(GREEN)MinIO stopped$(NC)"

# Setup commands
setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@echo "$(YELLOW)1. Installing Go tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(YELLOW)2. Creating necessary directories...$(NC)"
	@mkdir -p uploads/images uploads/processed uploads/processed/moved
	@mkdir -p web/static/css web/static/js
	@echo "$(YELLOW)3. Downloading dependencies...$(NC)"
	@go mod download
	@echo "$(GREEN)Setup complete!$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Start MinIO: $(YELLOW)make minio-up$(NC)"
	@echo "  2. Run the app: $(YELLOW)make run$(NC)"
	@echo "  3. Or use Docker: $(YELLOW)make docker-up$(NC)"

# Check commands
check: lint vet test ## Run all checks (lint, vet, test)

version: ## Show version information
	@echo "$(GREEN)Go version:$(NC)"
	@go version
	@echo ""
	@echo "$(GREEN)Module information:$(NC)"
	@go list -m -f '{{.Path}} {{.Version}}' all | head -10
	@echo ""
	@echo "$(GREEN)Docker version:$(NC)"
	@docker --version 2>/dev/null || echo "Docker not installed"
	@docker-compose --version 2>/dev/null || echo "Docker Compose not installed"

# Utility commands
generate-images: ## Generate test images
	@echo "$(GREEN)Generating test images...$(NC)"
	@if command -v convert &> /dev/null; then \
		for i in {1..3}; do \
			convert -size 800x600 gradient:blue-red -pointsize 40 -fill white \
				-draw "text 50,300 'Test Image $$i'" \
				-draw "text 50,350 'TaskS3 Project'" \
				uploads/images/test$$i.jpg; \
			echo "Created test$$i.jpg"; \
		done; \
	else \
		echo "$(YELLOW)ImageMagick not found. Creating placeholder files...$(NC)"; \
		for i in {1..3}; do \
			echo "Test image $$i" > uploads/images/test$$i.jpg; \
			echo "Test PNG $$i" > uploads/images/test$$i.png; \
		done; \
	fi
	@echo "$(GREEN)Test images generated in uploads/images/$(NC)"

process-images: ## Process test images
	@echo "$(GREEN)Processing test images...$(NC)"
	@curl -s -X POST http://localhost:8080/api/process
	@echo ""
	@echo "$(GREEN)Images processed. Check uploads/processed/$(NC)"

list-images: ## List images in S3
	@echo "$(GREEN)Listing images from S3...$(NC)"
	@curl -s http://localhost:8080/api/images | python3 -m json.tool 2>/dev/null || \
		curl -s http://localhost:8080/api/images

health: ## Check application health
	@echo "$(GREEN)Checking application health...$(NC)"
	@curl -s http://localhost:8080/health || echo "$(RED)Application not running$(NC)"

# Production build
prod-build: ## Build production binary (stripped, optimized)
	@echo "$(GREEN)Building production binary...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -X main.buildVersion=$$(git describe --tags 2>/dev/null || echo 'dev')" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@echo "$(GREEN)Production build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(NC)"
