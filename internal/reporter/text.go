package reporter

import (
	"fmt"
	"io"

	"github.com/blueberrycongee/docuguard/pkg/types"
	"github.com/fatih/color"
)

// TextReporter outputs results in human-readable text format.
type TextReporter struct {
	Color bool
}

// Report writes the check report to the given writer.
func (r *TextReporter) Report(w io.Writer, report *types.Report) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if !r.Color {
		color.NoColor = true
	}

	fmt.Fprintf(w, "\nDocuGuard Check Report\n")
	fmt.Fprintf(w, "========================================\n\n")

	for i, result := range report.Results {
		status := green("[PASS]")
		if !result.Consistent {
			status = red("[FAIL]")
		}

		fmt.Fprintf(w, "[%d] %s\n", i+1, status)
		fmt.Fprintf(w, "  Doc:        %s:%d\n", result.DocLoc.File, result.DocLoc.Line)
		fmt.Fprintf(w, "  Code:       %s:%d\n", result.CodeLoc.File, result.CodeLoc.Line)
		fmt.Fprintf(w, "  Confidence: %.0f%%\n", result.Confidence*100)
		fmt.Fprintf(w, "  Reason:     %s\n", result.Reason)

		if result.Suggestion != "" {
			fmt.Fprintf(w, "  Suggestion: %s\n", yellow(result.Suggestion))
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "========================================\n")
	fmt.Fprintf(w, "Summary: %d bindings, %s passed, %s failed\n",
		report.TotalBindings,
		green(fmt.Sprintf("%d", report.Consistent)),
		red(fmt.Sprintf("%d", report.Inconsistent)))
	fmt.Fprintf(w, "Time: %dms\n\n", report.ExecutionTimeMs)

	return nil
}
