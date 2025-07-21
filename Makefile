.PHONY: help dev infra-up infra-down infra-restart services-start services-stop services-restart inventory order storefront clean

# Default target
help:
	@echo "Available commands:"
	@echo "  dev              - Start infrastructure and all services"
	@echo "  infra-up         - Start infrastructure services (Kafka, DBs)"
	@echo "  infra-down       - Stop infrastructure services"
	@echo "  infra-restart    - Restart infrastructure services"
	@echo "  services-start   - Start all Go services locally with Air"
	@echo "  services-stop    - Stop all Go services"
	@echo "  services-restart - Restart all Go services"
	@echo "  inventory        - Start inventory service with Air"
	@echo "  order            - Start order service with Air"
	@echo "  storefront       - Start storefront service with Air"
	@echo "  clean            - Clean build artifacts"

# Start everything for development
dev: infra-up services-start

# Infrastructure management
infra-up iu:
	@echo "Starting infrastructure services..."
	docker-compose -f docker-compose.dev.yml up -d

infra-down id:
	@echo "Stopping infrastructure services..."
	docker-compose -f docker-compose.dev.yml down

infra-restart ir: infra-down infra-up

# Services management
services-start ss: inventory order storefront

services-stop st:
	@echo "Stopping all services..."
	@pkill -f "air" || true
	@pkill -f "inventory-service" || true
	@pkill -f "order-service" || true
	@pkill -f "storefront-service" || true

services-restart sr: services-stop services-start

# Individual service commands
inventory i:
	@echo "Starting inventory service with Air..."
	@cd services/inventory && make dev

order o:
	@echo "Starting order service with Air..."
	@cd services/order && make run

storefront s:
	@echo "Starting storefront service with Air..."
	@cd services/storefront && make dev

# Clean build artifacts
clean c:
	@echo "Cleaning build artifacts..."
	@find . -name "tmp" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "*.exe" -delete 2>/dev/null || true
	@find . -name "*.out" -delete 2>/dev/null || true 