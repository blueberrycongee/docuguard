package parser

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// GoParser Go 代码解析器
type GoParser struct {
	fset *token.FileSet
}

// NewGoParser 创建 Go 解析器
func NewGoParser() *GoParser {
	return &GoParser{
		fset: token.NewFileSet(),
	}
}

// ExtractSymbol 提取指定符号的代码
func (p *GoParser) ExtractSymbol(filePath string, symbolName string, symbolType types.BindingType) (string, int, error) {
	// 解析文件
	file, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", 0, err
	}

	var targetNode ast.Node
	var lineNum int

	// 遍历 AST 查找目标符号
	ast.Inspect(file, func(n ast.Node) bool {
		switch symbolType {
		case types.BindingFunc:
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == symbolName {
				targetNode = fn
				lineNum = p.fset.Position(fn.Pos()).Line
				return false
			}
		case types.BindingStruct:
			if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == symbolName {
				if _, isStruct := ts.Type.(*ast.StructType); isStruct {
					targetNode = ts
					lineNum = p.fset.Position(ts.Pos()).Line
					return false
				}
			}
		case types.BindingConst, types.BindingVar:
			if vs, ok := n.(*ast.ValueSpec); ok {
				for _, name := range vs.Names {
					if name.Name == symbolName {
						targetNode = vs
						lineNum = p.fset.Position(vs.Pos()).Line
						return false
					}
				}
			}
		}
		return true
	})

	if targetNode == nil {
		return "", 0, nil
	}

	// 格式化输出
	var buf bytes.Buffer
	if err := format.Node(&buf, p.fset, targetNode); err != nil {
		return "", 0, err
	}

	return buf.String(), lineNum, nil
}
