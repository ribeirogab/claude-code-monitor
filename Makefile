.PHONY: help run build build-intel build-arm build-universal clean install

APP_NAME := claude-code-monitor
BUILD_DIR := build
CMD_DIR := cmd/monitor

help:
	@echo "Available targets:"
	@echo "  make run              - Run the application in development mode"
	@echo "  make build            - Build for current architecture"
	@echo "  make build-intel      - Build for Intel (amd64)"
	@echo "  make build-arm        - Build for Apple Silicon (arm64)"
	@echo "  make build-universal  - Build universal binary (Intel + Apple Silicon)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make install          - Install to /usr/local/bin"
	@echo "  make help             - Show this help message"

run:
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_DIR)/main.go

build:
	@echo "Building $(APP_NAME) for current architecture..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

build-intel:
	@echo "Building $(APP_NAME) for Intel (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-amd64 $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-amd64"

build-arm:
	@echo "Building $(APP_NAME) for Apple Silicon (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-arm64 $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-arm64"

build-universal: build-intel build-arm
	@echo "Creating universal binary..."
	@lipo -create -output $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME)-amd64 $(BUILD_DIR)/$(APP_NAME)-arm64
	@rm $(BUILD_DIR)/$(APP_NAME)-amd64 $(BUILD_DIR)/$(APP_NAME)-arm64
	@echo "Universal binary created: $(BUILD_DIR)/$(APP_NAME)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

install: build-universal
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "Installation complete. Run '$(APP_NAME)' to start the application."
