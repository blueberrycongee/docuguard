package matcher

import (
	"regexp"
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// codeBlockRegex matches fenced code blocks.
var codeBlockRegex = regexp.MustCompile("(?s)```[a-z]*\n(.*?)```")

// BroadMatch performs broad keyword-based matching to find candidate document segments.
// It uses multiple matching strategies to avoid missing relevant documents.
// Returns all candidates without filtering by confidence threshold.
func BroadMatch(symbols []types.ChangedSymbol, segments []types.DocSegment) []types.RelevanceResult {
	var results []types.RelevanceResult

	for _, sym := range symbols {
		symLower := strings.ToLower(sym.Name)
		symWords := extractKeywords(sym.Name)

		for _, seg := range segments {
			score := 0.0
			var reasons []string

			contentLower := strings.ToLower(seg.Content)
			headingLower := strings.ToLower(seg.Heading)

			// Strategy 1: Exact symbol name match (highest priority)
			if strings.Contains(contentLower, symLower) || strings.Contains(headingLower, symLower) {
				score += 1.0
				reasons = append(reasons, "exact name match")
			}

			// Strategy 2: Match in code blocks (high priority)
			codeBlocks := codeBlockRegex.FindAllStringSubmatch(seg.Content, -1)
			for _, block := range codeBlocks {
				if len(block) > 1 && strings.Contains(strings.ToLower(block[1]), symLower) {
					score += 0.8
					reasons = append(reasons, "found in code block")
					break
				}
			}

			// Strategy 3: Keyword match from CamelCase split
			matchedWords := 0
			for _, word := range symWords {
				if len(word) > 2 && (strings.Contains(contentLower, word) || strings.Contains(headingLower, word)) {
					matchedWords++
				}
			}
			if matchedWords > 0 {
				keywordScore := float64(matchedWords) / float64(len(symWords)) * 0.5
				score += keywordScore
				reasons = append(reasons, "keyword match")
			}

			// Strategy 4: Partial name match (e.g., "Minimum" in "MinimumNArgs")
			if len(sym.Name) > 5 {
				prefix := strings.ToLower(sym.Name[:len(sym.Name)/2])
				if len(prefix) > 3 && strings.Contains(contentLower, prefix) {
					score += 0.3
					reasons = append(reasons, "partial name match")
				}
			}

			// Include if any match found
			if score > 0 {
				results = append(results, types.RelevanceResult{
					Segment:    seg,
					Symbol:     sym,
					IsRelevant: true,
					Confidence: min(score, 1.0),
					Reason:     strings.Join(reasons, ", "),
				})
			}
		}
	}

	return results
}

// GroupCandidatesBySymbol groups candidates by symbol for batch LLM processing.
func GroupCandidatesBySymbol(candidates []types.RelevanceResult) map[string][]types.RelevanceResult {
	groups := make(map[string][]types.RelevanceResult)
	for _, c := range candidates {
		key := c.Symbol.File + ":" + c.Symbol.Name
		groups[key] = append(groups[key], c)
	}
	return groups
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
