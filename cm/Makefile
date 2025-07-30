# Claude Manager (cm) Makefile

# Build variables
BINARY_NAME=claude-manager
BUILD_DIR=build
GO_FILES=$(shell find . -name '*.go')

# Alternative binary names
ALIASES=cm claude-web cmgr

# Go build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: all build clean install uninstall test fmt vet deps help

# Default target
all: build

# Build the binary
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES) go.mod go.sum
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Install to user's local bin
install: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"
	@echo ""
	@echo "Creating convenient aliases..."
	@for alias in $(ALIASES); do \
		ln -sf $(BINARY_NAME) ~/.local/bin/$$alias && echo "  ✅ $$alias -> $(BINARY_NAME)"; \
	done
	@echo ""
	@echo "Add ~/.local/bin to your PATH if not already there:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""
	@echo ""
	@echo "Available commands:"
	@echo "  claude-manager  (main binary)"
	@echo "  cm              (short alias)"
	@echo "  claude-web      (descriptive alias)"
	@echo "  cmgr            (manager alias)"

# Install to system-wide location (requires sudo)
install-system: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Uninstall from user's local bin
uninstall:
	@echo "Removing $(BINARY_NAME) from ~/.local/bin..."
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "Uninstalled from ~/.local/bin"

# Uninstall from system-wide location
uninstall-system:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled from /usr/local/bin"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Development build (with debug info)
build-dev: 
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) (development)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run the application
run: build-dev
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Cross-compilation targets
build-linux: 
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

build-macos:
	@mkdir -p $(BUILD_DIR)
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

build-windows:
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Build for all platforms
build-all: build-linux build-macos build-windows
	@echo "Built for all platforms"

# Check for dependencies
check-deps:
	@echo "Checking dependencies..."
	@go mod verify
	@go mod tidy -diff

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Show help
help:
	@echo "Claude Manager (cm) Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  build-dev      Build with debug info"
	@echo "  build-all      Build for all platforms"
	@echo "  install        Install to ~/.local/bin"
	@echo "  install-system Install to /usr/local/bin (requires sudo)"
	@echo "  uninstall      Remove from ~/.local/bin"
	@echo "  uninstall-system Remove from /usr/local/bin (requires sudo)"
	@echo "  run            Build and run the application"
	@echo "  test           Run tests"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  deps           Install dependencies"
	@echo "  update-deps    Update dependencies"
	@echo "  clean          Clean build artifacts"
	@echo "  help           Show this help"