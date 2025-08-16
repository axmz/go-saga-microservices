.PHONY: help dev infra-up infra-down 

help:
	@echo "FIRST CONNECT DOCKER NETWORK TO DEVCONTAINER NETWORK should be smth like this:"
	@echo "get network name: docker network ls"
	@echo "docker network connect go-saga-microservices-dev_default devcontainer"
	@echo "then start services in debug mode from vscode"
	@echo "Available commands:"
	@echo "  dev              - Start development infra only (no microservices)"
	@echo "  infra-up         - Start development infrastructure services (Kafka, DBs)"
	@echo "  infra-down       - Stop infrastructure services"
	@echo "  prod             - Start production docker-compose stack (prebuilt images)"
	@echo "  prod-down        - Stop production docker-compose stack"
	@echo "  prod-restart     - Restart production docker-compose stack"
	@echo "  images-build     - Build service images (compose prod)"
	@echo "  images-push      - Push service images (compose prod)"

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
	TAG=$(TAG) GO_ENV=prod docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod up -d

prod-up pu:
	@echo "Starting production stack..."
	TAG=$(TAG) GO_ENV=prod docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod up -d --build

prod-down pd:
	@echo "Stopping production stack..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod down

# IMAGES
.PHONY: images-build images-push docker-login

REGISTRY?=axmz
TAG?=latest

docker-login:
	@echo "Logging in to Docker Hub"
	@docker login

images-build:
	@echo "Building service images from prod compose..."
	REGISTRY=$(REGISTRY) TAG=$(TAG) docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

images-push: docker-login
	@echo "Pushing images to Docker Hub from prod compose..."
	REGISTRY=$(REGISTRY) TAG=$(TAG) docker-compose -f docker-compose.yml -f docker-compose.prod.yml push

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
