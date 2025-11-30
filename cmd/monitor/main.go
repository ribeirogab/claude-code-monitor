package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/getlantern/systray"

	"github.com/ribeirogab/claude-code-monitor/internal/config"
	"github.com/ribeirogab/claude-code-monitor/internal/executor"
	"github.com/ribeirogab/claude-code-monitor/internal/scheduler"
	"github.com/ribeirogab/claude-code-monitor/internal/updater"
)

// AppVersion is set at build time via ldflags
var AppVersion = "dev"

const (
	GitHubOwner = "ribeirogab"
	GitHubRepo  = "claude-code-monitor"
)

type UsageData struct {
	SessionPercent    int    `json:"session_percent"`
	SessionReset      string `json:"session_reset"`
	WeekAllPercent    int    `json:"week_all_percent"`
	WeekAllReset      string `json:"week_all_reset"`
	WeekOpusPercent   int    `json:"week_opus_percent"`
	WeekOpusReset     string `json:"week_opus_reset"`
	WeekSonnetPercent int    `json:"week_sonnet_percent"`
	WeekSonnetReset   string `json:"week_sonnet_reset"`
	Timestamp         string `json:"timestamp"`
}

type MenuItemRefs struct {
	sessionPercent    *systray.MenuItem
	sessionReset      *systray.MenuItem
	weekAllPercent    *systray.MenuItem
	weekAllReset      *systray.MenuItem
	weekOpusPercent   *systray.MenuItem
	weekOpusReset     *systray.MenuItem
	weekSonnetPercent *systray.MenuItem
	weekSonnetReset   *systray.MenuItem
	lastUpdate        *systray.MenuItem
}

var (
	sched             *scheduler.Scheduler
	menuRefs          *MenuItemRefs
	usageDataPath     string
	appConfig         *config.Config
	mUpdateNow        *systray.MenuItem
	intervalMenuItems map[int]*systray.MenuItem
	mDisabled         *systray.MenuItem
	outputDir         string
	scriptPath        string
	taskWithUpdate    func() error
	appUpdater        *updater.Updater
	mUpdateAvailable  *systray.MenuItem
)

func main() {
	// Setup logging
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	logDir := filepath.Join(homeDir, ".claude-code-monitor")
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "monitor.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Starting Claude Code Monitor...")

	// Print to console as well
	fmt.Printf("Claude Code Monitor starting...\n")
	fmt.Printf("Logs: %s/monitor.log\n", logDir)

	systray.Run(onReady, onExit)
}

