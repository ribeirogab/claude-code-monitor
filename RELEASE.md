# v1.1.1

## What's New

### New Claude CLI Format Support

- Added support for parsing `Week (Sonnet only)` from the updated Claude Code CLI
- Backward compatible with older `Week (Opus)` format
- New JSON fields: `week_sonnet_percent` and `week_sonnet_reset`

### Dynamic Menu Bar Icons

The menu bar icon now changes color based on your session usage:

| Usage Level | Icon | Description |
|-------------|------|-------------|
| 0-50% | ![Green](assets/icons/menubar-icon.png) | Safe usage level |
| 51-85% | ![Yellow](assets/icons/menubar-icon-yellow.png) | Moderate usage |
| 86-100% | ![Red](assets/icons/menubar-icon-red.png) | High usage, approaching limit |

### Update Checker

- Automatically checks for new releases on GitHub (every hour)
- Shows "Update Available (vX.X.X)" in the menu when a new version is available
- Click to open the GitHub releases page directly

### Improved Versioning

- App version is now automatically injected from git tags at build time
- No more hardcoded version numbers in the binary

## Bug Fixes

- Fixed duplicate lines in usage parsing caused by terminal redraws
- Fixed menu item alignment using fixed-width spacing

## Files Changed

- `claude-code-usage.sh` - Updated parsing logic for new CLI format
- `cmd/monitor/main.go` - Dynamic icons, update checker integration
- `internal/updater/` - New update checker module (github.go, updater.go, version.go)
- `assets/icons/` - Added menubar-icon-red.png and menubar-icon-yellow.png

## Download

Download the DMG installer below and follow the [installation instructions](https://github.com/ribeirogab/claude-code-monitor#installation).

**Note:** This app is not signed with an Apple Developer certificate. You will need to allow it in System Settings > Privacy & Security after the first launch attempt.
