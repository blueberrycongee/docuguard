package reporter

import (
	"fmt"
	"strings"

	"github.com/yourname/docuguard/pkg/types"
)

// FormatPRComment 格式化 PR 评论
func FormatPRComment(report *types.PRReport, repoURL string) string {
	var sb strings.Builder

	sb.WriteString("## 📋 DocuGuard 检查报告\n\n")

	if len(report.Results) == 0 {
		sb.WriteString("✅ 未发现文档与代码不一致的问题。\n\n")
		sb.WriteString(fmt.Sprintf("- 检测到 **%d** 个代码变更\n", report.TotalSymbols))
		sb.WriteString(fmt.Sprintf("- 扫描了 **%d** 个文档段落\n", report.TotalSegments))
		sb.WriteString(fmt.Sprintf("- 耗时 **%d** ms\n", report.ExecutionTimeMs))
	} else {
		// 统计
		inconsistentCount := 0
		for _, r := range report.Results {
			if !r.Consistent {
				inconsistentCount++
			}
		}

		if inconsistentCount > 0 {
			sb.WriteString(fmt.Sprintf("检测到 **%d** 处文档可能需要更新：\n\n", inconsistentCount))
			sb.WriteString("### ❌ 不一致\n\n")
			sb.WriteString("| 文档 | 代码 | 问题 |\n")
			sb.WriteString("|------|------|------|\n")

			for _, r := range report.Results {
				if !r.Consistent {
					docLink := formatDocLink(r.Segment.File, r.Segment.StartLine, repoURL)
					sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n",
						docLink,
						r.Symbol.Name,
						truncate(r.Reason, 50),
					))
				}
			}
			sb.WriteString("\n")
		}

		// 建议检查的项目
		suggestCount := len(report.Results) - inconsistentCount
		if suggestCount > 0 {
			sb.WriteString("### ⚠️ 建议检查\n\n")
			sb.WriteString("| 文档 | 相关代码 | 原因 |\n")
			sb.WriteString("|------|----------|------|\n")

			for _, r := range report.Results {
				if r.Consistent {
					docLink := formatDocLink(r.Segment.File, r.Segment.StartLine, repoURL)
					sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n",
						docLink,
						r.Symbol.Name,
						truncate(r.Reason, 50),
					))
				}
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("---\n")
	sb.WriteString("<sub>Powered by [DocuGuard](https://github.com/yourname/docuguard)</sub>\n")

	return sb.String()
}

// FormatPRCommentCompact 格式化紧凑版 PR 评论
func FormatPRCommentCompact(report *types.PRReport) string {
	var sb strings.Builder

	sb.WriteString("## 📋 DocuGuard\n\n")

	if report.Inconsistent == 0 {
		sb.WriteString("✅ 文档与代码一致\n")
	} else {
		sb.WriteString(fmt.Sprintf("⚠️ 发现 %d 处不一致\n\n", report.Inconsistent))

		for _, r := range report.Results {
			if !r.Consistent {
				sb.WriteString(fmt.Sprintf("- **%s** ↔ `%s`: %s\n",
					r.Segment.Heading,
					r.Symbol.Name,
					r.Reason,
				))
			}
		}
	}

	return sb.String()
}

// formatDocLink 格式化文档链接
func formatDocLink(file string, line int, repoURL string) string {
	if repoURL != "" {
		return fmt.Sprintf("[%s#L%d](%s/blob/HEAD/%s#L%d)", file, line, repoURL, file, line)
	}
	return fmt.Sprintf("%s#L%d", file, line)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FormatDetailedResult 格式化详细结果（用于 review comment）
func FormatDetailedResult(result types.PRCheckResult) string {
	var sb strings.Builder

	sb.WriteString("### DocuGuard 检测到潜在问题\n\n")
	sb.WriteString(fmt.Sprintf("**相关文档**: %s (第 %d 行)\n\n", result.Segment.File, result.Segment.StartLine))
	sb.WriteString(fmt.Sprintf("**标题**: %s\n\n", result.Segment.Heading))
	sb.WriteString(fmt.Sprintf("**问题**: %s\n\n", result.Reason))

	if result.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("**建议**: %s\n", result.Suggestion))
	}

	return sb.String()
}
