package reporter

import (
	"io"

	"github.com/yourname/docuguard/pkg/types"
)

// Reporter 报告输出接口
type Reporter interface {
	// Report 输出报告
	Report(w io.Writer, report *types.Report) error
}

// New 根据格式创建 Reporter
func New(format string, color bool) Reporter {
	switch format {
	case "json":
		return &JSONReporter{}
	case "github-actions":
		return &GitHubReporter{}
	default:
		return &TextReporter{Color: color}
	}
}
