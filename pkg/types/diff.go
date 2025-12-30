package types

// ChangeType 变更类型
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

// ChangedSymbol 表示代码中变更的符号
type ChangedSymbol struct {
	File       string      `json:"file"`        // 文件路径
	Name       string      `json:"name"`        // 符号名称（函数名、结构体名等）
	Type       BindingType `json:"type"`        // func / struct / const / var
	OldCode    string      `json:"old_code"`    // 变更前的代码
	NewCode    string      `json:"new_code"`    // 变更后的代码
	ChangeType ChangeType  `json:"change_type"` // added / modified / deleted
	StartLine  int         `json:"start_line"`  // 起始行号
	EndLine    int         `json:"end_line"`    // 结束行号
}

// FileDiff 文件级别的 diff 信息
type FileDiff struct {
	OldPath      string       `json:"old_path"`
	NewPath      string       `json:"new_path"`
	ChangeType   ChangeType   `json:"change_type"`
	ChangedLines []LineChange `json:"changed_lines"`
}

// LineChange 行变更信息
type LineChange struct {
	OldStart int `json:"old_start"`
	OldCount int `json:"old_count"`
	NewStart int `json:"new_start"`
	NewCount int `json:"new_count"`
}
