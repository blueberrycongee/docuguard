package matcher

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/blueberrycongee/docuguard/internal/llm"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

// Matcher performs relevance matching between documents and code.
type Matcher struct {
	llmClient llm.Client
}

// NewMatcher creates a new Matcher with the given LLM client.
func NewMatcher(client llm.Client) *Matcher {
	return &Matcher{llmClient: client}
}

// FindRelevantDocs finds document segments relevant to the changed symbols.
func (m *Matcher) FindRelevantDocs(
	ctx context.Context,
	symbols []types.ChangedSymbol,
	segments []types.DocSegment,
) ([]types.RelevanceResult, error) {
	var results []types.RelevanceResult

	candidates := m.preFilter(symbols, segments)

	for _, candidate := range candidates {
		result, err := m.checkRelevance(ctx, candidate.symbol, candidate.segment)
		if err != nil {
			continue
		}
		if result.IsRelevant {
			results = append(results, *result)
		}
	}

	return results, nil
}

type candidate struct {
	symbol  types.ChangedSymbol
	segment types.DocSegment
}

// preFilter performs keyword-based pre-filtering.
func (m *Matcher) preFilter(symbols []types.ChangedSymbol, segments []types.DocSegment) []candidate {
	var candidates []candidate

	for _, sym := range symbols {
		symWords := extractKeywords(sym.Name)
		symWords = append(symWords, strings.ToLower(sym.Name))

		for _, seg := range segments {
			content := strings.ToLower(seg.Content + " " + seg.Heading)

			for _, word := range symWords {
				if len(word) > 2 && strings.Contains(content, word) {
					candidates = append(candidates, candidate{
						symbol:  sym,
						segment: seg,
					})
					break
				}
			}
		}
	}

	return candidates
}

// extractKeywords extracts keywords from a symbol name.
func extractKeywords(name string) []string {
	var words []string
	var current strings.Builder

	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	return words
}

// checkRelevance uses LLM to check relevance between a symbol and segment.
func (m *Matcher) checkRelevance(
	ctx context.Context,
	symbol types.ChangedSymbol,
	segment types.DocSegment,
) (*types.RelevanceResult, error) {
	req := llm.AnalyzeRequest{
		DocContent:  segment.Content,
		CodeContent: symbol.NewCode,
		CodeSymbol:  symbol.Name,
		CodeFile:    symbol.File,
	}

	result, err := m.llmClient.Analyze(ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.RelevanceResult{
		Segment:    segment,
		Symbol:     symbol,
		IsRelevant: !result.Consistent,
		Confidence: result.Confidence,
		Reason:     result.Reason,
	}, nil
}

// QuickMatch performs fast keyword-based matching without LLM.
func QuickMatch(symbols []types.ChangedSymbol, segments []types.DocSegment) []types.RelevanceResult {
	var results []types.RelevanceResult

	for _, sym := range symbols {
		symWords := extractKeywords(sym.Name)
		symWords = append(symWords, strings.ToLower(sym.Name))

		for _, seg := range segments {
			content := strings.ToLower(seg.Content + " " + seg.Heading)
			matchCount := 0

			for _, word := range symWords {
				if len(word) > 2 && strings.Contains(content, word) {
					matchCount++
				}
			}

			if matchCount > 0 {
				confidence := float64(matchCount) / float64(len(symWords))
				results = append(results, types.RelevanceResult{
					Segment:    seg,
					Symbol:     sym,
					IsRelevant: true,
					Confidence: confidence,
					Reason:     "Keyword match",
				})
			}
		}
	}

	return results
}

// ParseRelevanceResponse parses an LLM response for relevance determination.
func ParseRelevanceResponse(response string) (bool, float64, string, error) {
	var result struct {
		Relevant   bool    `json:"relevant"`
		Confidence float64 `json:"confidence"`
		Reason     string  `json:"reason"`
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start >= 0 && end > start {
		jsonStr := response[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result.Relevant, result.Confidence, result.Reason, nil
		}
	}

	lower := strings.ToLower(response)
	relevant := strings.Contains(lower, "relevant") && !strings.Contains(lower, "not relevant")

	return relevant, 0.5, response, nil
}
