package reporter

import (
	"encoding/json"
	"io"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// JSONReporter JSON 格式报告
type JSONReporter struct{}

func (r *JSONReporter) Report(w io.Writer, report *types.Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
