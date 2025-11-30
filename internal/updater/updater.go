package updater

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion Version
	LatestVersion  Version
	ReleaseURL     string
	ReleaseName    string
	ReleaseNotes   string
	HasUpdate      bool
}

// Updater handles checking for updates
type Updater struct {
	currentVersion Version
	github         *GitHubClient
	checkInterval  time.Duration

	mu         sync.RWMutex
	lastCheck  time.Time
	lastResult *UpdateInfo

	// Callback when update is found
	OnUpdateFound func(info *UpdateInfo)
}

// Config holds updater configuration
type Config struct {
	Owner          string
	Repo           string
	CurrentVersion string
	CheckInterval  time.Duration
}

// New creates a new Updater instance
func New(cfg Config) *Updater {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 1 * time.Hour // Default: check every hour
	}

	return &Updater{
		currentVersion: ParseVersion(cfg.CurrentVersion),
		github:         NewGitHubClient(cfg.Owner, cfg.Repo),
		checkInterval:  cfg.CheckInterval,
	}
}

// CheckNow checks for updates immediately
func (u *Updater) CheckNow() (*UpdateInfo, error) {
	release, err := u.github.GetLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := ParseVersion(release.TagName)

	info := &UpdateInfo{
		CurrentVersion: u.currentVersion,
		LatestVersion:  latestVersion,
		ReleaseURL:     release.HTMLURL,
		ReleaseName:    release.Name,
		ReleaseNotes:   release.Body,
		HasUpdate:      latestVersion.IsNewerThan(u.currentVersion),
	}

	u.mu.Lock()
	u.lastCheck = time.Now()
	u.lastResult = info
	u.mu.Unlock()

	if info.HasUpdate && u.OnUpdateFound != nil {
		u.OnUpdateFound(info)
	}

	return info, nil
}

// GetLastResult returns the last check result
func (u *Updater) GetLastResult() *UpdateInfo {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.lastResult
}

// GetCurrentVersion returns the current version
func (u *Updater) GetCurrentVersion() Version {
	return u.currentVersion
}

// StartPeriodicCheck starts checking for updates periodically
func (u *Updater) StartPeriodicCheck() {
	// Check immediately on start
	go func() {
		if _, err := u.CheckNow(); err != nil {
			log.Printf("Update check failed: %v", err)
		}
	}()

	// Then check periodically
	go func() {
		ticker := time.NewTicker(u.checkInterval)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := u.CheckNow(); err != nil {
				log.Printf("Periodic update check failed: %v", err)
			}
		}
	}()
}

// OpenReleasePage opens the releases page in the default browser
func (u *Updater) OpenReleasePage() error {
	url := u.github.GetReleaseURL()
	return openURL(url)
}

// OpenLatestRelease opens the latest release page in the default browser
func (u *Updater) OpenLatestRelease() error {
	u.mu.RLock()
	result := u.lastResult
	u.mu.RUnlock()

	url := u.github.GetReleaseURL()
	if result != nil && result.ReleaseURL != "" {
		url = result.ReleaseURL
	}

	return openURL(url)
}

// openURL opens a URL in the default browser
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
