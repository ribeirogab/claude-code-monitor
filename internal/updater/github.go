package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	githubAPIURL    = "https://api.github.com/repos/%s/%s/releases/latest"
	defaultTimeout  = 10 * time.Second
)

// GitHubRelease represents a GitHub release from the API
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// GitHubClient handles GitHub API requests
type GitHubClient struct {
	Owner   string
	Repo    string
	client  *http.Client
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(owner, repo string) *GitHubClient {
	return &GitHubClient{
		Owner: owner,
		Repo:  repo,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (g *GitHubClient) GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, g.Owner, g.Repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "claude-code-monitor")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// GetReleaseURL returns the URL to the releases page
func (g *GitHubClient) GetReleaseURL() string {
	return fmt.Sprintf("https://github.com/%s/%s/releases", g.Owner, g.Repo)
}
