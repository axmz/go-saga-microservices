#!/bin/bash

set -e

echo "Setting up development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

echo "Go version: $(go version)"

# Install Air for hot reloading
echo "Installing Air for hot reloading..."
if ! command -v air &> /dev/null; then
    echo "Installing Air..."
    go install github.com/cosmtrek/air@latest
else
    echo "Air is already installed: $(air --version)"
fi

# Install golangci-lint for linting
echo "Installing golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
else
    echo "golangci-lint is already installed: $(golangci-lint --version)"
fi

# Install dependencies for each service
echo "Installing dependencies for all services..."

echo "Installing inventory service dependencies..."
cd services/inventory
go mod tidy
go mod download
cd ../..

echo "Installing order service dependencies..."
cd services/order
go mod tidy
go mod download
cd ../..

echo "Installing storefront service dependencies..."
cd services/storefront
go mod tidy
go mod download
cd ../..

echo "Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make help              - Show all available commands"
echo "  make dev               - Start infrastructure and all services"
echo "  make infra-up          - Start infrastructure services (Kafka, DBs)"
echo "  make services-start    - Start all Go services locally with Air"
echo "  make inventory         - Start inventory service with Air"
echo "  make order             - Start order service with Air"
echo "  make storefront        - Start storefront service with Air" 