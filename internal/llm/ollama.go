package llm

import (
	"context"
	"fmt"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// OllamaClient Ollama 本地模型客户端
type OllamaClient struct {
	model   string
	baseURL string
}

// NewOllamaClient 创建 Ollama 客户端
func NewOllamaClient(model, baseURL string) (*OllamaClient, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaClient{
		model:   model,
		baseURL: baseURL,
	}, nil
}

func (c *OllamaClient) Name() string {
	return "ollama"
}

func (c *OllamaClient) Analyze(ctx context.Context, req AnalyzeRequest) (*types.CheckResult, error) {
	// TODO: 实现 Ollama API 调用
	return nil, fmt.Errorf("ollama client not implemented yet")
}
