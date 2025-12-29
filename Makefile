.PHONY: build run check test clean docker docker-run fmt lint tidy

# Variables
BINARY=jobradar
VERSION=1.0.0
BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build the binary
build:
	go build $(BUILD_FLAGS) -o $(BINARY) ./cmd/jobradar

# Run the scheduler
run:
	go run ./cmd/jobradar run

# Check for new jobs immediately
check:
	go run ./cmd/jobradar check

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f jobradar.db
	rm -f coverage.out coverage.html

# Build Docker image
docker:
	docker build -t jobradar:$(VERSION) -f docker/Dockerfile .

# Run with Docker Compose
docker-run:
	docker-compose -f docker/docker-compose.yml up -d

# Stop Docker Compose
docker-stop:
	docker-compose -f docker/docker-compose.yml down

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Validate configuration
validate:
	go run ./cmd/jobradar validate

# View history
history:
	go run ./cmd/jobradar history

# View stats
stats:
	go run ./cmd/jobradar stats

# Install dependencies
deps:
	go mod download

# All: format, lint, test, build
all: fmt lint test build

