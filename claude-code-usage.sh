#!/bin/bash

# Output directory
OUTPUT_DIR="$HOME/.claude-code-monitor"
mkdir -p "$OUTPUT_DIR"

# Execution log file
EXEC_LOG="$OUTPUT_DIR/claude-code-usage-execution.log"

# Clear previous execution log
> "$EXEC_LOG"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$EXEC_LOG"
}

log "Starting Claude Code usage capture..."

# Check if jq is installed, install if missing
if ! command -v jq &> /dev/null; then
    log "jq not found, attempting to install..."

    # Check if brew is available
    if command -v brew &> /dev/null; then
        log "Installing jq with Homebrew..."
        if brew install jq >> "$EXEC_LOG" 2>&1; then
            log "jq installed successfully"
        else
            log "ERROR: Failed to install jq with Homebrew"
            exit 1
        fi
    else
        log "ERROR: jq is required but not installed, and Homebrew is not available."
        log "Install Homebrew first: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        log "Or install jq manually: brew install jq"
        exit 1
    fi
fi

# Check if expect is installed, install if missing
if ! command -v expect &> /dev/null; then
    log "expect not found, attempting to install..."

    # Check if brew is available
    if command -v brew &> /dev/null; then
        log "Installing expect with Homebrew..."
        if brew install expect >> "$EXEC_LOG" 2>&1; then
            log "expect installed successfully"
        else
            log "ERROR: Failed to install expect with Homebrew"
            exit 1
        fi
    else
        log "ERROR: expect is required but not installed, and Homebrew is not available."
        log "Install Homebrew first: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        log "Or install expect manually: brew install expect"
        exit 1
    fi
fi

# Check if .claude.json exists
CLAUDE_CONFIG="$HOME/.claude.json"
if [ ! -f "$CLAUDE_CONFIG" ]; then
    log "ERROR: ~/.claude.json not found."
    log "Run 'claude' at least once to initialize configuration."
    exit 1
fi

# Get current script directory
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
log "Script running from: $SCRIPT_DIR"

# Backup .claude.json
cp "$CLAUDE_CONFIG" "$CLAUDE_CONFIG.bak"
log "Backed up .claude.json to .claude.json.bak"

# Add current directory to trusted projects
log "Adding $SCRIPT_DIR to trusted directories..."
jq --arg dir "$SCRIPT_DIR" '
  .projects[$dir] = {
    "allowedTools": [],
    "mcpContextUris": [],
    "mcpServers": {},
    "enabledMcpjsonServers": [],
    "disabledMcpjsonServers": [],
    "hasTrustDialogAccepted": true,
    "projectOnboardingSeenCount": 0,
    "hasClaudeMdExternalIncludesApproved": false,
    "hasClaudeMdExternalIncludesWarningShown": false,
    "exampleFiles": []
  }' "$CLAUDE_CONFIG" > "$CLAUDE_CONFIG.tmp"

# Validate the temporary file is valid JSON
if ! jq empty "$CLAUDE_CONFIG.tmp" 2>/dev/null; then
    log "ERROR: Failed to update .claude.json - invalid JSON generated"
    rm "$CLAUDE_CONFIG.tmp"
    exit 1
fi

# Move temp file to original
mv "$CLAUDE_CONFIG.tmp" "$CLAUDE_CONFIG"
log "Successfully added $SCRIPT_DIR to trusted directories"

# Find claude CLI in common locations
CLAUDE_CMD=""

# Try common paths
SEARCH_PATHS=(
    "/usr/local/bin/claude"
    "/opt/homebrew/bin/claude"
    "$HOME/.local/bin/claude"
    "$HOME/.npm-global/bin/claude"
    "$HOME/.npm/bin/claude"
    "$HOME/.nvm/versions/node/v"*/bin/claude
)

for path in "${SEARCH_PATHS[@]}"; do
    # Expand glob patterns
    for expanded_path in $path; do
        if [ -x "$expanded_path" ]; then
            CLAUDE_CMD="$expanded_path"
            log "Found claude at: $CLAUDE_CMD"
            break 2
        fi
    done
done

