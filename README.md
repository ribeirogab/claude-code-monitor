# Claude Code Monitor

A macOS menu bar application that automatically monitors Claude Code CLI usage and saves statistics to your home directory.

## Features

- Runs as a macOS menu bar application
- Automatically executes usage monitoring every minute
- Saves usage data to `~/.claude-code-monitor/`
- Supports both Intel and Apple Silicon Macs
- Simple and lightweight
- Graceful shutdown

## Requirements

- macOS (Intel or Apple Silicon)
- [Claude Code CLI](https://code.claude.com/) installed and configured
- `expect` command-line tool (pre-installed on macOS)
- Go 1.16+ (for building from source)

## Installation

### Using Make (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/ribeirogab/claude-code-monitor.git
cd claude-code-monitor
```

2. Build and install:
```bash
make install
```

This will create a universal binary that works on both Intel and Apple Silicon Macs and install it to `/usr/local/bin/`.

3. Run the application:
```bash
claude-code-monitor
```

### Manual Build

Build for your current architecture:
```bash
make build
./build/claude-code-monitor
```

Build for specific architectures:
```bash
# Intel only
make build-intel

# Apple Silicon only
make build-arm

# Universal binary (both architectures)
make build-universal
```

## Usage

1. Start the application by running `claude-code-monitor` or by double-clicking the binary
2. A small icon will appear in your macOS menu bar
3. The application will automatically run the usage monitoring script every minute
4. Usage data is saved to `~/.claude-code-monitor/`:
   - `claude-code-usage.json` - Parsed usage statistics
   - `claude-code-usage.log` - Raw output from the monitoring script
   - `claude-code-usage-execution.log` - Execution timestamps and logs
5. Click the menu bar icon and select "Quit" to stop the application

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

1. The application runs as a menu bar app using `systray`
2. On startup, it creates a scheduler that runs every minute
3. The scheduler executes `claude-code-usage.sh` which:
   - Launches Claude Code CLI
   - Captures the `/usage` command output
   - Parses usage percentages and reset times
   - Generates JSON output
4. Generated files are moved to `~/.claude-code-monitor/`
5. The process repeats every minute until you quit the application

## Troubleshooting

**Application doesn't start:**
- Ensure `claude-code-usage.sh` is in the same directory as the binary
- Check that Claude Code CLI is installed and accessible

**No data being generated:**
- Verify that Claude Code CLI is properly configured
- Check logs in `~/.claude-code-monitor/claude-code-usage-execution.log`
- Ensure `expect` is installed (run `which expect`)

**Menu bar icon not showing:**
- This is a known limitation of some macOS versions
- The app is still running - check Activity Monitor for `claude-code-monitor`

## License

MIT License - see [LICENSE](LICENSE) file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Built with [systray](https://github.com/getlantern/systray) for menu bar integration
- Monitors [Claude Code CLI](https://code.claude.com/) usage
