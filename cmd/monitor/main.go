package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/ribeirogab/claude-code-monitor/internal/executor"
	"github.com/ribeirogab/claude-code-monitor/internal/scheduler"
)

var (
	sched *scheduler.Scheduler
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

	// Add Quit menu item
	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	log.Println("Menu item added")

	// Setup executor and scheduler
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	scriptPath := findScriptPath()
	log.Printf("Script path: %s", scriptPath)

	outputDir := filepath.Join(homeDir, ".claude-code-monitor")
	log.Printf("Output directory: %s", outputDir)

	exec := executor.New(scriptPath, outputDir)

	// Create scheduler with 1-minute interval
	sched = scheduler.New(time.Minute, exec.Execute)
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
