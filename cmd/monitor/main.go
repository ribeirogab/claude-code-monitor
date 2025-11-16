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
)

type UsageData struct {
	SessionPercent  int    `json:"session_percent"`
	SessionReset    string `json:"session_reset"`
	WeekAllPercent  int    `json:"week_all_percent"`
	WeekAllReset    string `json:"week_all_reset"`
	WeekOpusPercent int    `json:"week_opus_percent"`
	WeekOpusReset   string `json:"week_opus_reset"`
	Timestamp       string `json:"timestamp"`
}

type MenuItemRefs struct {
	sessionPercent  *systray.MenuItem
	sessionReset    *systray.MenuItem
	weekAllPercent  *systray.MenuItem
	weekAllReset    *systray.MenuItem
	weekOpusPercent *systray.MenuItem
	weekOpusReset   *systray.MenuItem
	lastUpdate      *systray.MenuItem
}

var (
	sched         *scheduler.Scheduler
	menuRefs      *MenuItemRefs
	usageDataPath string
	appConfig     *config.Config
	mUpdateNow    *systray.MenuItem
	mAutoUpdate   *systray.MenuItem
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
	mUpdateNow = systray.AddMenuItem("Update Now", "Manually update usage data")

	autoUpdateText := "Disable Auto-Update"
	if !appConfig.AutoUpdateEnabled {
		autoUpdateText = "Enable Auto-Update"
	}
	mAutoUpdate = systray.AddMenuItem(autoUpdateText, "Toggle automatic updates")

	systray.AddSeparator()

	// Add Quit menu item
	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	log.Println("Quit menu item added")

	scriptPath := findScriptPath()
	log.Printf("Script path: %s", scriptPath)
	log.Printf("Output directory: %s", outputDir)

	exec := executor.New(scriptPath, outputDir)

	// Wrapper to update menu after execution
	taskWithUpdate := func() error {
		err := exec.Execute()
		if err == nil {
			updateMenuItems()
			log.Println("Menu items updated")
		}
		return err
	}

	// Create scheduler with 1-minute interval
	sched = scheduler.New(time.Minute, taskWithUpdate)
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
			if err := taskWithUpdate(); err != nil {
				log.Printf("Manual update failed: %v", err)
			}
		}
	}()

	// Handle Auto-update toggle
	go func() {
		for range mAutoUpdate.ClickedCh {
			appConfig.AutoUpdateEnabled = !appConfig.AutoUpdateEnabled

			if appConfig.AutoUpdateEnabled {
				mAutoUpdate.SetTitle("Disable Auto-Update")
				sched.Resume()
				log.Println("Auto-update enabled")
			} else {
				mAutoUpdate.SetTitle("Enable Auto-Update")
				sched.Pause()
				log.Println("Auto-update disabled")
			}

			// Save config
			if err := config.SaveConfig(appConfig); err != nil {
				log.Printf("Failed to save config: %v", err)
			}
		}
	}()

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

func loadIcon() ([]byte, error) {
	// Try multiple paths for icon (dev mode and app bundle)
	paths := []string{
		"assets/icons/menubar-icon.png",
		"../Resources/assets/icons/menubar-icon.png",
	}

	// Add path relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(exeDir, "..", "Resources", "assets", "icons", "menubar-icon.png"),
			filepath.Join(exeDir, "assets", "icons", "menubar-icon.png"),
		)
	}

	// Try each path
	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			log.Printf("Icon loaded from: %s", path)
			return data, nil
		}
	}

	return nil, fmt.Errorf("icon file not found in any expected location")
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

	// Update Session
	if menuRefs.sessionPercent != nil {
		menuRefs.sessionPercent.SetTitle(fmt.Sprintf("Session    \t%02d%% \t%s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent)))
	}
	if menuRefs.sessionReset != nil {
		menuRefs.sessionReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset)))
	}

	// Update Week (All)
	if menuRefs.weekAllPercent != nil {
		menuRefs.weekAllPercent.SetTitle(fmt.Sprintf("Week (All)\t%02d%% \t%s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent)))
	}
	if menuRefs.weekAllReset != nil {
		menuRefs.weekAllReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset)))
	}

	// Update Week (Opus) - only if menu items exist
	if menuRefs.weekOpusPercent != nil {
		menuRefs.weekOpusPercent.SetTitle(fmt.Sprintf("Week (Opus)\t%02d%% \t%s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent)))
	}
	if menuRefs.weekOpusReset != nil {
		menuRefs.weekOpusReset.SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset)))
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
	var lastUpdateText string
	var showOpus bool

	if err != nil {
		sessionText = "Session    \tLoading..."
		sessionResetText = "resets: N/A"
		weekAllText = "Week (All)\tLoading..."
		weekAllResetText = "resets: N/A"
		weekOpusText = "Week (Opus)\tLoading..."
		weekOpusResetText = "resets: N/A"
		lastUpdateText = "N/A"
		showOpus = true
	} else {
		sessionText = fmt.Sprintf("Session    \t%02d%% \t%s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent))
		sessionResetText = fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset))
		weekAllText = fmt.Sprintf("Week (All)\t%02d%% \t%s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent))
		weekAllResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset))
		weekOpusText = fmt.Sprintf("Week (Opus)\t%02d%% \t%s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent))
		weekOpusResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset))
		lastUpdateText = formatTimestamp(usage.Timestamp)
		showOpus = hasOpusAccess(usage)
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

	// Week (Opus) - only create if user has Opus access
	if showOpus {
		menuRefs.weekOpusPercent = systray.AddMenuItem(weekOpusText, "")
		menuRefs.weekOpusReset = systray.AddMenuItem(weekOpusResetText, "")
		menuRefs.weekOpusReset.Disable()
		systray.AddSeparator()
	}

	menuRefs.lastUpdate = systray.AddMenuItem(lastUpdateText, "")
	menuRefs.lastUpdate.Disable()

	systray.AddSeparator()
}
