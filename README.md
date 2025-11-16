# Claude Code Monitor

A macOS menu bar application that monitors and displays Claude Code CLI usage statistics directly in your menubar.

## Features

- Displays usage statistics in menubar dropdown with visual indicators
- Shows Session, Week (All), and Week (Opus) usage percentages
- Color-coded emoji indicators (🟢 0-50%, 🟡 51-85%, 🔴 86-100%)
- Displays reset times for each metric
- Auto-updates every minute
- Menubar-only app (does not appear in Dock)
- Auto-installs dependencies (jq)
- Auto-configures directory trust
- Auto-detects Claude CLI location (including NVM installations)
- Saves detailed logs to `~/.claude-code-monitor/`
- Supports both Intel and Apple Silicon Macs
- Lightweight and runs in background

## Requirements

- macOS (Intel or Apple Silicon)
- [Claude Code CLI](https://code.claude.com/) installed and configured
- `expect` command-line tool (pre-installed on macOS)
- Homebrew (for auto-installing jq if needed)
- Go 1.16+ (for building from source)

Note: The app will automatically install `jq` via Homebrew if not found, and will auto-detect Claude CLI location even in NVM installations.

## Installation

### For End Users (macOS App)

1. Download `ClaudeCodeMonitor.app` from releases
2. Move it to `/Applications` folder
3. Double-click to open
4. The app icon will appear in your menubar

### For Developers

#### Generate .app for Distribution

```bash
git clone https://github.com/ribeirogab/claude-code-monitor.git
cd claude-code-monitor
make app
```

This creates `ClaudeCodeMonitor.app` ready for distribution.

#### Install from Source

```bash
make install
```

This installs the binary to `/usr/local/bin/`.

#### Build for Specific Architectures

```bash
# Current architecture
make build

# Intel only
make build-intel

# Apple Silicon only
make build-arm
```

Note: Cross-compilation requires CGO toolchains. Use `make app` to build for current architecture.

## Usage

1. Start the application by opening `ClaudeCodeMonitor.app`
2. A small icon will appear in your macOS menu bar
3. Click the icon to see current usage statistics:
   - Session usage with percentage and emoji indicator
   - Week (All models) usage
   - Week (Opus) usage
   - Reset times for each metric
   - Last update timestamp
4. The display auto-updates every minute
5. Usage data is also saved to `~/.claude-code-monitor/`:
   - `claude-code-usage.json` - Parsed usage statistics
   - `claude-code-usage.log` - Raw output from monitoring script
   - `claude-code-usage-execution.log` - Execution timestamps and logs
   - `monitor.log` - Application logs
6. Click the menu bar icon and select "Quit" to stop the application

### Visual Indicators

The app uses color-coded emojis to indicate usage levels:

- 🟢 Green (0-50%): Safe usage level
- 🟡 Yellow (51-85%): Moderate usage
- 🔴 Red (86-100%): High usage, approaching limit

## Output Format

The `claude-code-usage.json` file contains:

```json
{
  "session_percent": 10,
  "session_reset": "in 2 hours",
  "week_all_percent": 25,
  "week_all_reset": "in 5 days",
  "week_opus_percent": 15,
  "week_opus_reset": "in 5 days",
  "timestamp": "2025-11-15T10:30:00Z"
}
```

## Development

Run in development mode:

```bash
make run
```

Clean build artifacts:

```bash
make clean
```

Show all available commands:

```bash
make help
```

## Project Structure

```
.
├── cmd/
│   └── monitor/          # Main application entry point
│       └── main.go
├── internal/
│   ├── executor/         # Script execution logic
│   │   └── executor.go
│   └── scheduler/        # Periodic task scheduling
│       └── scheduler.go
├── assets/               # Application assets (icons, etc.)
├── claude-code-usage.sh  # Monitoring script
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── README.md             # This file
```

## How It Works

1. The application runs as a menubar-only app using `systray`
2. On startup, it:
   - Loads menubar icon from assets
   - Creates menu items for displaying usage stats
   - Starts a scheduler that runs every minute
3. The scheduler executes `claude-code-usage.sh` which:
   - Auto-installs `jq` via Homebrew if not found
   - Pre-configures directory trust in `~/.claude.json` to bypass security prompts
   - Auto-detects Claude CLI location (supports standard paths and NVM installations)
   - Launches Claude Code CLI using `expect`
   - Captures the `/usage` command output
   - Parses usage percentages and reset times
   - Generates JSON output with timestamp
4. After successful execution, the menubar display updates automatically
5. The process repeats every minute until you quit the application

## Troubleshooting

**Application doesn't start:**

- Check that Claude Code CLI is installed and accessible: `which claude`
- Verify Homebrew is installed (for auto-installing jq): `which brew`
- Check application logs in `~/.claude-code-monitor/monitor.log`

**No data being generated:**

- Verify that Claude Code CLI is properly configured
- Check logs in `~/.claude-code-monitor/claude-code-usage-execution.log`
- Ensure `expect` is installed (pre-installed on macOS): `which expect`
- If using NVM, ensure Node.js is properly installed

**Menu not updating:**

- Check if JSON file is being updated: `cat ~/.claude-code-monitor/claude-code-usage.json`
- View execution logs: `tail -f ~/.claude-code-monitor/claude-code-usage-execution.log`
- Restart the application

**Menu bar icon not showing:**

- Check if icon file exists in the app bundle
- The app will show "claude-code" as title text if icon is not found
- The app is still running - check Activity Monitor for `claude-code-monitor`

**Trust dialog appearing:**

- This should be auto-handled by the script
- If it persists, manually trust the directory in Claude Code CLI
- Check if `~/.claude.json` has the correct permissions

## License

MIT License - see [LICENSE](LICENSE) file for details
