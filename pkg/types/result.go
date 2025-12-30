package types

// CheckResult 检查结果
type CheckResult struct {
	Consistent bool     `json:"consistent"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
	DocLoc     Location `json:"doc_location"`
	CodeLoc    Location `json:"code_location"`
	Suggestion string   `json:"suggestion,omitempty"`
}

// Severity 严重程度
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Report 完整报告
type Report struct {
	TotalBindings   int           `json:"total_bindings"`
	Consistent      int           `json:"consistent"`
	Inconsistent    int           `json:"inconsistent"`
	Errors          int           `json:"errors"`
	Results         []CheckResult `json:"results"`
	ExecutionTimeMs int64         `json:"execution_time_ms"`
}
