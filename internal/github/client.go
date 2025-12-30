package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Client is a GitHub API client.
type Client struct {
	token   string
	owner   string
	repo    string
	baseURL string
	http    *http.Client
}

// NewClient creates a new GitHub client.
// If repo is empty, it attempts to detect the repository from git remote.
func NewClient(token, repo string) (*Client, error) {
	owner, repoName, err := parseRepo(repo)
	if err != nil {
		owner, repoName, err = getRepoFromRemote()
		if err != nil {
			return nil, fmt.Errorf("failed to determine repository: %w", err)
		}
	}

	return &Client{
		token:   token,
		owner:   owner,
		repo:    repoName,
		baseURL: "https://api.github.com",
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// parseRepo parses an owner/repo string.
func parseRepo(repo string) (string, string, error) {
	if repo == "" {
		return "", "", fmt.Errorf("empty repo")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format, expected owner/repo")
	}

	return parts[0], parts[1], nil
}

// getRepoFromRemote extracts repository information from git remote.
func getRepoFromRemote() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSpace(string(output))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		url = url[idx+len("github.com/"):]
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("failed to parse remote URL: %s", url)
}

// doRequest executes an HTTP request with authentication.
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.http.Do(req)
}

// GetOwner returns the repository owner.
func (c *Client) GetOwner() string {
	return c.owner
}

// GetRepo returns the repository name.
func (c *Client) GetRepo() string {
	return c.repo
}

// decodeResponse decodes a JSON response.
func decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
