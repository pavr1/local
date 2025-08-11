# Session Service Makefile

# Variables
SERVICE_NAME := session-service
BINARY_NAME := session-service
BUILD_DIR := bin
MAIN_FILE := main.go

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
.PHONY: run
run:
	@echo "Running $(SERVICE_NAME)..."
	$(GOCMD) run $(MAIN_FILE)

# Start the service (build and run)
.PHONY: start
start: build
	@echo "Starting $(SERVICE_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Stop the service
.PHONY: stop
stop:
	@echo "Stopping $(SERVICE_NAME)..."
	@pkill -f $(BINARY_NAME) || true

# Restart the service
.PHONY: restart
restart: stop start

# Fresh start (clean, build, run)
.PHONY: fresh
fresh: clean build start

# Development mode (run with hot reload)
.PHONY: dev
dev:
	@echo "Starting $(SERVICE_NAME) in development mode..."
	$(GOCMD) run $(MAIN_FILE)

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	$(GOCMD) vet ./...

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	$(GOCMD) list -json -deps ./... | grep -i "vulnerability" || echo "No known vulnerabilities found"

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all ./...

# Install the service
.PHONY: install
install: build
	@echo "Installing $(SERVICE_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ || echo "Installation requires sudo privileges"

# Uninstall the service
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(SERVICE_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME) || echo "Uninstallation requires sudo privileges"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  deps         - Install dependencies"
	@echo "  run          - Run the application"
	@echo "  start        - Build and start the service"
	@echo "  stop         - Stop the service"
	@echo "  restart      - Restart the service"
	@echo "  fresh        - Clean, build, and start"
	@echo "  dev          - Run in development mode"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  security     - Check for security vulnerabilities"
	@echo "  docs         - Generate documentation"
	@echo "  install      - Install the service"
	@echo "  uninstall    - Uninstall the service"
	@echo "  help         - Show this help message" 