# Makefile to build a Go binary for Linux on AMD64 architecture

# Name of binary
BINARY_NAME=go-identity-server

# Source files
SOURCE_FILES=main.go server.go token-service.go user-service.go vm-service.go

# Build directory
BUILD_DIR=./bin

build:
	@echo "Building binary"
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_FILES)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"


run:
	@echo "Running binary"
	go run $(SOURCE_FILES)

clean: 
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed."

.PHONY: build run clean