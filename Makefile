.PHONY: generate generate.proto generate.proto.protoc lint.proto clean.proto build run run.withauth run.ui clean help spicedb.start spicedb.stop spicedb.schema keycloak.start keycloak.start.postgres keycloak.start.server keycloak.stop envoy.start envoy.stop run.auth stop.auth

# Default target
.DEFAULT_GOAL := help

PODMAN ?= podman
APP_IMAGE = photos-ng:latest
REGISTRY = rhel2.tls.tupangiu.ro:5000/photos-ng
LATEST_TAG ?= latest

POSTGRES_IMAGE ?= docker.io/library/postgres:17
GIT_COMMIT=$(shell git rev-list -1 HEAD --abbrev-commit)
VERSION=$(shell cat VERSION)

# Project variables
BINARY_NAME=photos-ng
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./main.go
TMP_DATA_FOLDER=/tmp/photos-ng

# Keycloak targets
KEYCLOAK_DB_PORT=5433
KEYCLOAK_REALM=photos
OIDC_WELLKNOWN_URL=http://localhost:8000/realms/$(KEYCLOAK_REALM)/.well-known/openid-configuration
OIDC_CLIENT_ID=photos
OIDC_CLIENT_SECRET=bAz9ReuZ92gdOEsHG9H9aLZSjynPmo3o

# SpiceDB targets
SPICEDB_IMAGE ?= authzed/spicedb:latest
SPICEDB_GRPC_PORT=50051
SPICEDB_PRESHARED_KEY=dev-secret-key

# Db tragets
DB_HOST=localhost
DB_PORT=5432
ROOT_USER=postgres
ROOT_PWD=postgres
CONNSTR="postgresql://$(ROOT_USER):$(ROOT_PWD)@$(DB_HOST):$(DB_PORT)"

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

run.withauth:
	@echo "Create temp data root folder..."
	@mkdir -p $(TMP_DATA_FOLDER)
	@echo "Running $(BINARY_NAME) with authentication..."
	$(BINARY_PATH) serve \
		--log-level=info \
		--data-root-folder=$(TMP_DATA_FOLDER) \
		--authentication-enabled \
		--authentication-wellknown-endpoint=$(OIDC_WELLKNOWN_URL) \
		--authentication-client-id=$(OIDC_CLIENT_ID) \
		--authentication-client-secret=$(OIDC_CLIENT_SECRET) \
		--authorization-enabled \
		--authorization-spicedb-url=localhost:$(SPICEDB_GRPC_PORT) \
		--authorization-spicedb-preshared-key=$(SPICEDB_PRESHARED_KEY)

run.ui:
	cd ./ui && npm run start:dev

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	rm -rf $(TMP_DATA_FOLDER)
	@echo "Clean complete."

db.start:
	$(PODMAN) run --rm -p $(DB_PORT):5432 --name pg-photos -e POSTGRES_PASSWORD=$(ROOT_PWD) -d $(POSTGRES_IMAGE)

db.stop:
	$(PODMAN) stop pg-photos

db.migrate:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(CONNSTR) GOOSE_MIGRATION_DIR=$(CURDIR)/pkg/migrations/sql goose up

spicedb.start:
	$(PODMAN) run --rm -d \
		--name spicedb-dev \
		-p $(SPICEDB_GRPC_PORT):50051 \
		$(SPICEDB_IMAGE) serve \
		--grpc-preshared-key "$(SPICEDB_PRESHARED_KEY)"

spicedb.stop:
	$(PODMAN) stop spicedb-dev

spicedb.schema:
	zed schema write $(CURDIR)/resources/schema.zed \
		--endpoint=localhost:$(SPICEDB_GRPC_PORT) \
		--token="$(SPICEDB_PRESHARED_KEY)" \
		--insecure

keycloak.start: keycloak.start.postgres keycloak.start.server

keycloak.start.postgres:
	$(PODMAN) play kube $(CURDIR)/resources/keycloak-postgres.yml
	@echo "Waiting for PostgreSQL to be ready..."
	@until pg_isready -h localhost -p $(KEYCLOAK_DB_PORT) -U keycloak > /dev/null 2>&1; do sleep 1; done
	@echo "PostgreSQL is ready"

keycloak.start.server:
	$(PODMAN) play kube $(CURDIR)/resources/keycloak.yml

keycloak.stop:
	-$(PODMAN) play kube --down $(CURDIR)/resources/keycloak.yml
	-$(PODMAN) play kube --down $(CURDIR)/resources/keycloak-postgres.yml

# Envoy targets
ENVOY_IMAGE ?= docker.io/envoyproxy/envoy:v1.32-latest

envoy.start:
	$(PODMAN) run -d --rm\
		--name envoy-oauth2 \
		--network host \
		-v $(CURDIR)/resources/envoy.yaml:/etc/envoy/envoy.yaml \
		$(ENVOY_IMAGE)

envoy.stop:
	$(PODMAN) stop envoy-oauth2

# Auth stack targets (SpiceDB + Keycloak + Envoy)
run.auth: spicedb.start spicedb.schema keycloak.start envoy.start
	@echo "Auth stack started:"
	@echo "  - SpiceDB:  localhost:$(SPICEDB_GRPC_PORT)"
	@echo "  - Keycloak: http://localhost:8000"
	@echo "  - Envoy:    http://localhost:7070"

stop.auth:
	-$(MAKE) envoy.stop
	-$(MAKE) keycloak.stop
	-$(MAKE) spicedb.stop


#####################
# Container targets #
#####################

# Build the application image
podman.build: ## Build the Finante application container
	podman build -f Containerfile --build-arg GIT_SHA=$(GIT_COMMIT) -t $(APP_IMAGE) .

# Push image to remote registry
podman.push: podman.build
	podman tag $(APP_IMAGE) $(REGISTRY):$(GIT_COMMIT)
	podman push $(REGISTRY):$(GIT_COMMIT)
	podman tag $(APP_IMAGE) $(REGISTRY):$(LATEST_TAG)
	podman push $(REGISTRY):$(LATEST_TAG)

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
	@echo "  run               - Run the application directly"
	@echo "  run.withauth      - Run the application with OIDC authentication"
	@echo "  run.ui            - Run the UI development server"
	@echo "  clean             - Clean build artifacts"
	@echo "  db.start          - Start the database"
	@echo "  db.stop           - Stop the database"
	@echo "  db.migrate        - Migrate the database"
	@echo "  spicedb.start     - Start SpiceDB dev container"
	@echo "  spicedb.stop      - Stop SpiceDB dev container"
	@echo "  spicedb.schema    - Import schema.zed into SpiceDB"
	@echo "  keycloak.start    - Start Keycloak with PostgreSQL"
	@echo "  keycloak.stop     - Stop Keycloak pod"
	@echo "  envoy.start       - Start Envoy OAuth2 proxy"
	@echo "  envoy.stop        - Stop Envoy proxy"
	@echo "  run.auth          - Start full auth stack (SpiceDB + Keycloak + Envoy)"
	@echo "  stop.auth         - Stop full auth stack"
	@echo "  help              - Show this help message"
