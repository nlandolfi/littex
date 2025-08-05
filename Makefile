# Makefile for LitTex project

# Go commands
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := $(GO) test
GOMOD := $(GO) mod
GOBUILD := $(GO) build

# Binary name and paths
BINARY_NAME := lit
CMD_DIR := ./cmd/lit
BINARY_PATH := $(CMD_DIR)/$(BINARY_NAME)
COVERAGE_FILE := coverage.out

# Build flags
BUILD_FLAGS := -v
LDFLAGS := -ldflags="-s -w"

# Test flags
TEST_FLAGS := -race
TEST_VERBOSE_FLAGS := -v $(TEST_FLAGS)
COVERAGE_FLAGS := -coverprofile=$(COVERAGE_FILE) -covermode=atomic

# Lint flags
LINT_FLAGS := run --config .golangci.yml

# Define all phony targets (targets that don't create files)
.PHONY: all build test test-verbose lint fmt clean install mod-tidy coverage help

# Default target
all: help

# Build the lit binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run all tests with race detection
test:
	@echo "Running tests with race detection..."
	@$(GOTEST) $(TEST_FLAGS) ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@$(GOTEST) $(TEST_VERBOSE_FLAGS) ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@$(GOLINT) $(LINT_FLAGS)

# Format all Go files
fmt:
	@echo "Formatting Go files..."
	@$(GOFMT) -s -w .
	@echo "Formatting complete"

# Clean built binaries
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE)
	@echo "Clean complete"

# Install the lit binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(LDFLAGS) $(CMD_DIR)
	@echo "Installation complete"

# Run go mod tidy
mod-tidy:
	@echo "Tidying Go modules..."
	@$(GOMOD) tidy
	@echo "Tidy complete"

# Generate test coverage report
coverage:
	@echo "Generating test coverage report..."
	@$(GOTEST) $(COVERAGE_FLAGS) ./...
	@$(GO) tool cover -html=$(COVERAGE_FILE)

# Show available targets
help:
	@echo "LitTex Makefile targets:"
	@echo "  build         - Build the lit binary"
	@echo "  test          - Run all tests with race detection"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format all Go files"
	@echo "  clean         - Remove built binaries and coverage files"
	@echo "  install       - Install the lit binary to GOPATH/bin"
	@echo "  mod-tidy      - Run go mod tidy"
	@echo "  coverage      - Generate test coverage report"
	@echo "  help          - Show this help message"
