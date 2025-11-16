#!/bin/bash

DIST_DIR="dist"
BUNDLE_NAME="ClaudeCodeMonitor.app"
BUNDLE_PATH="${DIST_DIR}/${BUNDLE_NAME}"
BINARY_PATH="build/claude-code-monitor"
SCRIPT_PATH="claude-code-usage.sh"
INFO_PLIST_PATH="assets/Info.plist"
MENUBAR_ICON_PATH="assets/icons/menubar-icon.png"
APP_ICON_PATH="assets/icons/app-icon.icns"

# Create app bundle structure
mkdir -p "${BUNDLE_PATH}/Contents/MacOS"
mkdir -p "${BUNDLE_PATH}/Contents/Resources/assets/icons"

# Copy binary
cp "${BINARY_PATH}" "${BUNDLE_PATH}/Contents/MacOS/claude-code-monitor"
chmod +x "${BUNDLE_PATH}/Contents/MacOS/claude-code-monitor"

# Copy script
cp "${SCRIPT_PATH}" "${BUNDLE_PATH}/Contents/MacOS/${SCRIPT_PATH}"
chmod +x "${BUNDLE_PATH}/Contents/MacOS/${SCRIPT_PATH}"

# Copy Info.plist
cp "${INFO_PLIST_PATH}" "${BUNDLE_PATH}/Contents/Info.plist"

# Copy menubar icon
cp "${MENUBAR_ICON_PATH}" "${BUNDLE_PATH}/Contents/Resources/assets/icons/menubar-icon.png"

# Copy app icon
cp "${APP_ICON_PATH}" "${BUNDLE_PATH}/Contents/Resources/app-icon.icns"

echo "App bundle created: ${BUNDLE_PATH}"
echo "To run: open ${BUNDLE_PATH}"
