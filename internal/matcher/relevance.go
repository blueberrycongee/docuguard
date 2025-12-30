package matcher

import (
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

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
