// Package types defines core data structures used throughout DocuGuard.
package types

// ChangeType represents the type of change in a diff.
type ChangeType string

const (
	// ChangeAdded indicates a newly added symbol.
	ChangeAdded ChangeType = "added"
	// ChangeModified indicates a modified symbol.
	ChangeModified ChangeType = "modified"
	// ChangeDeleted indicates a deleted symbol.
	ChangeDeleted ChangeType = "deleted"
)

// ChangedSymbol represents a code symbol that has been changed.
type ChangedSymbol struct {
	// File is the path to the file containing the symbol.
	File string `json:"file"`
	// Name is the symbol name (function, struct, etc.).
	Name string `json:"name"`
	// Type is the symbol type (func, struct, const, var).
	Type BindingType `json:"type"`
	// OldCode is the code before the change.
	OldCode string `json:"old_code"`
	// NewCode is the code after the change.
	NewCode string `json:"new_code"`
	// ChangeType indicates whether the symbol was added, modified, or deleted.
	ChangeType ChangeType `json:"change_type"`
	// StartLine is the starting line number of the symbol.
	StartLine int `json:"start_line"`
	// EndLine is the ending line number of the symbol.
	EndLine int `json:"end_line"`
}

// FileDiff represents diff information for a single file.
type FileDiff struct {
	// OldPath is the file path before the change.
	OldPath string `json:"old_path"`
	// NewPath is the file path after the change.
	NewPath string `json:"new_path"`
	// ChangeType indicates the type of file change.
	ChangeType ChangeType `json:"change_type"`
	// ChangedLines contains the line change information.
	ChangedLines []LineChange `json:"changed_lines"`
	// AddedLines contains the actual added line content from diff.
	AddedLines []string `json:"added_lines,omitempty"`
	// RemovedLines contains the actual removed line content from diff.
	RemovedLines []string `json:"removed_lines,omitempty"`
}

// LineChange represents a hunk of changed lines in a diff.
type LineChange struct {
	// OldStart is the starting line in the old file.
	OldStart int `json:"old_start"`
	// OldCount is the number of lines in the old file.
	OldCount int `json:"old_count"`
	// NewStart is the starting line in the new file.
	NewStart int `json:"new_start"`
	// NewCount is the number of lines in the new file.
	NewCount int `json:"new_count"`
}
