# Define variables
PROJECT_NAME := tuskbot
GO_FLAGS := -trimpath -ldflags="-s -w"

.PHONY: all build test clean run format llamacpp release-linux release-macos _build_linux_amd64 _build_darwin_arm64

# Default target
all: deps build test

# For dev purposes
build:
	@echo "Building $(PROJECT_NAME)..."
	@go build $(GO_FLAGS) -o bin/tusk cmd/tusk/*.go
	@du -h bin/tusk
	@echo "$(PROJECT_NAME) built successfully."

run: build
	@echo "Running $(PROJECT_NAME)..."
	@./bin/tusk start

# Release targets
release-linux:
	@echo "🚀 Launching Docker build for Linux..."
	@DOCKER_BUILDKIT=1 docker build \
		--file build/release/Dockerfile \
		--target export \
		--output bin \
		--build-arg BUILD_TARGET=linux_amd64 \
		.
	@mv bin/tusk bin/tusk-linux-amd64

release-linux-arm64:
	@echo "🚀 Launching Docker build for Linux ARM64..."
	@DOCKER_BUILDKIT=1 docker build \
		--file build/release/Dockerfile \
		--target export \
		--output bin \
		--build-arg BUILD_TARGET=linux_arm64 \
		.
	@mv bin/tusk bin/tusk-linux-arm64

release-macos:
	@echo "🚀 Launching Docker build for macOS..."
	@DOCKER_BUILDKIT=1 docker build \
		--file build/release/Dockerfile \
		--target export \
		--output bin \
		--build-arg BUILD_TARGET=darwin_arm64 \
		.
	@mv bin/tusk bin/tusk-darwin-arm64

_build_linux_amd64:
	@echo "🐧 Internal: Compiling for Linux (Static Musl)..."
	@CGO_ENABLED=1 \
		GOOS=linux GOARCH=amd64 \
		CC="zig cc -target x86_64-linux-musl" \
		CXX="zig c++ -target x86_64-linux-musl" \
		go build $(GO_FLAGS) -o bin/tusk cmd/tusk/*.go

_build_linux_arm64:
	@echo "🐧 Internal: Compiling for Linux ARM64 (Static Musl)..."
	@CGO_ENABLED=1 \
		GOOS=linux GOARCH=arm64 \
		CC="zig cc -target aarch64-linux-musl" \
		CXX="zig c++ -target aarch64-linux-musl" \
		go build $(GO_FLAGS) -o bin/tusk cmd/tusk/*.go

_build_darwin_arm64:
	@echo "🍎 Internal: Compiling for macOS (Apple Silicon)..."
	@CGO_ENABLED=1 \
		GOOS=darwin GOARCH=arm64 \
		CC="zig cc -target aarch64-macos" \
		CXX="zig c++ -target aarch64-macos" \
		CGO_LDFLAGS=" \
			-L$(MACOS_SDK)/usr/lib \
			-F$(MACOS_SDK)/System/Library/Frameworks \
		" \
		go build $(GO_FLAGS) -o bin/tusk cmd/tusk/*.go

# Build production Docker image (requires linux binary first)
DOCKER_IMAGE_NAME ?= tuskbot
DOCKER_IMAGE_TAG ?= latest
DOCKERFILE := build/docker/Dockerfile

release-image:
	@echo "🐳 Building Docker image $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	@docker build \
		-f $(DOCKERFILE) \
		-t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
		.
	@echo "✅ Image built: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"
	@docker images $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) --format "Size: {{.Size}}"

# Testing rargets
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests completed successfully."

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./internal/...

heap:
	@go build -gcflags="-m" ./internal/... 2>&1  | grep escapes

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@rm -rf vendor/
	@go mod tidy
	@go mod vendor
	@echo "Dependencies installed successfully."

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -rf vendor/
	@echo "Cleanup completed successfully."

# Add migration
migration:
	@echo "Creating new migration..."
	@goose -dir ./internal/storage/sqlite/migrations create $(name) sql
	@echo "New migration created successfully."

# LLAMACPP BUILDER

LLAMA_DOCKER_IMG := llama-builder
LLAMA_DOCKERFILE := build/llamacpp/Dockerfile
LLAMA_CONTEXT    := build/llamacpp
LLAMA_DEST_DIR   := pkg/llamacpp
LLAMA_TEMP_CONT  := llama-extract-temp

llamacpp:
	@echo "--- 🐳 Building Llama.cpp artifacts via Docker ---"
	@# 1. Building Docker image
	docker build -t $(LLAMA_DOCKER_IMG) -f $(LLAMA_DOCKERFILE) $(LLAMA_CONTEXT)

	@echo "--- 📦 Extracting artifacts ---"
	@# 2. Removing temporary containers
	@docker rm -f $(LLAMA_TEMP_CONT) 2>/dev/null || true

	@# 3. Creating container from image for extraction
	docker create --name $(LLAMA_TEMP_CONT) $(LLAMA_DOCKER_IMG)

	@# 4. Clean old artifacts
	rm -rf $(LLAMA_DEST_DIR)/lib $(LLAMA_DEST_DIR)/include

	@# 5. Copying contents of /output directory from container to pkg/llamacpp
	@# Warning: the trailing slash "/." copies the CONTENT of the directory, not the directory itself
	docker cp $(LLAMA_TEMP_CONT):/output/. $(LLAMA_DEST_DIR)/

	@# 6. Removing temporary container
	docker rm $(LLAMA_TEMP_CONT)

	@echo "--- ✅ Done! Artifacts placed in $(LLAMA_DEST_DIR) ---"
	@ls -R $(LLAMA_DEST_DIR)/lib

# SQLITE VEC BUILDER

SQLITE_DOCKER_IMG := sqlite-builder
SQLITE_DOCKERFILE := build/sqlite/Dockerfile
SQLITE_CONTEXT    := build/sqlite
SQLITE_DEST_DIR   := pkg/sqlite
SQLITE_TEMP_CONT  := sqlite-extract-temp

sqlite:
	@echo "--- 🗄️  Building SQLite-Vec artifacts via Docker ---"
	@# 1. Building Docker image
	docker build -t $(SQLITE_DOCKER_IMG) -f $(SQLITE_DOCKERFILE) .

	@echo "--- 📦 Extracting artifacts ---"
	@# 2. Removing temporary containers
	@docker rm -f $(SQLITE_TEMP_CONT) 2>/dev/null || true

	@# 3. Creating container from image for extraction
	docker create --name $(SQLITE_TEMP_CONT) $(SQLITE_DOCKER_IMG)

	@# 4. Clean old artifacts
	rm -rf $(SQLITE_DEST_DIR)/lib $(SQLITE_DEST_DIR)/include

	@# 5. Copying contents of /output directory from container to pkg/sqlite
	@mkdir -p $(SQLITE_DEST_DIR)
	docker cp $(SQLITE_TEMP_CONT):/output/. $(SQLITE_DEST_DIR)/

	@# 6. Removing temporary container
	docker rm $(SQLITE_TEMP_CONT)

	@echo "--- ✅ Done! Artifacts placed in $(SQLITE_DEST_DIR) ---"
	@ls -R $(SQLITE_DEST_DIR)/lib
