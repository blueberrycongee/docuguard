package llm

import (
	"context"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// Client LLM 客户端接口
type Client interface {
	// Analyze 分析文档与代码的一致性
	Analyze(ctx context.Context, req AnalyzeRequest) (*types.CheckResult, error)
	// Name 返回客户端名称
	Name() string
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	DocContent  string `json:"doc_content"`
	CodeContent string `json:"code_content"`
	CodeSymbol  string `json:"code_symbol"`
	CodeFile    string `json:"code_file"`
}

// NewClient 根据配置创建 LLM 客户端
func NewClient(provider, model, apiKey, baseURL string) (Client, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient(model, apiKey, baseURL)
	case "anthropic":
		return NewAnthropicClient(model, apiKey, baseURL)
	case "ollama":
		return NewOllamaClient(model, baseURL)
	default:
		return NewOpenAIClient(model, apiKey, baseURL)
	}
}