func onReady() {
	log.Println("onReady() called")

	// Load and set icon
	iconData, err := loadIcon()
	if err == nil {
		systray.SetIcon(iconData)
		log.Println("Icon loaded and set")
	} else {
		systray.SetTitle("claude-code")
		log.Printf("Icon not found, using title. Error: %v", err)
	}

	systray.SetTooltip("Claude Code Usage Monitor")

	// Setup paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	outputDir := filepath.Join(homeDir, ".claude-code-monitor")
	usageDataPath = filepath.Join(outputDir, "claude-code-usage.json")

	// Load configuration
	appConfig, err = config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		appConfig = config.DefaultConfig()
	}
	log.Printf("Config loaded: auto-update=%v", appConfig.AutoUpdateEnabled)

	// Create menu items with usage data
	createMenuItems()
	log.Println("Usage menu items created")

	// Add control menu items
	mUpdateNow = systray.AddMenuItem("Update Now", "")

	systray.AddSeparator()

	// Add Settings menu with submenu
	mSettings := systray.AddMenuItem("Settings", "")

	// Add Auto-Update submenu
	mAutoUpdateMenu := mSettings.AddSubMenuItem("Auto-Update", "")

	// Initialize interval menu items map
	intervalMenuItems = make(map[int]*systray.MenuItem)

	// Add "Disabled" option
	mDisabled = mAutoUpdateMenu.AddSubMenuItemCheckbox("Disabled", "", !appConfig.AutoUpdateEnabled)

	// Add interval options (in seconds)
	intervals := []struct {
		seconds int
		label   string
	}{
		{60, "1 minute"},
		{300, "5 minutes"},
		{600, "10 minutes"},
		{1800, "30 minutes"},
		{3600, "60 minutes"},
	}

	for _, interval := range intervals {
		isChecked := appConfig.AutoUpdateEnabled && interval.seconds == appConfig.UpdateInterval
		item := mAutoUpdateMenu.AddSubMenuItemCheckbox(interval.label, "", isChecked)
		intervalMenuItems[interval.seconds] = item
	}

	systray.AddSeparator()

	// Add update available menu item (hidden by default)
	mUpdateAvailable = systray.AddMenuItem("", "")
	mUpdateAvailable.Hide()

	// Initialize updater
	appUpdater = updater.New(updater.Config{
		Owner:          GitHubOwner,
		Repo:           GitHubRepo,
		CurrentVersion: AppVersion,
		CheckInterval:  1 * time.Hour,
	})

	// Set callback for when update is found
	appUpdater.OnUpdateFound = func(info *updater.UpdateInfo) {
		if info.HasUpdate {
			mUpdateAvailable.SetTitle(fmt.Sprintf("Update Available (v%s)", info.LatestVersion.String()))
			mUpdateAvailable.Show()
			log.Printf("Update available: %s -> %s", info.CurrentVersion.String(), info.LatestVersion.String())
		}
	}

	// Start periodic update checks
	appUpdater.StartPeriodicCheck()
	log.Println("Updater started")

	// Handle update available click
	go func() {
		for range mUpdateAvailable.ClickedCh {
			log.Println("Opening releases page...")
			if err := appUpdater.OpenLatestRelease(); err != nil {
				log.Printf("Failed to open releases page: %v", err)
			}
		}
	}()

	systray.AddSeparator()

	// Add Quit menu item
	mQuit := systray.AddMenuItem("Quit", "")
	log.Println("Quit menu item added")

	scriptPath = findScriptPath()
	log.Printf("Script path: %s", scriptPath)
	log.Printf("Output directory: %s", outputDir)

	exec := executor.New(scriptPath, outputDir)

	// Wrapper to update menu after execution
	taskWithUpdate = func() error {
		err := exec.Execute()
		if err == nil {
			updateMenuItems()
			log.Println("Menu items updated")
		}
		return err
	}

	// Create scheduler with configured interval
	interval := time.Duration(appConfig.UpdateInterval) * time.Second
	sched = scheduler.New(interval, taskWithUpdate)
	log.Println("Scheduler created")

	// Start scheduler in background
	go sched.Start()

	// Pause scheduler if auto-update is disabled
	if !appConfig.AutoUpdateEnabled {
		sched.Pause()
		log.Println("Scheduler started (paused)")
	} else {
		log.Println("Scheduler started")
	}

	// Handle Update Now button
	go func() {
		for range mUpdateNow.ClickedCh {
			log.Println("Manual update triggered")
			mUpdateNow.SetTitle("Updating...")
			mUpdateNow.Disable()

			if err := taskWithUpdate(); err != nil {
				log.Printf("Manual update failed: %v", err)
			}

			mUpdateNow.Enable()
			mUpdateNow.SetTitle("Update Now")
		}
	}()

	// Handle "Disabled" option
	go func() {
		for range mDisabled.ClickedCh {
			log.Println("Auto-update disabled")

			// Check Disabled, uncheck all intervals
			mDisabled.Check()
			for _, mi := range intervalMenuItems {
				mi.Uncheck()
			}

			// Update config
			appConfig.AutoUpdateEnabled = false
			if err := config.SaveConfig(appConfig); err != nil {
				log.Printf("Failed to save config: %v", err)
			}

			// Pause scheduler
			sched.Pause()
		}
	}()

	// Handle interval changes
	for intervalSeconds, menuItem := range intervalMenuItems {
		seconds := intervalSeconds // capture for closure
		item := menuItem
		go func() {
			for range item.ClickedCh {
				log.Printf("Changing interval to %d seconds and enabling auto-update", seconds)

				// Uncheck Disabled
				mDisabled.Uncheck()

				// Update checkmarks for intervals
				for s, mi := range intervalMenuItems {
					if s == seconds {
						mi.Check()
					} else {
						mi.Uncheck()
					}
				}

				// Update config
				appConfig.AutoUpdateEnabled = true
				appConfig.UpdateInterval = seconds
				if err := config.SaveConfig(appConfig); err != nil {
					log.Printf("Failed to save config: %v", err)
				}

				// Restart scheduler with new interval
				sched.Stop()

				interval := time.Duration(seconds) * time.Second
				sched = scheduler.New(interval, taskWithUpdate)
				go sched.Start()

				log.Printf("Scheduler restarted with %d second interval (auto-update enabled)", seconds)
			}
		}()
	}

	// Wait for quit signal
	go func() {
		<-mQuit.ClickedCh
		log.Println("Quit clicked")
		systray.Quit()
	}()

	log.Println("Application ready and running")
}

