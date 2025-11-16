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

var (
	sched         *scheduler.Scheduler
	menuItems     []*systray.MenuItem
	usageDataPath string
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

	// Create menu items with usage data
	createMenuItems()
	log.Println("Usage menu items created")

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
	log.Println("Scheduler started")

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

	if len(menuItems) >= 7 {
		// Update Session
		menuItems[0].SetTitle(fmt.Sprintf("Session    \t%02d%% \t%s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent)))
		menuItems[1].SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset)))

		// Update Week (All)
		menuItems[2].SetTitle(fmt.Sprintf("Week (All)\t%02d%% \t%s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent)))
		menuItems[3].SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset)))

		// Update Week (Opus)
		menuItems[4].SetTitle(fmt.Sprintf("Week (Opus)\t%02d%% \t%s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent)))
		menuItems[5].SetTitle(fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset)))

		// Update Last update
		menuItems[6].SetTitle(formatTimestamp(usage.Timestamp))
	}
}

func createMenuItems() {
	usage, err := loadUsageData()

	var sessionText, sessionResetText string
	var weekAllText, weekAllResetText string
	var weekOpusText, weekOpusResetText string
	var lastUpdateText string

	if err != nil {
		sessionText = "Session    \tLoading..."
		sessionResetText = "resets: N/A"
		weekAllText = "Week (All)\tLoading..."
		weekAllResetText = "resets: N/A"
		weekOpusText = "Week (Opus)\tLoading..."
		weekOpusResetText = "resets: N/A"
		lastUpdateText = "N/A"
	} else {
		sessionText = fmt.Sprintf("Session    \t%02d%% \t%s", usage.SessionPercent, getUsageEmoji(usage.SessionPercent))
		sessionResetText = fmt.Sprintf("resets %s", removeTimezone(usage.SessionReset))
		weekAllText = fmt.Sprintf("Week (All)\t%02d%% \t%s", usage.WeekAllPercent, getUsageEmoji(usage.WeekAllPercent))
		weekAllResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekAllReset))
		weekOpusText = fmt.Sprintf("Week (Opus)\t%02d%% \t%s", usage.WeekOpusPercent, getUsageEmoji(usage.WeekOpusPercent))
		weekOpusResetText = fmt.Sprintf("resets %s", removeTimezone(usage.WeekOpusReset))
		lastUpdateText = formatTimestamp(usage.Timestamp)
	}

	// Session
	menuItems = append(menuItems, systray.AddMenuItem(sessionText, ""))
	sessionResetItem := systray.AddMenuItem(sessionResetText, "")
	sessionResetItem.Disable()
	menuItems = append(menuItems, sessionResetItem)
	systray.AddSeparator()

	// Week (All)
	menuItems = append(menuItems, systray.AddMenuItem(weekAllText, ""))
	weekAllResetItem := systray.AddMenuItem(weekAllResetText, "")
	weekAllResetItem.Disable()
	menuItems = append(menuItems, weekAllResetItem)
	systray.AddSeparator()

	// Week (Opus)
	menuItems = append(menuItems, systray.AddMenuItem(weekOpusText, ""))
	weekOpusResetItem := systray.AddMenuItem(weekOpusResetText, "")
	weekOpusResetItem.Disable()
	menuItems = append(menuItems, weekOpusResetItem)
	systray.AddSeparator()

	lastUpdateItem := systray.AddMenuItem(lastUpdateText, "")
	lastUpdateItem.Disable()
	menuItems = append(menuItems, lastUpdateItem)

	systray.AddSeparator()
}
