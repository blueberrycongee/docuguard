package engine

import (
	"context"
	"time"

	"github.com/yourname/docuguard/internal/config"
	"github.com/yourname/docuguard/internal/git"
	"github.com/yourname/docuguard/internal/llm"
	"github.com/yourname/docuguard/internal/matcher"
	"github.com/yourname/docuguard/internal/scanner"
	"github.com/yourname/docuguard/pkg/types"
)

// PREngine PR 检查引擎
type PREngine struct {
	cfg       *config.Config
	llmClient llm.Client
	matcher   *matcher.Matcher
}

// PRCheckOptions PR 检查选项
type PRCheckOptions struct {
	BaseBranch  string   // 基准分支
	DocPatterns []string // 文档匹配模式
	SkipLLM     bool     // 跳过 LLM 检查（仅使用关键词匹配）
}

// NewPREngine 创建 PR 检查引擎
func NewPREngine(cfg *config.Config) (*PREngine, error) {
	client, err := llm.NewClient(
		cfg.LLM.Provider,
		cfg.LLM.Model,
		cfg.LLM.APIKey,
		cfg.LLM.BaseURL,
	)
	if err != nil {
		return nil, err
	}

	return &PREngine{
		cfg:       cfg,
		llmClient: client,
		matcher:   matcher.NewMatcher(client),
	}, nil
}

// Check 执行 PR 检查
func (e *PREngine) Check(ctx context.Context, opts PRCheckOptions) (*types.PRReport, error) {
	startTime := time.Now()
	report := &types.PRReport{}

	// 1. 获取 diff
	diff, err := git.GetDiff(opts.BaseBranch)
	if err != nil {
		// 尝试获取未提交的变更
		diff, err = git.GetDiffUncommitted()
		if err != nil {
			return nil, err
		}
	}

	if diff == "" {
		report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return report, nil
	}

	// 2. 提取变更符号
	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diff)
	if err != nil {
		return nil, err
	}
	report.TotalSymbols = len(symbols)

	if len(symbols) == 0 {
		report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return report, nil
	}

	// 3. 扫描文档
	segments, err := scanner.ScanMarkdownDir(".", opts.DocPatterns)
	if err != nil {
		return nil, err
	}
	report.TotalSegments = len(segments)

	if len(segments) == 0 {
		report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return report, nil
	}

	// 4. 找相关文档
	var relevantPairs []types.RelevanceResult
	if opts.SkipLLM {
		// 仅使用关键词匹配
		relevantPairs = matcher.QuickMatch(symbols, segments)
	} else {
		// 使用 LLM 进行相关性判断
		relevantPairs, err = e.matcher.FindRelevantDocs(ctx, symbols, segments)
		if err != nil {
			// 降级到关键词匹配
			relevantPairs = matcher.QuickMatch(symbols, segments)
		}
	}
	report.RelevantPairs = len(relevantPairs)

	// 5. 检查一致性
	for _, pair := range relevantPairs {
		result := e.checkConsistency(ctx, pair.Segment, pair.Symbol, opts.SkipLLM)
		report.Results = append(report.Results, result)
		if !result.Consistent {
			report.Inconsistent++
		}
	}

	report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return report, nil
}

// checkConsistency 检查单个文档-代码对的一致性
func (e *PREngine) checkConsistency(
	ctx context.Context,
	segment types.DocSegment,
	symbol types.ChangedSymbol,
	skipLLM bool,
) types.PRCheckResult {
	result := types.PRCheckResult{
		Segment: segment,
		Symbol:  symbol,
	}

	if skipLLM {
		// 简单的关键词检查
		result.Consistent = true
		result.Confidence = 0.5
		result.Reason = "Keyword match only, LLM check skipped"
		return result
	}

	// 使用 LLM 检查一致性
	req := llm.AnalyzeRequest{
		DocContent:  segment.Content,
		CodeContent: symbol.NewCode,
		CodeSymbol:  symbol.Name,
		CodeFile:    symbol.File,
	}

	llmResult, err := e.llmClient.Analyze(ctx, req)
	if err != nil {
		result.Consistent = true
		result.Confidence = 0.0
		result.Reason = "LLM check failed: " + err.Error()
		return result
	}

	result.Consistent = llmResult.Consistent
	result.Confidence = llmResult.Confidence
	result.Reason = llmResult.Reason
	result.Suggestion = llmResult.Suggestion

	return result
}

// CheckFromDiff 从 diff 字符串直接检查
func (e *PREngine) CheckFromDiff(ctx context.Context, diffContent string, opts PRCheckOptions) (*types.PRReport, error) {
	startTime := time.Now()
	report := &types.PRReport{}

	// 提取变更符号
	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diffContent)
	if err != nil {
		return nil, err
	}
	report.TotalSymbols = len(symbols)

	if len(symbols) == 0 {
		report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return report, nil
	}

	// 扫描文档
	segments, err := scanner.ScanMarkdownDir(".", opts.DocPatterns)
	if err != nil {
		return nil, err
	}
	report.TotalSegments = len(segments)

	// 找相关文档并检查一致性
	relevantPairs := matcher.QuickMatch(symbols, segments)
	report.RelevantPairs = len(relevantPairs)

	for _, pair := range relevantPairs {
		result := e.checkConsistency(ctx, pair.Segment, pair.Symbol, opts.SkipLLM)
		report.Results = append(report.Results, result)
		if !result.Consistent {
			report.Inconsistent++
		}
	}

	report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return report, nil
}
