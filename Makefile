.PHONY: generate generate.proto generate.proto.protoc lint.proto clean.proto build run clean help

# Default target
.DEFAULT_GOAL := help

PODMAN ?= podman
APP_IMAGE = photos-ng:latest
REGISTRY = quay.io/ctupangiu/photos-ng
REMOTE_TAG ?= latest

POSTGRES_IMAGE ?= docker.io/library/postgres:17
GIT_COMMIT=$(shell git rev-list -1 HEAD --abbrev-commit)
VERSION=$(shell cat VERSION)

# Project variables
BINARY_NAME=photos-ng
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./main.go
TMP_DATA_FOLDER=/tmp/photos-ng

# Generate code (OpenAPI, protobuf, etc.)
generate:
	@echo "Generating code..."
	go generate ./...
	@echo "Code generation complete."

# Generate protobuf code using buf in container
generate.proto:
	@echo "Generating protobuf code with buf in container..."
	$(PODMAN) run --rm \
		-v $(CURDIR)/api/v1/grpc:/workspace \
		-w /workspace \
		bufbuild/buf:latest \
		generate
	@echo "Protobuf generation complete."

# Generate protobuf code using protoc in container
generate.proto.protoc:
	@echo "Generating protobuf code with protoc in container..."
	$(PODMAN) run --rm \
		-v $(CURDIR)/api/v1/grpc:/workspace \
		-w /workspace \
		namely/protoc-all:1.51_1 \
		-f *.proto -l go -o .
	@echo "Protobuf generation complete."

# Lint protobuf files in container
lint.proto:
	@echo "Linting protobuf files in container..."
	$(PODMAN) run --rm \
		-v $(CURDIR)/api/v1/grpc:/workspace \
		-w /workspace \
		bufbuild/buf:latest \
		lint
	@echo "Protobuf linting complete."

# Clean protobuf generated files
clean.proto:
	@echo "Cleaning protobuf generated files..."
	rm -f api/v1/grpc/*.pb.go
	@echo "Protobuf clean complete."

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-X main.sha=${GIT_COMMIT}" -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Run the application
run:
	@echo "Create temp data root folder..."
	@mkdir -p $(TMP_DATA_FOLDER)
	@echo "Using temp directory: $$TMP_DIR"
	@echo "Running $(BINARY_NAME)..."
	$(BINARY_PATH) serve --data-root-folder=$(TMP_DATA_FOLDER)

run.mnt:
	@echo "Running $(BINARY_NAME)..."
	$(BINARY_PATH) serve --data-root-folder=/mnt

run.ui:
	cd ./ui && npm run start:dev

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	rm -rf $(TMP_DATA_FOLDER)
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


#####################
# Container targets #
#####################

# Build the application image
podman.build: ## Build the Finante application container
	podman build -f Containerfile --build-arg GIT_SHA=$(GIT_COMMIT) -t $(APP_IMAGE) .

# Tag image for remote registry
podman.tag: ## Tag the local image for remote registry
	podman tag $(APP_IMAGE) $(REGISTRY):$(REMOTE_TAG)

# Push image to remote registry
podman.push: podman.tag ## Push the container image to quay.io/ctupangiu/finance
	podman push $(REGISTRY):$(REMOTE_TAG)

# Build and push in one command
deploy.image: podman.build podman.push ## Build and push the container image to remote registry
# Display help information
help:
	@echo "Available targets:"
	@echo "  generate          - Generate code from specs and run go generate"
	@echo "  generate.proto    - Generate protobuf code using buf in container"
	@echo "  generate.proto.protoc - Generate protobuf code using protoc in container"
	@echo "  lint.proto        - Lint protobuf files in container"
	@echo "  clean.proto       - Clean generated protobuf files"
	@echo "  build             - Build the application binary"
	@echo "  run               - Run the application directly with go run"
	@echo "  clean             - Clean build artifacts"
	@echo "  db.start          - Start the database"
	@echo "  db.stop           - Stop the database"
	@echo "  db.migrate        - Migrate the database" 
	@echo "  help              - Show this help message"
