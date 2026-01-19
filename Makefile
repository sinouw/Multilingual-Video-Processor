.PHONY: help test test-coverage lint build clean deploy run-local install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make build          - Build binary"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deploy         - Deploy to Cloud Functions"
	@echo "  make run-local      - Run locally with Functions Framework"
	@echo "  make install-tools  - Install development tools"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Build binary
build:
	@echo "Building binary..."
	@go build -o function ./cmd/cloudfunction
	@echo "Binary built: ./function"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f function
	@rm -f coverage.out coverage.html
	@rm -f gosec-report.json gosec-report.sarif
	@go clean ./...

# Deploy to Cloud Functions
deploy:
	@echo "Deploying to Cloud Functions..."
	@./deploy.sh

# Run locally with Functions Framework
run-local:
	@echo "Starting local server..."
	@echo "Make sure Functions Framework is installed: go install github.com/GoogleCloudPlatform/functions-framework-go/cmd/functions-framework@latest"
	@functions-framework --target=TranslateVideo --port=8080

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/GoogleCloudPlatform/functions-framework-go/cmd/functions-framework@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"
