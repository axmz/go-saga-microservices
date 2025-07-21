# Development Setup

This guide explains how to set up and run the microservices locally for development.

## Overview

The development environment is split into two parts:
- **Infrastructure services** (Kafka, PostgreSQL databases) run in Docker containers
- **Go services** (inventory, order, storefront) run locally with Air for hot reloading

## Prerequisites

- Go 1.24.4 or later
- Docker and Docker Compose
- Make

## Initial Setup

1. **Run the setup script** to install development dependencies:
   ```bash
   ./setup-dev.sh
   ```

   This will install:
   - Air (for hot reloading)
   - golangci-lint (for code linting)
   - Dependencies for all services

## Development Workflow

### Quick Start

To start everything for development:
```bash
make dev
```

This will:
1. Start infrastructure services (Kafka, databases) in Docker
2. Start all Go services locally with Air for hot reloading

### Individual Commands

#### Infrastructure Management
```bash
# Start infrastructure services (Kafka, databases)
make infra-up

# Stop infrastructure services
make infra-down

# Restart infrastructure services
make infra-restart
```

#### Service Management
```bash
# Start all Go services with Air
make services-start

# Stop all Go services
make services-stop

# Restart all Go services
make services-restart

# Start individual services
make inventory    # Start inventory service
make order        # Start order service
make storefront   # Start storefront service
```

#### Utility Commands
```bash
# Show all available commands
make help

# Clean build artifacts
make clean
```

## Service URLs

When running locally, services will be available at:
- **Storefront Service**: http://localhost:8080
- **Inventory Service**: http://localhost:8081
- **Order Service**: http://localhost:8082

## Service Communication

Services communicate using the same hostnames as in the container environment:
- `http://inventory-service:8080`
- `http://order-service:8080`
- `http://storefront-service:8080`

This is achieved by adding entries to your `/etc/hosts` file or using a local DNS resolver.

## Database Connections

Databases are accessible at:
- **Inventory DB**: localhost:5433
- **Order DB**: localhost:5434
- **Kafka**: localhost:9092

## Hot Reloading

Each Go service uses Air for hot reloading. When you modify any `.go` file, the service will automatically rebuild and restart.

## Individual Service Development

You can work on individual services by navigating to their directories:

```bash
# Inventory service
cd services/inventory
make dev

# Order service
cd services/order
make dev

# Storefront service
cd services/storefront
make dev
```

## Available Make Commands per Service

Each service has its own Makefile with commands:

```bash
make dev      # Start with Air (hot reload)
make build    # Build the service
make run      # Run directly with go run
make test     # Run tests
make clean    # Clean build artifacts
make deps     # Install dependencies
make fmt      # Format code
make lint     # Lint code
```

## Troubleshooting

### Port Conflicts
If you get port conflicts, check if services are already running:
```bash
# Check for running Air processes
ps aux | grep air

# Kill all Air processes
pkill -f air
```

### Database Issues
If databases aren't starting properly:
```bash
# Check Docker containers
docker ps

# Check logs
docker-compose logs
```

### Air Issues
If Air isn't working properly:
```bash
# Reinstall Air
go install github.com/cosmtrek/air@latest

# Check Air version
air --version
```

## Production vs Development

- **Development**: Uses `docker-compose.yml` for infrastructure + local Go services
- **Production**: Uses `docker-compose.prod.yml` for all services in containers

## File Structure

```
go-saga-microservices/
├── docker-compose.yml         # Infrastructure only (development)
├── docker-compose.prod.yml    # All services (production)
├── Makefile                   # Root development commands
├── setup-dev.sh               # Development setup script
├── services/
│   ├── inventory/
│   │   ├── .air.toml        # Air config
│   │   ├── Makefile         # Service-specific commands
│   │   └── cmd/main.go      # Service entry point
│   ├── order/
│   │   ├── .air.toml        # Air config
│   │   ├── Makefile         # Service-specific commands
│   │   └── cmd/main.go      # Service entry point
│   └── storefront/
│       ├── .air.toml        # Air config
│       ├── Makefile         # Service-specific commands
│       └── cmd/main.go      # Service entry point
``` 