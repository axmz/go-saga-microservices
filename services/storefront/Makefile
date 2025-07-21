.PHONY: dev build run test clean deps fmt lint

# Development with Air (hot reload)
dev:
	@echo "Starting storefront service with Air..."
	@air

# Build the application
build:
	@echo "Building storefront service..."
	@go build -o storefront-service ./cmd

# Run the application directly
run:
	@echo "Running storefront service..."
	@go run ./cmd

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f storefront-service
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