func onExit() {
	log.Println("onExit() called")
	if sched != nil {
		sched.Stop()
	}
	log.Println("Application exited")
}

func findScriptPath() string {
	// Try to find the script in common locations
	locations := []string{
		"./claude-code-usage.sh",
		"../claude-code-usage.sh",
		"../../claude-code-usage.sh",
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		locations = append([]string{
			filepath.Join(execDir, "claude-code-usage.sh"),
			filepath.Join(execDir, "..", "claude-code-usage.sh"),
		}, locations...)
	}

	for _, loc := range locations {
		absPath, err := filepath.Abs(loc)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	log.Fatal("Could not find claude-code-usage.sh script")
	return ""
}

func loadIconByName(iconName string) ([]byte, error) {
	// Try multiple paths for icon (dev mode and app bundle)
	paths := []string{
		fmt.Sprintf("assets/icons/%s.png", iconName),
		fmt.Sprintf("../Resources/assets/icons/%s.png", iconName),
	}

	// Add path relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(exeDir, "..", "Resources", "assets", "icons", fmt.Sprintf("%s.png", iconName)),
			filepath.Join(exeDir, "assets", "icons", fmt.Sprintf("%s.png", iconName)),
		)
	}

	// Try each path
	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			log.Printf("Icon loaded from: %s", path)
			return data, nil
		}
	}

	return nil, fmt.Errorf("icon %s not found in any expected location", iconName)
}

func loadIcon() ([]byte, error) {
	return loadIconByName("menubar-icon")
}

func updateIcon(sessionPercent int) {
	var iconName string

	if sessionPercent > 85 {
		iconName = "menubar-icon-red"
	} else if sessionPercent > 50 {
		iconName = "menubar-icon-yellow"
	} else {
		iconName = "menubar-icon"
	}

	iconData, err := loadIconByName(iconName)
	if err != nil {
		log.Printf("Failed to load icon %s: %v", iconName, err)
		return
	}

	systray.SetIcon(iconData)
	log.Printf("Icon updated to: %s (session: %d%%)", iconName, sessionPercent)
}

func loadUsageData() (*UsageData, error) {
	data, err := os.ReadFile(usageDataPath)
	if err != nil {
		return nil, err
	}

	var usage UsageData
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, err
	}

	return &usage, nil
}

func hasOpusAccess(usage *UsageData) bool {
	return usage.WeekOpusReset != ""
}

func hasSonnetAccess(usage *UsageData) bool {
	return usage.WeekSonnetReset != ""
}

func formatTimestamp(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "Unknown"
	}

	// Convert to local timezone
	local := t.Local()
	return local.Format("Jan 2 at 3:04pm")
}

func removeTimezone(text string) string {
	// Remove timezone like "(America/Sao_Paulo)" from text
	re := regexp.MustCompile(`\s*\([^)]+\)`)
	return re.ReplaceAllString(text, "")
}

func getUsageEmoji(percent int) string {
	if percent <= 50 {
		return "🟢"
	} else if percent <= 85 {
		return "🟡"
	}
	return "🔴"
}

