.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o order-service ./cmd

# Run the application
run:
	go run ./cmd

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f order-service

# Build Docker image
docker-build:
	docker build -t order-service .

# Run Docker container
docker-run:
	docker run -p 8080:8080 order-service

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Database migration (example - you'd need a migration tool)
migrate-up:
	# Example: migrate -path migrations -database "postgres://order:orderpass@localhost:5434/order?sslmode=disable" up

migrate-down:
	# Example: migrate -path migrations -database "postgres://order:orderpass@localhost:5434/order?sslmode=disable" down 