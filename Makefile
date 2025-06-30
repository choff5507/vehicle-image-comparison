# Vehicle Image Comparison Makefile

.PHONY: build test clean install deps run-example help

# Default target
all: build

# Build the application
build:
	@echo "Building vehicle-compare..."
	go build -o vehicle-compare cmd/main.go

# Build for production with optimizations
build-prod:
	@echo "Building vehicle-compare for production..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o vehicle-compare cmd/main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./test -tags=integration

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golint ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f vehicle-compare
	go clean

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing vehicle-compare..."
	cp vehicle-compare $$GOPATH/bin/

# Run example comparison (requires test images)
run-example:
	@echo "Running example comparison..."
	@if [ -f "test/testdata/example1.jpg" ] && [ -f "test/testdata/example2.jpg" ]; then \
		./vehicle-compare -image1 test/testdata/example1.jpg -image2 test/testdata/example2.jpg -verbose; \
	else \
		echo "Example images not found. Please add example1.jpg and example2.jpg to test/testdata/"; \
	fi

# Create test data directory
setup-testdata:
	@echo "Creating test data directory..."
	mkdir -p test/testdata
	@echo "Please add test images (*.jpg) to test/testdata/ directory"

# Check dependencies
check-deps:
	@echo "Checking dependencies..."
	@command -v pkg-config >/dev/null 2>&1 || { echo >&2 "pkg-config is required but not installed."; exit 1; }
	@pkg-config --exists opencv4 || { echo >&2 "OpenCV 4 is required but not found."; exit 1; }
	@echo "All dependencies satisfied"

# Development setup
dev-setup: check-deps deps setup-testdata
	@echo "Development environment setup complete"

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-prod     - Build for production with optimizations"
	@echo "  deps           - Install Go dependencies"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  vet            - Vet code"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run-example    - Run example comparison"
	@echo "  setup-testdata - Create test data directory"
	@echo "  check-deps     - Check system dependencies"
	@echo "  dev-setup      - Complete development environment setup"
	@echo "  help           - Show this help message"