func updateMenuItems() {
	usage, err := loadUsageData()
	if err != nil {
		log.Printf("Failed to load usage data: %v", err)
		return
	}

	if menuRefs == nil {
		return
	}

	// Update icon based on session usage
	updateIcon(usage.SessionPercent)

	// Update Session
	if menuRefs.sessionPercent != nil {
		menuRefs.sessionPercent.SetTitle(fmt.Sprintf("Session              %02d%%   %s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent)))
	}
	if menuRefs.sessionReset != nil {
		menuRefs.sessionReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset)))
	}

	// Update Week (All)
	if menuRefs.weekAllPercent != nil {
		menuRefs.weekAllPercent.SetTitle(fmt.Sprintf("Week (All)          %02d%%   %s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent)))
	}
	if menuRefs.weekAllReset != nil {
		menuRefs.weekAllReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset)))
	}

	// Update Week (Opus) - only if menu items exist
	if menuRefs.weekOpusPercent != nil {
		menuRefs.weekOpusPercent.SetTitle(fmt.Sprintf("Week (Opus)       %02d%%   %s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent)))
	}
	if menuRefs.weekOpusReset != nil {
		menuRefs.weekOpusReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset)))
	}

	// Update Week (Sonnet) - only if menu items exist
	if menuRefs.weekSonnetPercent != nil {
		menuRefs.weekSonnetPercent.SetTitle(fmt.Sprintf("Week (Sonnet)   %02d%%   %s", usage.WeekSonnetPercent, getUsageEmoji(usage.WeekSonnetPercent)))
	}
	if menuRefs.weekSonnetReset != nil {
		menuRefs.weekSonnetReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekSonnetReset)))
	}

	// Update Last update
	if menuRefs.lastUpdate != nil {
		menuRefs.lastUpdate.SetTitle(formatTimestamp(usage.Timestamp))
	}
}

func createMenuItems() {
	usage, err := loadUsageData()

	menuRefs = &MenuItemRefs{}

	var sessionText, sessionResetText string
	var weekAllText, weekAllResetText string
	var weekOpusText, weekOpusResetText string
	var weekSonnetText, weekSonnetResetText string
	var lastUpdateText string
	var showOpus, showSonnet bool

	if err != nil {
		sessionText = "Session         Loading..."
		sessionResetText = "resets: N/A"
		weekAllText = "Week (All)      Loading..."
		weekAllResetText = "resets: N/A"
		weekOpusText = "Week (Opus)     Loading..."
		weekOpusResetText = "resets: N/A"
		weekSonnetText = "Week (Sonnet)   Loading..."
		weekSonnetResetText = "resets: N/A"
		lastUpdateText = "N/A"
		showOpus = false
		showSonnet = true // Default to showing Sonnet for new users
	} else {
		sessionText = fmt.Sprintf("Session              %02d%%   %s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent))
		sessionResetText = fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset))
		weekAllText = fmt.Sprintf("Week (All)          %02d%%   %s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent))
		weekAllResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset))
		weekOpusText = fmt.Sprintf("Week (Opus)       %02d%%   %s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent))
		weekOpusResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset))
		weekSonnetText = fmt.Sprintf("Week (Sonnet)   %02d%%   %s", usage.WeekSonnetPercent, getUsageEmoji(usage.WeekSonnetPercent))
		weekSonnetResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekSonnetReset))
		lastUpdateText = formatTimestamp(usage.Timestamp)
		showOpus = hasOpusAccess(usage)
		showSonnet = hasSonnetAccess(usage)
	}

	// Session
	menuRefs.sessionPercent = systray.AddMenuItem(sessionText, "")
	menuRefs.sessionReset = systray.AddMenuItem(sessionResetText, "")
	menuRefs.sessionReset.Disable()
	systray.AddSeparator()

	// Week (All)
	menuRefs.weekAllPercent = systray.AddMenuItem(weekAllText, "")
	menuRefs.weekAllReset = systray.AddMenuItem(weekAllResetText, "")
	menuRefs.weekAllReset.Disable()
	systray.AddSeparator()

	// Week (Opus) - only create if user has Opus access (old format)
	if showOpus {
		menuRefs.weekOpusPercent = systray.AddMenuItem(weekOpusText, "")
		menuRefs.weekOpusReset = systray.AddMenuItem(weekOpusResetText, "")
		menuRefs.weekOpusReset.Disable()
		systray.AddSeparator()
	}

	// Week (Sonnet) - only create if user has Sonnet access (new format)
	if showSonnet {
		menuRefs.weekSonnetPercent = systray.AddMenuItem(weekSonnetText, "")
		menuRefs.weekSonnetReset = systray.AddMenuItem(weekSonnetResetText, "")
		menuRefs.weekSonnetReset.Disable()
		systray.AddSeparator()
	}

	menuRefs.lastUpdate = systray.AddMenuItem(lastUpdateText, "")
	menuRefs.lastUpdate.Disable()
}
