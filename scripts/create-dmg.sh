#!/bin/bash

set -e

APP_NAME="ClaudeCodeMonitor"
DMG_NAME="ClaudeCodeMonitor"
VERSION="1.0.0"

# Paths
DIST_DIR="dist"
APP_BUNDLE="${DIST_DIR}/${APP_NAME}.app"
DMG_DIR="dmg-build"
DMG_FILE="${DIST_DIR}/${DMG_NAME}-${VERSION}.dmg"

# Check if app bundle exists
if [ ! -d "$APP_BUNDLE" ]; then
    echo "Error: $APP_BUNDLE not found. Run 'make app' first."
    exit 1
fi

echo "Creating DMG installer..."

# Clean up previous build
rm -rf "$DMG_DIR"
rm -f "$DMG_FILE"

# Create temporary directory
mkdir -p "$DMG_DIR"

# Copy app bundle
echo "Copying app bundle..."
cp -R "$APP_BUNDLE" "$DMG_DIR/"

# Create Applications symlink for easy installation
echo "Creating Applications symlink..."
ln -s /Applications "$DMG_DIR/Applications"

# Create DMG
echo "Creating DMG file..."
hdiutil create -volname "$DMG_NAME" \
    -srcfolder "$DMG_DIR" \
    -ov \
    -format UDZO \
    "$DMG_FILE"

# Clean up
echo "Cleaning up..."
rm -rf "$DMG_DIR"

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
