#!/bin/bash

set -e

APP_NAME="ClaudeCodeMonitor"
DMG_NAME="ClaudeCodeMonitor"
VERSION="1.0.1"

# Paths
DIST_DIR="dist"
APP_BUNDLE="${DIST_DIR}/${APP_NAME}.app"
DMG_FILE="${DIST_DIR}/${DMG_NAME}-${VERSION}.dmg"

# Check if app bundle exists
if [ ! -d "$APP_BUNDLE" ]; then
    echo "Error: $APP_BUNDLE not found. Run 'make app' first."
    exit 1
fi

echo "Creating DMG installer..."

# Check if create-dmg is installed, install if missing
if ! command -v create-dmg &> /dev/null; then
    echo "create-dmg not found, attempting to install..."

    if command -v brew &> /dev/null; then
        echo "Installing create-dmg with Homebrew..."
        if brew install create-dmg > /dev/null 2>&1; then
            echo "create-dmg installed successfully"
        else
            echo "ERROR: Failed to install create-dmg with Homebrew"
            exit 1
        fi
    else
        echo "ERROR: create-dmg is required but not installed, and Homebrew is not available."
        echo "Install Homebrew first or install create-dmg manually: brew install create-dmg"
        exit 1
    fi
fi

# Clean up previous DMG
rm -f "$DMG_FILE"

# Create DMG with create-dmg (much prettier than hdiutil)
echo "Creating DMG file with custom styling..."
create-dmg \
    --volname "$DMG_NAME" \
    --volicon "assets/icons/app-icon.icns" \
    --window-pos 200 120 \
    --window-size 600 400 \
    --icon-size 100 \
    --icon "$APP_NAME.app" 150 150 \
    --hide-extension "$APP_NAME.app" \
    --app-drop-link 450 150 \
    --no-internet-enable \
    "$DMG_FILE" \
    "$APP_BUNDLE" \
    > /dev/null 2>&1

echo ""
echo "========================================="
echo "DMG created successfully!"
echo "========================================="
echo "File: $DMG_FILE"
echo "Size: $(du -h "$DMG_FILE" | cut -f1)"
echo ""
echo "To install:"
echo "1. Double-click $DMG_FILE"
echo "2. Drag $APP_NAME to Applications folder"
echo ""
