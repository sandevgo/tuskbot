# Define variables
PROJECT_NAME := tuskbot

# Linker flags: -s (disable symbol table) -w (disable DWARF generation)
LDFLAGS := -ldflags="-s -w"
GCFLAGS := -gcflags="" # -m
GO_BUILD_ENV := CGO_ENABLED=1

.PHONY: all build test clean run format llamacpp

# Default target
all: deps build test

# Build the project
build:
	@echo "Building $(PROJECT_NAME)..."
	@go build -trimpath $(GCFLAGS) $(LDFLAGS) -o bin/tusk cmd/tusk/*.go
	@du -h bin/tusk
	@echo "$(PROJECT_NAME) built successfully."

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests completed successfully."

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./internal/...

# Run the application
run: build
	@echo "Running $(PROJECT_NAME)..."
	@./bin/tusk start

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
	@rm -rf var/data
	@rm -rf var/raw
	@echo "Cleanup completed successfully."

heap:
	@go build -gcflags="-m" ./internal/... 2>&1  | grep escapes

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
