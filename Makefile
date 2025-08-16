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
	@echo "  prod             - Start production docker-compose stack for AMD64 (prebuilt images)"
	@echo "  prod-local       - Start production docker-compose stack for ARM64 (local build)"
	@echo "  prod-down        - Stop production docker-compose stack"
	@echo "  prod-restart     - Restart production docker-compose stack"
	@echo "  images-build     - Build service images for AMD64 (compose prod)"
	@echo "  images-push      - Push service images for AMD64 (compose prod)"
	@echo "  images-build-local - Build images for current platform locally"
	@echo "  images-build-multi - Build multi-platform images (AMD64 + ARM64) - build only"
	@echo "  images-push-multi  - Build and push multi-platform images (AMD64 + ARM64)"

# DEV
dev: infra-up

infra-up iu:
	@echo "Starting infrastructure services..."
	GO_ENV=dev docker compose -p go-saga-microservices-dev up -d --build

infra-down id:
	@echo "Stopping infrastructure services..."
	docker compose -p go-saga-microservices-dev down

# PRODUCTION
.PHONY: prod prod-local prod-up prod-down

prod p:
	@echo "Pulling images for production stack (AMD64)..."
	DOCKER_PLATFORM=linux/amd64 TAG=$(TAG) GO_ENV=prod docker compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod pull
	@echo "Starting production stack from pulled images (no build)..."
	DOCKER_PLATFORM=linux/amd64 TAG=$(TAG) GO_ENV=prod docker compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod up -d

prod-local pl:
	@echo "Starting production stack locally (ARM64)..."
	DOCKER_PLATFORM=linux/arm64 TAG=$(TAG) GO_ENV=prod docker compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod up -d --build

prod-up pu:
	@echo "Starting production stack (AMD64)..."
	DOCKER_PLATFORM=linux/amd64 TAG=$(TAG) GO_ENV=prod docker compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod up -d --build

prod-down pd:
	@echo "Stopping production stack..."
	docker compose -f docker-compose.yml -f docker-compose.prod.yml -p go-saga-microservices-prod down

# IMAGES
.PHONY: images-build images-push images-build-local images-build-multi images-push-multi docker-login buildx-setup

REGISTRY?=axmz
TAG?=latest
SERVICES=inventory order storefront payment

buildx-setup:
	@echo "Setting up Docker Buildx for multi-platform builds..."
	@docker buildx create --use --name multiplatform --driver docker-container 2>/dev/null || echo "Buildx builder already exists"
	@docker buildx inspect --bootstrap

docker-login:
	@echo "Logging in to Docker Hub"
	@docker login

images-build:
	@echo "Building service images from prod compose (AMD64)..."
	DOCKER_PLATFORM=linux/amd64 REGISTRY=$(REGISTRY) TAG=$(TAG) docker compose -f docker-compose.yml -f docker-compose.prod.yml build

images-push: docker-login
	@echo "Pushing images to Docker Hub from prod compose..."
	DOCKER_PLATFORM=linux/amd64 REGISTRY=$(REGISTRY) TAG=$(TAG) docker compose -f docker-compose.yml -f docker-compose.prod.yml push

images-build-local: buildx-setup
	@echo "Building images for current platform ($(shell uname -m))..."
	@for service in $(SERVICES); do \
		echo "Building $$service for current platform..."; \
		docker buildx build \
			--file ./infra/service.Dockerfile \
			--build-arg SERVICE=$$service \
			--build-arg MAIN=main.go \
			--tag $(REGISTRY)/go-saga-microservices-$$service:$(TAG) \
			--load \
			.; \
	done

images-build-multi: buildx-setup
	@echo "Building multi-platform images (AMD64 + ARM64) - Note: Cannot load locally, will build only..."
	@for service in $(SERVICES); do \
		echo "Building $$service for multiple platforms..."; \
		docker buildx build \
			--platform linux/amd64,linux/arm64 \
			--file ./infra/service.Dockerfile \
			--build-arg SERVICE=$$service \
			--build-arg MAIN=main.go \
			--tag $(REGISTRY)/go-saga-microservices-$$service:$(TAG) \
			.; \
	done

images-push-multi: buildx-setup docker-login
	@echo "Building and pushing multi-platform images (AMD64 + ARM64)..."
	@for service in $(SERVICES); do \
		echo "Building and pushing $$service for multiple platforms..."; \
		docker buildx build \
			--platform linux/amd64,linux/arm64 \
			--file ./infra/service.Dockerfile \
			--build-arg SERVICE=$$service \
			--build-arg MAIN=main.go \
			--tag $(REGISTRY)/go-saga-microservices-$$service:$(TAG) \
			--push \
			.; \
	done

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
