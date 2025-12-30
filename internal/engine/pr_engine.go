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
	// UseTwoStage enables two-stage matching (broad match + LLM relevance filter).
	UseTwoStage bool
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

	var relevantPairs []types.RelevanceResult

	if opts.UseTwoStage && !opts.SkipLLM {
		// Two-stage matching: broad match + LLM relevance filter
		relevantPairs, err = e.twoStageMatch(ctx, symbols, segments)
		if err != nil {
			// Fallback to quick match on error
			relevantPairs = matcher.QuickMatch(symbols, segments)
		}
	} else {
		// Original quick match
		relevantPairs = matcher.QuickMatch(symbols, segments)
	}
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

// twoStageMatch performs two-stage matching:
// Stage 1: Broad keyword matching to find candidates
// Stage 2: LLM batch relevance check to filter candidates
func (e *PREngine) twoStageMatch(ctx context.Context, symbols []types.ChangedSymbol, segments []types.DocSegment) ([]types.RelevanceResult, error) {
	// Stage 1: Broad match to get candidates
	candidates := matcher.BroadMatch(symbols, segments)
	if len(candidates) == 0 {
		return nil, nil
	}

	// Group candidates by symbol for batch processing
	groups := matcher.GroupCandidatesBySymbol(candidates)

	var results []types.RelevanceResult

	// Stage 2: LLM relevance check for each symbol
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		symbol := group[0].Symbol
		candidateSegments := make([]types.DocSegment, len(group))
		for i, c := range group {
			candidateSegments[i] = c.Segment
		}

		// Batch relevance check
		req := llm.RelevanceRequest{
			Symbol:     symbol,
			Candidates: candidateSegments,
		}

		relevantIndices, err := e.llmClient.CheckRelevanceBatch(ctx, req)
		if err != nil {
			// On error, include all candidates (conservative approach)
			results = append(results, group...)
			continue
		}

		// Only include relevant candidates
		for _, idx := range relevantIndices {
			if idx >= 0 && idx < len(group) {
				group[idx].Reason = "LLM confirmed relevant"
				results = append(results, group[idx])
			}
		}
	}

	return results, nil
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
