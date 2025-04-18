.PHONY: build clean install test run release

BINARY_NAME=chat-cli
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=build
MODULE_PATH=github.com/valdezdata/chat-cli
LDFLAGS=-ldflags="-X $(MODULE_PATH)/internal/version.Version=$(VERSION) -X $(MODULE_PATH)/internal/version.Commit=$(COMMIT) -X $(MODULE_PATH)/internal/version.BuildDate=$(BUILD_DATE)"

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Install the application to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME) $(VERSION)..."
	go install $(LDFLAGS) .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(LDFLAGS) .

# Create a new release
release:
	@echo "Creating release $(VERSION)..."
	@read -p "Enter new version (e.g., v1.0.0): " new_version; \
	git tag -a $$new_version -m "Release $$new_version"; \
	echo "Tagged $$new_version"; \
	echo "Run 'git push origin $$new_version' to push the tag to GitHub"
