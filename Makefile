.PHONY: build test run clean docker-build docker-run

# Build the application
build:
	go build -o pub-sub-system .

# Run tests
test:
	go test ./...

# Run the application
run:
	go run main.go

# Clean build artifacts
clean:
	rm -f pub-sub-system
	go clean

# Build Docker image
docker-build:
	docker build -t pub-sub-system .

# Run Docker container
docker-run:
	docker run -p 8080:8080 pub-sub-system

# Run with docker-compose
docker-compose-up:
	docker-compose up --build

# Stop docker-compose
docker-compose-down:
	docker-compose down

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

# Run all checks
check: fmt lint test build

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  run            - Run the application locally"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-compose-up   - Run with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  check          - Run all checks"
	@echo "  help           - Show this help"
