package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

// OpenAIClient OpenAI 客户端
type OpenAIClient struct {
	client  *resty.Client
	model   string
	baseURL string
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(model, apiKey, baseURL string) (*OpenAIClient, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json")

	return &OpenAIClient{
		client:  client,
		model:   model,
		baseURL: baseURL,
	}, nil
}

func (c *OpenAIClient) Name() string {
	return "openai"
}

// Analyze 执行分析
func (c *OpenAIClient) Analyze(ctx context.Context, req AnalyzeRequest) (*types.CheckResult, error) {
	prompt := buildPrompt(req)

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&response).
		Post("/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	var result types.CheckResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
