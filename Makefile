.PHONY: help dev infra-up infra-down 

help:
	@echo "FIRST CONNECT DOCKER NETWORK TO DEVCONTAINER NETWORK should be smth like this:"
	@echo "get network name: docker network ls"
	@echo "docker network connect go-saga-microservices-dev_default devcontainer"
	@echo "then start services in debug mode from vscode"
	@echo "Available commands:"
	@echo "  dev              - Start infrastructure and all services"
	@echo "  infra-up         - Start infrastructure services (Kafka, DBs)"
	@echo "  infra-down       - Stop infrastructure services"
	@echo "  prod          - Start production docker-compose stack"
	@echo "  prod-down        - Stop production docker-compose stack"
	@echo "  prod-restart     - Restart production docker-compose stack"

# DEV
dev: infra-up

infra-up iu:
	@echo "Starting infrastructure services..."
	GO_ENV=dev docker-compose -p go-saga-microservices-dev up -d --build

infra-down id:
	@echo "Stopping infrastructure services..."
	docker-compose -p go-saga-microservices-dev down

# PRODUCTION
.PHONY: prod prod-up prod-down

prod p:
	@echo "Building production stack..."
	GO_ENV=prod docker-compose --profile prod -p go-saga-microservices-prod up -d --build

prod-up pu:
	@echo "Starting production stack..."
	GO_ENV=prod docker-compose --profile prod -p go-saga-microservices-prod up -d

prod-down pd:
	@echo "Stopping production stack..."
	docker-compose --profile prod -p go-saga-microservices-prod down

# PROTOBUF
.PHONY: buf-install buf-gen

# Install buf CLI and protoc-gen-go if not present
buf-install:
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "Installing buf CLI..."; \
		GO111MODULE=on go install github.com/bufbuild/buf/cmd/buf@latest; \
	else \
		echo "buf already installed"; \
	fi

# Generate Go code from proto files using buf
buf-gen: buf-install
	@echo "Generating Go code from proto files using buf..."
	@buf generate
	@echo "go mod tidy in pkg/proto"
	@cd ./pkg/proto && go mod tidy
