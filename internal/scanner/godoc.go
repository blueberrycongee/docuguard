package scanner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/yourname/docuguard/pkg/types"
)

// ScanGoDoc 扫描 Go 文件中的文档注释
func ScanGoDoc(filePath string) ([]types.DocSegment, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var segments []types.DocSegment

	// 包注释
	if file.Doc != nil {
		seg := types.DocSegment{
			File:      filePath,
			StartLine: fset.Position(file.Doc.Pos()).Line,
			EndLine:   fset.Position(file.Doc.End()).Line,
			Heading:   "Package " + file.Name.Name,
			Content:   file.Doc.Text(),
			Type:      "godoc",
			Level:     1,
		}
		segments = append(segments, seg)
	}

	// 遍历声明
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil {
				seg := types.DocSegment{
					File:      filePath,
					StartLine: fset.Position(d.Doc.Pos()).Line,
					EndLine:   fset.Position(d.Doc.End()).Line,
					Heading:   "func " + d.Name.Name,
					Content:   d.Doc.Text(),
					Type:      "godoc",
					Level:     2,
				}
				segments = append(segments, seg)
			}

		case *ast.GenDecl:
			if d.Doc != nil {
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						seg := types.DocSegment{
							File:      filePath,
							StartLine: fset.Position(d.Doc.Pos()).Line,
							EndLine:   fset.Position(d.Doc.End()).Line,
							Heading:   "type " + s.Name.Name,
							Content:   d.Doc.Text(),
							Type:      "godoc",
							Level:     2,
						}
						segments = append(segments, seg)
					}
				}
			}
		}
	}

	return segments, nil
}

// ScanGoDocDir 扫描目录下所有 Go 文件的文档注释
func ScanGoDocDir(dir string) ([]types.DocSegment, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var segments []types.DocSegment
	for _, pkg := range pkgs {
		for filePath, file := range pkg.Files {
			// 跳过测试文件
			if strings.HasSuffix(filePath, "_test.go") {
				continue
			}

			fileSegments, err := extractDocSegments(fset, filePath, file)
			if err != nil {
				continue
			}
			segments = append(segments, fileSegments...)
		}
	}

	return segments, nil
}

func extractDocSegments(fset *token.FileSet, filePath string, file *ast.File) ([]types.DocSegment, error) {
	var segments []types.DocSegment

	// 包注释
	if file.Doc != nil {
		seg := types.DocSegment{
			File:      filePath,
			StartLine: fset.Position(file.Doc.Pos()).Line,
			EndLine:   fset.Position(file.Doc.End()).Line,
			Heading:   "Package " + file.Name.Name,
			Content:   file.Doc.Text(),
			Type:      "godoc",
			Level:     1,
		}
		segments = append(segments, seg)
	}

	// 遍历声明
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil {
				seg := types.DocSegment{
					File:      filePath,
					StartLine: fset.Position(d.Doc.Pos()).Line,
					EndLine:   fset.Position(d.Doc.End()).Line,
					Heading:   "func " + d.Name.Name,
					Content:   d.Doc.Text(),
					Type:      "godoc",
					Level:     2,
				}
				segments = append(segments, seg)
			}

		case *ast.GenDecl:
			if d.Doc != nil {
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						seg := types.DocSegment{
							File:      filePath,
							StartLine: fset.Position(d.Doc.Pos()).Line,
							EndLine:   fset.Position(d.Doc.End()).Line,
							Heading:   "type " + s.Name.Name,
							Content:   d.Doc.Text(),
							Type:      "godoc",
							Level:     2,
						}
						segments = append(segments, seg)
					}
				}
			}
		}
	}

	return segments, nil
}
