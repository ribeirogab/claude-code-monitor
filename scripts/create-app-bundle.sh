#!/bin/bash

BUNDLE_NAME="ClaudeCodeMonitor.app"
BINARY_PATH="build/claude-code-monitor"
SCRIPT_PATH="claude-code-usage.sh"
INFO_PLIST_PATH="assets/Info.plist"

# Create app bundle structure
mkdir -p "${BUNDLE_NAME}/Contents/MacOS"
mkdir -p "${BUNDLE_NAME}/Contents/Resources"

# Copy binary
cp "${BINARY_PATH}" "${BUNDLE_NAME}/Contents/MacOS/claude-code-monitor"
chmod +x "${BUNDLE_NAME}/Contents/MacOS/claude-code-monitor"

# Copy script
cp "${SCRIPT_PATH}" "${BUNDLE_NAME}/Contents/MacOS/${SCRIPT_PATH}"
chmod +x "${BUNDLE_NAME}/Contents/MacOS/${SCRIPT_PATH}"

# Copy Info.plist
cp "${INFO_PLIST_PATH}" "${BUNDLE_NAME}/Contents/Info.plist"

echo "App bundle created: ${BUNDLE_NAME}"
echo "To run: open ${BUNDLE_NAME}"
