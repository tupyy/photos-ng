.PHONY: generate build run clean help

# Default target
.DEFAULT_GOAL := help

PODMAN ?= podman
POSTGRES_IMAGE ?= docker.io/library/postgres:17

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

# Db tragets
DB_HOST=localhost
DB_PORT=5432
ROOT_USER=postgres
ROOT_PWD=postgres
CONNSTR="postgresql://$(ROOT_USER):$(ROOT_PWD)@$(DB_HOST):$(DB_PORT)"

db.start:
	$(PODMAN) run --rm -p $(DB_PORT):5432 --name pg-photos -e POSTGRES_PASSWORD=$(ROOT_PWD) -d $(POSTGRES_IMAGE)

db.stop:
	$(PODMAN) stop pg-photos

db.migrate:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(CONNSTR) GOOSE_MIGRATION_DIR=$(CURDIR)/internal/datastore/pg/migrations/sql goose up

# Display help information
help:
	@echo "Available targets:"
	@echo "  generate  	- Generate code from specs and run go generate"
	@echo "  build     	- Build the application binary"
	@echo "  run       	- Run the application directly with go run"
	@echo "  clean     	- Clean build artifacts"
	@echo "  db.start  	- Start the database"
	@echo "  db.stop   	- Stop the database"
	@echo "  db.migrate	- Migrate the database" 
	@echo "  help      - Show this help message"
