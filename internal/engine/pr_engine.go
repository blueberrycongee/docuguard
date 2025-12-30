package engine

import (
	"context"
	"time"

	"github.com/blueberrycongee/docuguard/internal/config"
	"github.com/blueberrycongee/docuguard/internal/git"
	"github.com/blueberrycongee/docuguard/internal/llm"
	"github.com/blueberrycongee/docuguard/internal/matcher"
	"github.com/blueberrycongee/docuguard/internal/scanner"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

// PREngine handles PR-based documentation consistency checking.
type PREngine struct {
	cfg       *config.Config
	llmClient llm.Client
}

// PRCheckOptions contains options for PR checking.
type PRCheckOptions struct {
	// BaseBranch is the base branch for comparison.
	BaseBranch string
	// DocPatterns are glob patterns for documentation files.
	DocPatterns []string
	// SkipLLM skips LLM analysis and uses keyword matching only.
	SkipLLM bool
}

// NewPREngine creates a new PREngine with the given configuration.
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
	}, nil
}

// CheckFromDiff performs a consistency check from diff content.
func (e *PREngine) CheckFromDiff(ctx context.Context, diffContent string, opts PRCheckOptions) (*types.PRReport, error) {
	startTime := time.Now()
	report := &types.PRReport{}

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

	segments, err := scanner.ScanMarkdownDir(".", opts.DocPatterns)
	if err != nil {
		return nil, err
	}
	report.TotalSegments = len(segments)

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

// checkConsistency checks consistency between a document segment and code symbol.
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
		result.Related = true
		result.Consistent = true
		result.Confidence = 0.5
		result.Reason = "Keyword match only, LLM check skipped"
		return result
	}

	req := llm.AnalyzeRequest{
		DocContent:  segment.Content,
		CodeContent: symbol.NewCode,
		CodeSymbol:  symbol.Name,
		CodeFile:    symbol.File,
	}

	llmResult, err := e.llmClient.Analyze(ctx, req)
	if err != nil {
		result.Related = true
		result.Consistent = true
		result.Confidence = 0.0
		result.Reason = "LLM check failed: " + err.Error()
		return result
	}

	result.Related = llmResult.Related
	result.Consistent = llmResult.Consistent
	result.Confidence = llmResult.Confidence
	result.Reason = llmResult.Reason
	result.Suggestion = llmResult.Suggestion

	return result
}
