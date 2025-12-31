package git

import (
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// SymbolExtractor extracts changed symbols from git diffs.
type SymbolExtractor struct{}

// NewSymbolExtractor creates a new SymbolExtractor.
func NewSymbolExtractor() *SymbolExtractor {
	return &SymbolExtractor{}
}

// ExtractChangedSymbols extracts changed Go symbols from diff content.
func (e *SymbolExtractor) ExtractChangedSymbols(diffContent string) ([]types.ChangedSymbol, error) {
	// Use the new method that parses diff content directly
	fileDiffs, err := ParseDiffWithContent(diffContent)
	if err != nil {
		return nil, err
	}

	goFiles := FilterGoFiles(fileDiffs)

	var symbols []types.ChangedSymbol
	for _, fd := range goFiles {
		// Extract symbols from added lines (new/modified code)
		addedSymbols := ExtractSymbolsFromDiffLines(fd.AddedLines)
		for _, name := range addedSymbols {
			sym := types.ChangedSymbol{
				File:       fd.NewPath,
				Name:       name,
				Type:       guessSymbolType(name, fd.AddedLines),
				ChangeType: fd.ChangeType,
			}
			// Extract code directly from diff lines (not from local file)
			sym.NewCode = extractSymbolCodeFromLines(name, fd.AddedLines)
			sym.OldCode = extractSymbolCodeFromLines(name, fd.RemovedLines)
			symbols = append(symbols, sym)
		}

		// For deleted files, extract from removed lines
		if fd.ChangeType == types.ChangeDeleted {
			removedSymbols := ExtractSymbolsFromDiffLines(fd.RemovedLines)
			for _, name := range removedSymbols {
				sym := types.ChangedSymbol{
					File:       fd.OldPath,
					Name:       name,
					Type:       guessSymbolType(name, fd.RemovedLines),
					ChangeType: types.ChangeDeleted,
					OldCode:    extractSymbolCodeFromLines(name, fd.RemovedLines),
				}
				symbols = append(symbols, sym)
			}
		}
	}

	return symbols, nil
}

// extractSymbolCodeFromLines extracts the code for a symbol from diff lines.
func extractSymbolCodeFromLines(symbolName string, lines []string) string {
	var codeLines []string
	inSymbol := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line contains the symbol definition
		if strings.Contains(line, symbolName) && strings.Contains(line, "=") {
			inSymbol = true
		}

		if inSymbol {
			codeLines = append(codeLines, trimmed)
			// For simple const/var assignments, one line is enough
			if strings.Contains(line, "=") && !strings.HasSuffix(trimmed, "{") {
				break
			}
		}

		// Also capture comment lines before the symbol
		if strings.HasPrefix(trimmed, "//") && strings.Contains(trimmed, symbolName) {
			codeLines = append(codeLines, trimmed)
		}
	}

	return strings.Join(codeLines, "\n")
}

// guessSymbolType tries to determine the symbol type from context.
func guessSymbolType(name string, lines []string) types.BindingType {
	for _, line := range lines {
		if funcDeclRegex.MatchString(line) && strings.Contains(line, name) {
			return types.BindingFunc
		}
		if typeRegex.MatchString(line) && strings.Contains(line, name) {
			return types.BindingStruct
		}
	}
	// Default to const for simple assignments
	return types.BindingConst
}
