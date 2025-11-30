.PHONY: help run dev build build-intel build-arm build-universal app app-intel app-universal dmg dmg-universal clean install

APP_NAME := claude-code-monitor
BUILD_DIR := build
DIST_DIR := dist
CMD_DIR := cmd/monitor
BUNDLE_NAME := ClaudeCodeMonitor.app
VERSION := $(shell /usr/libexec/PlistBuddy -c "Print :CFBundleShortVersionString" assets/Info.plist 2>/dev/null || echo "dev")
LDFLAGS := -X main.AppVersion=$(VERSION)

help:
	@echo "Available targets:"
	@echo "  make run              - Run the application in development mode"
	@echo "  make dev              - Build and execute the binary"
	@echo "  make build            - Build for current architecture"
	@echo "  make build-intel      - Build for Intel (amd64)"
	@echo "  make build-arm        - Build for Apple Silicon (arm64)"
	@echo "  make build-universal  - Build universal binary (Intel + Apple Silicon)"
	@echo "  make app              - Create macOS app bundle for current architecture"
	@echo "  make app-intel        - Create macOS app bundle for Intel only"
	@echo "  make app-universal    - Create macOS app bundle with universal binary"
	@echo "  make dmg              - Create DMG installer from existing app bundle"
	@echo "  make dmg-universal    - Build universal app and create DMG installer"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make install          - Install to /usr/local/bin"
	@echo "  make help             - Show this help message"

run:
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_DIR)/main.go

dev: clean build
	@echo "Killing existing instances..."
	@killall $(APP_NAME) 2>/dev/null || true
	@echo "Starting $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

build:
	@echo "Building $(APP_NAME) $(VERSION) for current architecture..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

build-intel:
	@echo "Building $(APP_NAME) $(VERSION) for Intel (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-amd64 $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-amd64"

build-arm:
	@echo "Building $(APP_NAME) $(VERSION) for Apple Silicon (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-arm64 $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-arm64"

build-universal: build-intel build-arm
	@echo "Creating universal binary..."
	@lipo -create -output $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME)-amd64 $(BUILD_DIR)/$(APP_NAME)-arm64
	@rm $(BUILD_DIR)/$(APP_NAME)-amd64 $(BUILD_DIR)/$(APP_NAME)-arm64
	@echo "Universal binary created: $(BUILD_DIR)/$(APP_NAME)"

app: build
	@echo "Creating macOS app bundle..."
	@mkdir -p $(DIST_DIR)
	@./scripts/create-app-bundle.sh
	@echo "App bundle created: $(DIST_DIR)/$(BUNDLE_NAME)"
	@echo "To run: open $(DIST_DIR)/$(BUNDLE_NAME)"

app-intel: build-intel
	@echo "Creating macOS app bundle for Intel..."
	@mkdir -p $(DIST_DIR)
	@cp $(BUILD_DIR)/$(APP_NAME)-amd64 $(BUILD_DIR)/$(APP_NAME)
	@./scripts/create-app-bundle.sh
	@rm $(BUILD_DIR)/$(APP_NAME)
	@echo "App bundle created: $(DIST_DIR)/$(BUNDLE_NAME) (Intel only)"
	@echo "To run: open $(DIST_DIR)/$(BUNDLE_NAME)"

app-universal: build-universal
	@echo "Creating macOS app bundle with universal binary..."
	@mkdir -p $(DIST_DIR)
	@./scripts/create-app-bundle.sh
	@echo "App bundle created: $(DIST_DIR)/$(BUNDLE_NAME) (Universal - Intel + Apple Silicon)"
	@echo "To run: open $(DIST_DIR)/$(BUNDLE_NAME)"

dmg:
	@./scripts/create-dmg.sh

dmg-universal: app-universal
	@./scripts/create-dmg.sh

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf dmg-build
	@echo "Clean complete"

install: build-universal
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "Installation complete. Run '$(APP_NAME)' to start the application."
