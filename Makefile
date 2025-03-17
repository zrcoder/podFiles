# Version information
VERSION ?= 0.0.8
IMAGE_NAME ?= podfiles
IMAGE_TAG ?= $(VERSION)

# Default target
.DEFAULT_GOAL := all

# Test the application and generate install script
all: test gen-script

# Generate the install script
gen-script:
	@echo "Generating install script..."
	@go run ./cmd/deploy/main.go

# Build and tag docker image
build-image:
	@echo "Building docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest
	@echo "Docker image $(IMAGE_NAME):$(IMAGE_TAG) built successfully."
# Build and tag docker image with local sdk
build-image-localsdk:
	@echo "Building docker image $(IMAGE_NAME):$(IMAGE_TAG) with local sdk..."
	@docker build -t $(IMAGE_NAME):localsdk-$(IMAGE_TAG) -f Dockerfile-localsdk .
	@echo "Docker image $(IMAGE_NAME):localsdk-$(IMAGE_TAG) built successfully."

# Build the application
build:
	@echo "Building binary..."
	@go build -o ./tmp/main ./cmd/podFiles

# Build the application with local sdk
build-localsdk:
	@echo "Building binary with local sdk..."
	@go build -tags local_sdk -o ./tmp/main ./cmd/podFiles

# Run the application locally
run:
	@echo "Running locally..."
	@DEV=1 KUBECONFIG=~/.kube/config go run ./cmd/podFiles

# Run with docker-compose, supporting both V1 and V2
docker-run:
	@echo "Starting with docker-compose..."
	@if command -v docker compose >/dev/null 2>&1; then \
		docker compose up --build; \
	else \
		docker-compose up --build; \
	fi

# Stop docker-compose services
docker-down:
	@echo "Stopping docker-compose..."
	@if command -v docker compose >/dev/null 2>&1; then \
		docker compose down; \
	else \
		docker-compose down; \
	fi

# Run tests with coverage
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f install.sh

# Live reload for development
watch:
	@if command -v air >/dev/null 2>&1; then \
		echo "Starting air..."; \
		air; \
	else \
		echo "Air is not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
		echo "Starting air..."; \
		air; \
	fi

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Image: $(IMAGE_NAME):$(IMAGE_TAG)"

# Help target
help:
	@echo "podFiles $(VERSION)"
	@echo
	@echo "Available targets:"
	@echo "  all          - Test and generate install script"
	@echo "  build        - Build the binary"
	@echo "  build-image  - Build docker image ($(IMAGE_NAME):$(IMAGE_TAG))"
	@echo "  run          - Run locally"
	@echo "  docker-run   - Run with docker-compose"
	@echo "  docker-down  - Stop docker-compose services"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  watch        - Run with live reload"
	@echo "  gen-script   - Generate install script"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help"

.PHONY: all build run test clean watch docker-run docker-down gen-script build-image help version
