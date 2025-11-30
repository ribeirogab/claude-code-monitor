# AGENTS.md

This file provides guidance to Agents when working with code in this repository.

## Project Overview

Claude Code Monitor is a macOS menu bar application written in Go that monitors and displays Claude Code CLI usage statistics. It uses the `systray` library for menu bar integration and executes a bash script to capture usage data from the Claude CLI.

## Build Commands

```bash
make build            # Build for current architecture (version from git tag)
make build-universal  # Build universal binary (Intel + Apple Silicon)
make app-universal    # Create .app bundle with universal binary
make dmg-universal    # Create DMG installer
make dev              # Clean build and run
make run              # Run with go run (no build)
make clean            # Remove build artifacts
```

The app version is automatically injected from git tags via ldflags (`-X main.AppVersion=$(VERSION)`).

## Architecture

### Core Components

- **cmd/monitor/main.go**: Main entry point. Handles systray initialization, menu creation, and coordinates all components.

- **internal/scheduler/**: Manages periodic execution of the usage fetching script at configurable intervals.

- **internal/executor/**: Runs `claude-code-usage.sh` and handles script execution.

- **internal/config/**: Manages user settings (auto-update interval, enabled state). Persists to `~/.claude-code-monitor/config.json`.

- **internal/updater/**: Checks GitHub releases for new versions. Contains version parsing (`version.go`), GitHub API client (`github.go`), and update logic (`updater.go`).

### Data Flow

1. Scheduler triggers executor at configured interval
2. Executor runs `claude-code-usage.sh` which:
   - Uses `expect` to automate Claude CLI interaction
   - Parses `/usage` command output
   - Writes JSON to `~/.claude-code-monitor/claude-code-usage.json`
3. Main app reads JSON and updates menu items
4. Menu bar icon color changes based on session usage percentage

### Key Files

- **claude-code-usage.sh**: Bash script that captures Claude CLI usage. Supports both "Week (Sonnet)" and "Week (Opus)" formats for backward compatibility.

- **assets/icons/**: Menu bar icons (menubar-icon.png, menubar-icon-yellow.png, menubar-icon-red.png) that change based on usage level.

## Data Storage

All runtime data stored in `~/.claude-code-monitor/`:
- `config.json` - User preferences
- `claude-code-usage.json` - Parsed usage data
- `monitor.log` - Application logs
