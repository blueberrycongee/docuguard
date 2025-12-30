package llm

import (
	"context"
	"fmt"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// AnthropicClient Anthropic 客户端
type AnthropicClient struct {
	model   string
	apiKey  string
	baseURL string
}

// NewAnthropicClient 创建 Anthropic 客户端
func NewAnthropicClient(model, apiKey, baseURL string) (*AnthropicClient, error) {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	return &AnthropicClient{
		model:   model,
		apiKey:  apiKey,
		baseURL: baseURL,
	}, nil
}

func (c *AnthropicClient) Name() string {
	return "anthropic"
}

func (c *AnthropicClient) Analyze(ctx context.Context, req AnalyzeRequest) (*types.CheckResult, error) {
	// TODO: 实现 Anthropic API 调用
	return nil, fmt.Errorf("anthropic client not implemented yet")
}
