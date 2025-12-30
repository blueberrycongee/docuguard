package git

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// SymbolExtractor extracts changed symbols from git diffs.
type SymbolExtractor struct {
	fset *token.FileSet
}

// NewSymbolExtractor creates a new SymbolExtractor.
func NewSymbolExtractor() *SymbolExtractor {
	return &SymbolExtractor{
		fset: token.NewFileSet(),
	}
}

// ExtractChangedSymbols extracts changed Go symbols from diff content.
func (e *SymbolExtractor) ExtractChangedSymbols(diffContent string) ([]types.ChangedSymbol, error) {
	fileDiffs, err := ParseDiff(diffContent)
	if err != nil {
		return nil, err
	}

	goFiles := FilterGoFiles(fileDiffs)

	var symbols []types.ChangedSymbol
	for _, fd := range goFiles {
		fileSymbols, err := e.extractSymbolsFromFile(fd)
		if err != nil {
			continue
		}
		symbols = append(symbols, fileSymbols...)
	}

	return symbols, nil
}

// extractSymbolsFromFile extracts changed symbols from a single file diff.
func (e *SymbolExtractor) extractSymbolsFromFile(fd types.FileDiff) ([]types.ChangedSymbol, error) {
	switch fd.ChangeType {
	case types.ChangeDeleted:
		oldContent, err := e.getOldFileContent(fd.OldPath)
		if err != nil {
			return nil, err
		}
		return e.extractAllSymbols(fd.OldPath, oldContent, types.ChangeDeleted)

	case types.ChangeAdded:
		content, err := os.ReadFile(fd.NewPath)
		if err != nil {
			return nil, err
		}
		return e.extractAllSymbols(fd.NewPath, string(content), types.ChangeAdded)

	case types.ChangeModified:
		return e.extractModifiedSymbols(fd)
	}

	return nil, nil
}

// extractModifiedSymbols extracts symbols that were modified in a file.
func (e *SymbolExtractor) extractModifiedSymbols(fd types.FileDiff) ([]types.ChangedSymbol, error) {
	content, err := os.ReadFile(fd.NewPath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fd.NewPath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	changedRanges := make(map[int]bool)
	for _, lc := range fd.ChangedLines {
		for i := lc.NewStart; i < lc.NewStart+lc.NewCount; i++ {
			changedRanges[i] = true
		}
	}

	var symbols []types.ChangedSymbol

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		var sym *types.ChangedSymbol

		switch node := n.(type) {
		case *ast.FuncDecl:
			startLine := fset.Position(node.Pos()).Line
			endLine := fset.Position(node.End()).Line
			if e.overlapsWithChanges(startLine, endLine, changedRanges) {
				code := e.nodeToString(fset, node)
				sym = &types.ChangedSymbol{
					File:       fd.NewPath,
					Name:       node.Name.Name,
					Type:       types.BindingFunc,
					NewCode:    code,
					ChangeType: types.ChangeModified,
					StartLine:  startLine,
					EndLine:    endLine,
				}
			}

		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if _, isStruct := ts.Type.(*ast.StructType); isStruct {
							startLine := fset.Position(ts.Pos()).Line
							endLine := fset.Position(ts.End()).Line
							if e.overlapsWithChanges(startLine, endLine, changedRanges) {
								code := e.nodeToString(fset, ts)
								sym = &types.ChangedSymbol{
									File:       fd.NewPath,
									Name:       ts.Name.Name,
									Type:       types.BindingStruct,
									NewCode:    code,
									ChangeType: types.ChangeModified,
									StartLine:  startLine,
									EndLine:    endLine,
								}
							}
						}
					}
				}
			} else if node.Tok == token.CONST || node.Tok == token.VAR {
				for _, spec := range node.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						startLine := fset.Position(vs.Pos()).Line
						endLine := fset.Position(vs.End()).Line
						if e.overlapsWithChanges(startLine, endLine, changedRanges) {
							code := e.nodeToString(fset, vs)
							symType := types.BindingConst
							if node.Tok == token.VAR {
								symType = types.BindingVar
							}
							for _, name := range vs.Names {
								sym = &types.ChangedSymbol{
									File:       fd.NewPath,
									Name:       name.Name,
									Type:       symType,
									NewCode:    code,
									ChangeType: types.ChangeModified,
									StartLine:  startLine,
									EndLine:    endLine,
								}
							}
						}
					}
				}
			}
		}

		if sym != nil {
			oldCode, _ := e.getOldSymbolCode(fd.OldPath, sym.Name, sym.Type)
			sym.OldCode = oldCode
			symbols = append(symbols, *sym)
		}

		return true
	})

	return symbols, nil
}

