# Go parameters
BINARY_NAME=shadow-pulse
MAIN_PATH=./cmd/shadow-pulse

# Versioning
VERSION := $(shell git describe --tags --always --dirty)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags="-X 'main.version=$(VERSION)' -X 'main.buildDate=$(BUILD_DATE)'"

# Default target - runs when 'make' is called without arguments
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME) (v$(VERSION))..."
	@go build -buildvcs=false $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(BINARY_NAME) built successfully."

# Clean the binary
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@echo "Cleanup complete."

# Test the application
test:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests complete."

.PHONY: all build clean test
