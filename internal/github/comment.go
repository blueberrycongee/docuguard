package github

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CreateComment 在 PR 上创建评论
func (c *Client) CreateComment(prNumber int, body string) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", c.owner, c.repo, prNumber)

	payload := map[string]string{"body": body}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("POST", path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		decodeResponse(resp, &result)
		return fmt.Errorf("failed to create comment: %v", result)
	}

	return nil
}

// CreateReviewComment 在 PR 的具体代码行上创建评论
func (c *Client) CreateReviewComment(prNumber int, file string, line int, body string) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", c.owner, c.repo, prNumber)

	payload := map[string]interface{}{
		"body": body,
		"path": file,
		"line": line,
		"side": "RIGHT",
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("POST", path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		decodeResponse(resp, &result)
		return fmt.Errorf("failed to create review comment: %v", result)
	}

	return nil
}

// UpdateComment 更新评论
func (c *Client) UpdateComment(commentID int64, body string) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d", c.owner, c.repo, commentID)

	payload := map[string]string{"body": body}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PATCH", path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		decodeResponse(resp, &result)
		return fmt.Errorf("failed to update comment: %v", result)
	}

	return nil
}

// FindExistingComment 查找已存在的 DocuGuard 评论
func (c *Client) FindExistingComment(prNumber int) (int64, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", c.owner, c.repo, prNumber)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var comments []struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}

	if err := decodeResponse(resp, &comments); err != nil {
		return 0, err
	}

	// 查找包含 DocuGuard 标记的评论
	for _, comment := range comments {
		if bytes.Contains([]byte(comment.Body), []byte("DocuGuard")) {
			return comment.ID, nil
		}
	}

	return 0, nil
}
