#!/bin/bash

# Add common paths where claude CLI might be installed
export PATH="/usr/local/bin:/opt/homebrew/bin:$HOME/.npm-global/bin:$HOME/.npm/bin:/usr/bin:/bin:$PATH"

# Execution log file
EXEC_LOG="claude-code-usage-execution.log"

# Clear previous execution log
> "$EXEC_LOG"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$EXEC_LOG"
}

log "Starting Claude Code usage capture..."

# Use expect to automate the interaction
(
expect << 'EXPECT_END'
spawn claude /usage

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
) > claude-code-usage.log 2>&1

log "Checking result..."

# Check if "Current session" appears in the log
if grep -q "Current session" claude-code-usage.log; then
    log "SUCCESS! Usage information captured."

    # Parse the data from the log
    log "Parsing usage data..."

    # Extract percentages and reset times
    SESSION_PERCENT=$(grep -A 1 "Current session" claude-code-usage.log | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    SESSION_RESET=$(grep -A 2 "Current session" claude-code-usage.log | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    WEEK_ALL_PERCENT=$(grep -A 1 "Current week (all models)" claude-code-usage.log | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    WEEK_ALL_RESET=$(grep -A 2 "Current week (all models)" claude-code-usage.log | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    WEEK_OPUS_PERCENT=$(grep -A 1 "Current week (Opus)" claude-code-usage.log | grep "used" | grep -o '[0-9]*% used' | grep -o '[0-9]*')
    WEEK_OPUS_RESET=$(grep -A 2 "Current week (Opus)" claude-code-usage.log | grep "Resets" | sed 's/.*Resets\s*\(.*\)/\1/' | sed 's/\[38;[0-9;]*m//g' | sed 's/\[39m//g' | sed 's/\[22m//g' | sed 's/\[1m//g' | sed 's/\x1b//g' | tr -d '\r' | xargs)

    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Create JSON file
    log "Creating JSON output..."
    cat > claude-code-usage.json <<EOF
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

    log "JSON file created: claude-code-usage.json"

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
    echo "Files created:"
    echo "  - claude-code-usage.log (raw output)"
    echo "  - claude-code-usage.json (parsed data)"
    echo "  - claude-code-usage-execution.log (execution log)"
    echo ""

    exit 0
else
    log "FAILED. 'Current session' not found in log."
    exit 1
fi