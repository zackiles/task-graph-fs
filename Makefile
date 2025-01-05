.PHONY: build test clean install run

# Binary name
BINARY_NAME=tgfs
# Build directory
BUILD_DIR=bin

# Go build flags
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
BUILD_FLAGS=-trimpath -ldflags="-s -w"

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tgfs
	@echo "Done! Binary is in $(BUILD_DIR)/$(BINARY_NAME)"

# Install the application globally
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(BUILD_FLAGS) ./cmd/tgfs
	@echo "Done! Binary installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Run the application (builds first)
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Dev build - faster compilation, includes debug info
dev:
	@echo "Building dev version..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tgfs
	@echo "Done! Binary is in $(BUILD_DIR)/$(BINARY_NAME)"

# Watch for changes and rebuild (requires watchexec)
watch:
	@if command -v watchexec >/dev/null 2>&1; then \
		echo "Watching for changes..."; \
		watchexec -e go -r "make dev"; \
	else \
		echo "Error: watchexec is not installed. Install with 'brew install watchexec'"; \
		exit 1; \
	fi 