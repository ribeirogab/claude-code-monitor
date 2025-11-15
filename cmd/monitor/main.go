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

	systray.SetTitle("Claude Monitor")
	systray.SetTooltip("Claude Code Usage Monitor")
	log.Println("Title and tooltip set")

	// Set a simple icon (monochrome dot)
	systray.SetIcon(getIcon())
	log.Println("Icon set")

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

func getIcon() []byte {
	// Simple monochrome icon (8x8 black dot)
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff, 0x61, 0x00, 0x00, 0x00,
		0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x49, 0x6d, 0x61, 0x67,
		0x65, 0x52, 0x65, 0x61, 0x64, 0x79, 0x71, 0xc9, 0x65, 0x3c, 0x00, 0x00,
		0x00, 0x18, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda, 0x62, 0x60, 0x18, 0x05,
		0xa3, 0x60, 0x14, 0x8c, 0x02, 0x08, 0x00, 0x00, 0x04, 0x10, 0x00, 0x01,
		0x27, 0x28, 0x4d, 0x4e, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
		0xae, 0x42, 0x60, 0x82,
	}
}