# If not found in common paths, try using 'which' with updated PATH
if [ -z "$CLAUDE_CMD" ]; then
    export PATH="/usr/local/bin:/opt/homebrew/bin:$HOME/.npm-global/bin:$HOME/.npm/bin:$HOME/.nvm/versions/node/$(ls -1 $HOME/.nvm/versions/node 2>/dev/null | tail -1)/bin:/usr/bin:/bin:$PATH"
    CLAUDE_CMD=$(which claude 2>/dev/null)
    if [ -n "$CLAUDE_CMD" ]; then
        log "Found claude using which: $CLAUDE_CMD"
    fi
fi

# If still not found, exit with error
if [ -z "$CLAUDE_CMD" ]; then
    log "ERROR: claude CLI not found. Please install it first."
    exit 1
fi

# Export PATH to include the directory where claude (and node) is located
CLAUDE_DIR=$(dirname "$CLAUDE_CMD")
export PATH="$CLAUDE_DIR:$PATH"
log "Updated PATH to include: $CLAUDE_DIR"

# Use expect to automate the interaction
(
expect << EXPECT_END
spawn $CLAUDE_CMD /usage

# Wait for usage screen to load completely
sleep 5

# Send ESC to exit usage screen
send "\033"

# Wait for ESC to be processed
sleep 1

# Exit Claude
send "exit\r"
expect eof
EXPECT_END
) > "$OUTPUT_DIR/claude-code-usage.log" 2>&1

log "Checking result..."

# Check if "Current session" appears in the log
if grep -q "Current session" "$OUTPUT_DIR/claude-code-usage.log"; then
    log "SUCCESS! Usage information captured."

    # Parse the data from the log
    log "Parsing usage data..."

    # Extract percentages and reset times
    SESSION_PERCENT=$(grep -A 1 "Current session" "$OUTPUT_DIR/claude-code-usage.log" | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    SESSION_RESET=$(grep -A 2 "Current session" "$OUTPUT_DIR/claude-code-usage.log" | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    WEEK_ALL_PERCENT=$(grep -A 1 "Current week (all models)" "$OUTPUT_DIR/claude-code-usage.log" | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    WEEK_ALL_RESET=$(grep -A 2 "Current week (all models)" "$OUTPUT_DIR/claude-code-usage.log" | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    WEEK_OPUS_PERCENT=$(grep -A 1 "Current week (Opus)" "$OUTPUT_DIR/claude-code-usage.log" | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    WEEK_OPUS_RESET=$(grep -A 2 "Current week (Opus)" "$OUTPUT_DIR/claude-code-usage.log" | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Create JSON file
    log "Creating JSON output..."
    cat > "$OUTPUT_DIR/claude-code-usage.json" <<EOF
{
  "session_percent": ${SESSION_PERCENT:-0},
  "session_reset": "${SESSION_RESET}",
  "week_all_percent": ${WEEK_ALL_PERCENT:-0},
  "week_all_reset": "${WEEK_ALL_RESET}",
  "week_opus_percent": ${WEEK_OPUS_PERCENT:-0},
  "week_opus_reset": "${WEEK_OPUS_RESET}",
  "timestamp": "${TIMESTAMP}"
}
EOF

    log "JSON file created: $OUTPUT_DIR/claude-code-usage.json"

    # Display summary
    echo ""
    echo "========================================="
    echo "Usage Summary:"
    echo "========================================="
    echo "Session:        ${SESSION_PERCENT}% (resets ${SESSION_RESET})"
    echo "Week (All):     ${WEEK_ALL_PERCENT}% (resets ${WEEK_ALL_RESET})"
    echo "Week (Opus):    ${WEEK_OPUS_PERCENT}% (resets ${WEEK_OPUS_RESET})"
    echo "========================================="
    echo ""
    echo "Files created in $OUTPUT_DIR:"
    echo "  - claude-code-usage.log (raw output)"
    echo "  - claude-code-usage.json (parsed data)"
    echo "  - claude-code-usage-execution.log (execution log)"
    echo ""

    exit 0
else
    log "FAILED. 'Current session' not found in log."
    exit 1
fi