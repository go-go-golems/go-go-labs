# Meshtastic TUI Makefile

.PHONY: build test clean proto-gen proto-clean format lint help

# Build configuration
BINARY_NAME=meshtastic-tui
CMD_DIR=.
BUILD_DIR=./build
PROTO_DIR=../../../meshtastic-protobufs/meshtastic
PB_DIR=./pkg/pb

# Go configuration
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@mkdir -p $(PB_DIR)
	@if [ ! -f $(PROTO_DIR)/mesh.proto ]; then \
		echo "Error: meshtastic-protobufs not found. Please run: git submodule update --init --recursive"; \
		exit 1; \
	fi
	@command -v protoc >/dev/null 2>&1 || { echo "protoc not found. Please install Protocol Buffers compiler."; exit 1; }
	protoc --go_out=$(PB_DIR) --go_opt=paths=source_relative \
		--proto_path=../../../meshtastic-protobufs \
		$(PROTO_DIR)/*.proto
	@echo "Protobuf generation complete"

# Clean generated protobuf files
proto-clean:
	@echo "Cleaning protobuf files..."
	@rm -rf $(PB_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover ./...

# Format code
format:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Please install it."; exit 1; }
	golangci-lint run -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/bin/$(BINARY_NAME)
	@echo "Installed to $(HOME)/bin/$(BINARY_NAME)"

# Run the TUI
run-tui: build
	@echo "Running TUI..."
	$(BUILD_DIR)/$(BINARY_NAME) tui

# Run with debug logging
run-debug: build
	@echo "Running with debug logging..."
	$(BUILD_DIR)/$(BINARY_NAME) --log-level debug tui

# Show device info
info: build
	@echo "Getting device info..."
	$(BUILD_DIR)/$(BINARY_NAME) info

# Listen for messages
listen: build
	@echo "Listening for messages..."
	$(BUILD_DIR)/$(BINARY_NAME) listen

# Development workflow
dev: format lint test build

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  deps          - Install dependencies"
	@echo "  proto-gen     - Generate protobuf files"
	@echo "  proto-clean   - Clean generated protobuf files"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  format        - Format code"
	@echo "  lint          - Lint code"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to ~/bin"
	@echo "  run-tui       - Run the TUI"
	@echo "  run-debug     - Run with debug logging"
	@echo "  info          - Show device info"
	@echo "  listen        - Listen for messages"
	@echo "  dev           - Development workflow (format, lint, test, build)"
	@echo "  help          - Show this help"
