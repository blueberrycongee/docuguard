package reporter

import (
	"fmt"
	"io"

	"github.com/yourname/docuguard/pkg/types"
)

// GitHubReporter GitHub Actions 格式报告
type GitHubReporter struct{}

func (r *GitHubReporter) Report(w io.Writer, report *types.Report) error {
	for _, result := range report.Results {
		if !result.Consistent {
			// GitHub Actions 注释格式
			fmt.Fprintf(w, "::error file=%s,line=%d,title=文档不一致::%s\n",
				result.DocLoc.File,
				result.DocLoc.Line,
				result.Reason)
		}
	}

	// 输出汇总
	if report.Inconsistent > 0 {
		fmt.Fprintf(w, "::error::发现 %d 处文档与代码不一致\n", report.Inconsistent)
	} else {
		fmt.Fprintf(w, "::notice::所有文档与代码一致 ✅\n")
	}

	return nil
}
