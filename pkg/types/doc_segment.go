package types

// DocSegment represents a section of documentation.
type DocSegment struct {
	// File is the path to the documentation file.
	File string `json:"file"`
	// StartLine is the starting line number.
	StartLine int `json:"start_line"`
	// EndLine is the ending line number.
	EndLine int `json:"end_line"`
	// Heading is the section heading.
	Heading string `json:"heading"`
	// Content is the full content of the section.
	Content string `json:"content"`
	// Type is the document type (markdown, godoc).
	Type string `json:"type"`
	// Level is the heading level (1-6 for markdown).
	Level int `json:"level"`
}

// RelevanceResult represents the result of a relevance check
// between a document segment and a code symbol.
type RelevanceResult struct {
	// Segment is the documentation segment.
	Segment DocSegment `json:"segment"`
	// Symbol is the changed code symbol.
	Symbol ChangedSymbol `json:"symbol"`
	// IsRelevant indicates whether the segment is relevant to the symbol.
	IsRelevant bool `json:"is_relevant"`
	// Confidence is the confidence score (0-1).
	Confidence float64 `json:"confidence"`
	// Reason explains the relevance determination.
	Reason string `json:"reason"`
}

// PRReport represents the complete PR check report.
type PRReport struct {
	// TotalSymbols is the number of changed symbols detected.
	TotalSymbols int `json:"total_symbols"`
	// TotalSegments is the number of document segments scanned.
	TotalSegments int `json:"total_segments"`
	// RelevantPairs is the number of relevant document-code pairs found.
	RelevantPairs int `json:"relevant_pairs"`
	// Inconsistent is the number of inconsistencies found.
	Inconsistent int `json:"inconsistent"`
	// Results contains the individual check results.
	Results []PRCheckResult `json:"results"`
	// ExecutionTimeMs is the execution time in milliseconds.
	ExecutionTimeMs int64 `json:"execution_time_ms"`
}

// PRCheckResult represents a single PR check result.
type PRCheckResult struct {
	// Segment is the documentation segment checked.
	Segment DocSegment `json:"segment"`
	// Symbol is the code symbol checked against.
	Symbol ChangedSymbol `json:"symbol"`
	// Related indicates whether the documentation is about this specific code.
	Related bool `json:"related"`
	// Consistent indicates whether the documentation matches the code.
	Consistent bool `json:"consistent"`
	// Confidence is the confidence score (0-1).
	Confidence float64 `json:"confidence"`
	// Reason explains the consistency determination.
	Reason string `json:"reason"`
	// Suggestion provides a recommendation for fixing inconsistencies.
	Suggestion string `json:"suggestion,omitempty"`
}
