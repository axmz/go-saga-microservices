.PHONY: dev build run test clean deps fmt lint migrate-up migrate-down

# Development with Air (hot reload)
dev:
	@echo "Starting payment service with Air..."
	@air

# Build the application
build:
	@echo "Building payment service..."
	@go build -o payment-service ./cmd

# Run the application directly
run:
	@echo "Running payment service..."
	@go run ./cmd

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f payment-service
	@rm -rf tmp

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run