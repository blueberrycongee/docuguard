package llm

import (
	"context"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// MockClient 测试用 Mock 客户端
type MockClient struct {
	Result *types.CheckResult
	Err    error
}

func NewMockClient(result *types.CheckResult, err error) *MockClient {
	return &MockClient{Result: result, Err: err}
}

func (c *MockClient) Name() string {
	return "mock"
}

func (c *MockClient) Analyze(ctx context.Context, req AnalyzeRequest) (*types.CheckResult, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.Result, nil
}
