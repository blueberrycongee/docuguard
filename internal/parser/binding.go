package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

var (
	// 匹配 docuguard 注释
	bindStartRe = regexp.MustCompile(`<!--\s*docuguard:start\s*-->`)
	bindEndRe   = regexp.MustCompile(`<!--\s*docuguard:end\s*-->`)
	bindCodeRe  = regexp.MustCompile(`<!--\s*docuguard:bindCode\s+path="([^"]+)"\s+(\w+)="([^"]+)"\s*-->`)
)

// ExtractBindings 从 Markdown 文件中提取绑定关系
func ExtractBindings(filePath string) ([]types.Binding, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var bindings []types.Binding
	var currentBinding *types.Binding
	var docContent strings.Builder
	inBlock := false
	lineNum := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 检测块开始
		if bindStartRe.MatchString(line) {
			inBlock = true
			docContent.Reset()
			continue
		}

		// 检测块结束
		if bindEndRe.MatchString(line) {
			if currentBinding != nil {
				currentBinding.DocContent = strings.TrimSpace(docContent.String())
				bindings = append(bindings, *currentBinding)
				currentBinding = nil
			}
			inBlock = false
			continue
		}

		// 在块内检测绑定声明
		if inBlock {
			if matches := bindCodeRe.FindStringSubmatch(line); matches != nil {
				currentBinding = &types.Binding{
					DocFile:    filePath,
					DocLine:    lineNum,
					CodeFile:   matches[1],
					CodeType:   parseBindingType(matches[2]),
					CodeSymbol: matches[3],
				}
			} else if currentBinding != nil {
				// 收集文档内容
				docContent.WriteString(line)
				docContent.WriteString("\n")
			}
		}
	}

	return bindings, scanner.Err()
}

func parseBindingType(s string) types.BindingType {
	switch strings.ToLower(s) {
	case "func":
		return types.BindingFunc
	case "struct":
		return types.BindingStruct
	case "const":
		return types.BindingConst
	case "var":
		return types.BindingVar
	default:
		return types.BindingFunc
	}
}
