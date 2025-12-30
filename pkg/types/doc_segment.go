package types

// DocSegment 文档段落
type DocSegment struct {
	File      string `json:"file"`       // 文件路径
	StartLine int    `json:"start_line"` // 起始行
	EndLine   int    `json:"end_line"`   // 结束行
	Heading   string `json:"heading"`    // 所属标题
	Content   string `json:"content"`    // 段落内容
	Type      string `json:"type"`       // markdown / godoc
	Level     int    `json:"level"`      // 标题级别 (1-6)
}

// RelevanceResult 相关性判断结果
type RelevanceResult struct {
	Segment    DocSegment    `json:"segment"`
	Symbol     ChangedSymbol `json:"symbol"`
	IsRelevant bool          `json:"is_relevant"`
	Confidence float64       `json:"confidence"`
	Reason     string        `json:"reason"`
}

// PRReport PR 检查报告
type PRReport struct {
	TotalSymbols    int              `json:"total_symbols"`
	TotalSegments   int              `json:"total_segments"`
	RelevantPairs   int              `json:"relevant_pairs"`
	Inconsistent    int              `json:"inconsistent"`
	Results         []PRCheckResult  `json:"results"`
	ExecutionTimeMs int64            `json:"execution_time_ms"`
}

// PRCheckResult PR 检查单项结果
type PRCheckResult struct {
	Segment     DocSegment    `json:"segment"`
	Symbol      ChangedSymbol `json:"symbol"`
	Consistent  bool          `json:"consistent"`
	Confidence  float64       `json:"confidence"`
	Reason      string        `json:"reason"`
	Suggestion  string        `json:"suggestion,omitempty"`
}
