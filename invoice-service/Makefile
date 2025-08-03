# Variables
BUILD_DIR=bin
SERVICE_NAME=invoice-service
BINARY_NAME=invoice-service
MAIN_FILE=main.go

# Docker variables
DOCKER_IMAGE=icecream-invoice-service

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build targets
.PHONY: all build clean test deps run docker-build docker-run start stop logs fresh

all: deps build

build:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests for $(SERVICE_NAME)..."
	$(GOTEST) -v ./...
	@echo "✅ Tests completed"

deps:
	@echo "Downloading dependencies for $(SERVICE_NAME)..."
	$(GOMOD) download
	$(GOMOD) tidy

run: build
	@echo "Running $(SERVICE_NAME)..."
	 ./$(BUILD_DIR)/$(BINARY_NAME)

# Docker commands
docker-build:
	@echo "Building Docker image for $(SERVICE_NAME)..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "Running $(SERVICE_NAME) in Docker..."
	docker run --rm -p 8085:8085 --name $(SERVICE_NAME) $(DOCKER_IMAGE)

start: docker-run

stop:
	@echo "Stopping $(SERVICE_NAME)..."
	docker stop $(SERVICE_NAME) || true

logs:
	@echo "Showing logs for $(SERVICE_NAME)..."
	docker logs $(SERVICE_NAME)

fresh: clean deps build 