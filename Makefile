.PHONY: generate build run clean help

# Default target
.DEFAULT_GOAL := help

# Project variables
BINARY_NAME=photos-ng
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./main.go

# Generate code (OpenAPI, protobuf, etc.)
generate:
	@echo "Generating code..."
	go generate ./...
	@echo "Code generation complete."

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(BINARY_PATH) serve

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "Clean complete."

# Display help information
help:
	@echo "Available targets:"
	@echo "  generate  - Generate code from specs and run go generate"
	@echo "  build     - Build the application binary"
	@echo "  run       - Run the application directly with go run"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help message"
