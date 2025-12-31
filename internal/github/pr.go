package github

import (
	"fmt"
	"io"
	"net/http"
)

// PRInfo contains information about a pull request.
type PRInfo struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	BaseBranch string `json:"base_branch"`
	HeadBranch string `json:"head_branch"`
	HTMLURL    string `json:"html_url"`
	DiffURL    string `json:"diff_url"`
}

// PRFile represents a file changed in a pull request.
type PRFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"` // added, removed, modified, renamed
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

// GetPRInfo retrieves information about a pull request.
func (c *Client) GetPRInfo(prNumber int) (*PRInfo, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", c.owner, c.repo, prNumber)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var pr struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		DiffURL string `json:"diff_url"`
		Base    struct {
			Ref string `json:"ref"`
		} `json:"base"`
		Head struct {
			Ref string `json:"ref"`
		} `json:"head"`
	}

	if err := decodeResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &PRInfo{
		Number:     pr.Number,
		Title:      pr.Title,
		State:      pr.State,
		BaseBranch: pr.Base.Ref,
		HeadBranch: pr.Head.Ref,
		HTMLURL:    pr.HTMLURL,
		DiffURL:    pr.DiffURL,
	}, nil
}

// GetPRFiles retrieves the list of files changed in a pull request.
func (c *Client) GetPRFiles(prNumber int) ([]PRFile, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/files", c.owner, c.repo, prNumber)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var files []PRFile
	if err := decodeResponse(resp, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// GetPRDiff retrieves the diff content of a pull request.
func (c *Client) GetPRDiff(prNumber int) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, c.owner, c.repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// BuildDiffFromFiles constructs a diff string from PR files.
func BuildDiffFromFiles(files []PRFile) string {
	var diff string
	for _, f := range files {
		if f.Patch != "" {
			diff += fmt.Sprintf("diff --git a/%s b/%s\n", f.Filename, f.Filename)
			diff += f.Patch + "\n"
		}
	}
	return diff
}