// overlapsWithChanges checks if a line range overlaps with changed lines.
func (e *SymbolExtractor) overlapsWithChanges(startLine, endLine int, changedRanges map[int]bool) bool {
	for line := startLine; line <= endLine; line++ {
		if changedRanges[line] {
			return true
		}
	}
	return false
}

// extractAllSymbols extracts all symbols from file content.
func (e *SymbolExtractor) extractAllSymbols(filePath, content string, changeType types.ChangeType) ([]types.ChangedSymbol, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []types.ChangedSymbol

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			code := e.nodeToString(fset, node)
			sym := types.ChangedSymbol{
				File:       filePath,
				Name:       node.Name.Name,
				Type:       types.BindingFunc,
				ChangeType: changeType,
				StartLine:  fset.Position(node.Pos()).Line,
				EndLine:    fset.Position(node.End()).Line,
			}
			if changeType == types.ChangeAdded {
				sym.NewCode = code
			} else {
				sym.OldCode = code
			}
			symbols = append(symbols, sym)

		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if _, isStruct := ts.Type.(*ast.StructType); isStruct {
							code := e.nodeToString(fset, ts)
							sym := types.ChangedSymbol{
								File:       filePath,
								Name:       ts.Name.Name,
								Type:       types.BindingStruct,
								ChangeType: changeType,
								StartLine:  fset.Position(ts.Pos()).Line,
								EndLine:    fset.Position(ts.End()).Line,
							}
							if changeType == types.ChangeAdded {
								sym.NewCode = code
							} else {
								sym.OldCode = code
							}
							symbols = append(symbols, sym)
						}
					}
				}
			}
		}
		return true
	})

	return symbols, nil
}

// nodeToString converts an AST node to its string representation.
func (e *SymbolExtractor) nodeToString(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}

// getOldFileContent retrieves the old version of a file from git.
func (e *SymbolExtractor) getOldFileContent(filePath string) (string, error) {
	cmd := exec.Command("git", "show", "HEAD:"+filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// getOldSymbolCode retrieves the old version of a symbol's code.
func (e *SymbolExtractor) getOldSymbolCode(filePath, symbolName string, symbolType types.BindingType) (string, error) {
	oldContent, err := e.getOldFileContent(filePath)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, oldContent, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var result string
	ast.Inspect(file, func(n ast.Node) bool {
		switch symbolType {
		case types.BindingFunc:
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == symbolName {
				result = e.nodeToString(fset, fn)
				return false
			}
		case types.BindingStruct:
			if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == symbolName {
				if _, isStruct := ts.Type.(*ast.StructType); isStruct {
					result = e.nodeToString(fset, ts)
					return false
				}
			}
		}
		return true
	})

	return result, nil
}

// ExtractChangedSymbolsFromBase extracts changed symbols from a base branch.
func ExtractChangedSymbolsFromBase(baseBranch string) ([]types.ChangedSymbol, error) {
	diff, err := GetDiff(baseBranch)
	if err != nil {
		return nil, err
	}

	extractor := NewSymbolExtractor()
	return extractor.ExtractChangedSymbols(diff)
}

// GetChangedGoFiles returns a list of changed Go files from a base branch.
func GetChangedGoFiles(baseBranch string) ([]string, error) {
	diff, err := GetDiff(baseBranch)
	if err != nil {
		return nil, err
	}

	fileDiffs, err := ParseDiff(diff)
	if err != nil {
		return nil, err
	}

	goFiles := FilterGoFiles(fileDiffs)
	var files []string
	for _, fd := range goFiles {
		if fd.ChangeType != types.ChangeDeleted {
			files = append(files, fd.NewPath)
		}
	}

	return files, nil
}
