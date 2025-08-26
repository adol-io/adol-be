# Variables
APP_NAME=adol
DOCKER_IMAGE=adol-pos
MIGRATION_DIR=migrations
DB_URL=postgres://postgres:postgres@localhost:5432/adol_pos?sslmode=disable

# Build commands
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) cmd/api/main.go

.PHONY: build-docker
build-docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

# Run commands
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	go run cmd/api/main.go

.PHONY: run-docker
run-docker:
	@echo "Running with Docker Compose..."
	docker-compose up -d

.PHONY: stop-docker
stop-docker:
	@echo "Stopping Docker services..."
	docker-compose down

# Development commands
.PHONY: dev
dev:
	@echo "Starting development server with hot reload..."
	air

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	docker-compose down -v
	docker system prune -f

# Database commands
.PHONY: db-up
db-up:
	@echo "Running database migrations..."
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up

.PHONY: db-down
db-down:
	@echo "Rolling back database migrations..."
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down

.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up

.PHONY: db-create-migration
db-create-migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $$name

# Test commands
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# Linting and formatting
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Security
.PHONY: security
security:
	@echo "Running security checks..."
	gosec ./...

# API documentation
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/api/main.go

# Health check
.PHONY: health
health:
	@echo "Checking API health..."
	curl -f http://localhost:8080/health || exit 1

# Setup development environment
.PHONY: setup
setup:
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@echo "Then run: make db-up && make run"

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Production deployment
.PHONY: deploy
deploy:
	@echo "Deploying to production..."
	docker-compose -f docker-compose.prod.yml up -d

# Backup database
.PHONY: backup
backup:
	@echo "Creating database backup..."
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker exec adol_postgres pg_dump -U postgres adol_pos > backup_$$timestamp.sql; \
	echo "Backup created: backup_$$timestamp.sql"

# Load sample data
.PHONY: seed
seed:
	@echo "Loading sample data..."
	@echo "TODO: Implement sample data loading"

# Generate API client
.PHONY: generate-client
generate-client:
	@echo "Generating API client..."
	@echo "TODO: Implement API client generation"

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build              Build the application"
	@echo "  build-docker       Build Docker image"
	@echo "  run                Run the application locally"
	@echo "  run-docker         Run with Docker Compose"
	@echo "  stop-docker        Stop Docker services"
	@echo "  dev                Start development server with hot reload"
	@echo "  deps               Install dependencies"
	@echo "  clean              Clean up build artifacts and Docker containers"
	@echo "  db-up              Run database migrations"
	@echo "  db-down            Rollback database migrations"
	@echo "  db-reset           Reset database (down then up)"
	@echo "  db-create-migration Create new migration file"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  test-integration   Run integration tests"
	@echo "  lint               Run linter"
	@echo "  fmt                Format code"
	@echo "  security           Run security checks"
	@echo "  docs               Generate API documentation"
	@echo "  health             Check API health"
	@echo "  setup              Setup development environment"
	@echo "  install-tools      Install development tools"
	@echo "  deploy             Deploy to production"
	@echo "  backup             Backup database"
	@echo "  seed               Load sample data"
	@echo "  help               Show this help message"

# Default target
.DEFAULT_GOAL := help