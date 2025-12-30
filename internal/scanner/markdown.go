package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yourname/docuguard/pkg/types"
)

var (
	// 匹配 Markdown 标题
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	// 匹配代码块开始/结束
	codeBlockRegex = regexp.MustCompile("^```")
)

// ScanMarkdown 扫描单个 Markdown 文件，按标题分段
func ScanMarkdown(filePath string) ([]types.DocSegment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var segments []types.DocSegment
	var currentSegment *types.DocSegment
	var contentBuilder strings.Builder
	inCodeBlock := false
	lineNum := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 检查代码块
		if codeBlockRegex.MatchString(line) {
			inCodeBlock = !inCodeBlock
			if currentSegment != nil {
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}
			continue
		}

		// 在代码块内，直接添加内容
		if inCodeBlock {
			if currentSegment != nil {
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}
			continue
		}

		// 检查是否是标题
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			// 保存之前的段落
			if currentSegment != nil {
				currentSegment.Content = strings.TrimSpace(contentBuilder.String())
				currentSegment.EndLine = lineNum - 1
				if currentSegment.Content != "" {
					segments = append(segments, *currentSegment)
				}
			}

			// 开始新段落
			level := len(matches[1])
			heading := strings.TrimSpace(matches[2])
			currentSegment = &types.DocSegment{
				File:      filePath,
				StartLine: lineNum,
				Heading:   heading,
				Type:      "markdown",
				Level:     level,
			}
			contentBuilder.Reset()
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
			continue
		}

		// 普通内容
		if currentSegment != nil {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	// 保存最后一个段落
	if currentSegment != nil {
		currentSegment.Content = strings.TrimSpace(contentBuilder.String())
		currentSegment.EndLine = lineNum
		if currentSegment.Content != "" {
			segments = append(segments, *currentSegment)
		}
	}

	return segments, scanner.Err()
}

// ScanMarkdownDir 扫描目录下所有匹配的 Markdown 文件
func ScanMarkdownDir(rootDir string, patterns []string) ([]types.DocSegment, error) {
	var allSegments []types.DocSegment

	for _, pattern := range patterns {
		// 处理 glob 模式
		matches, err := filepath.Glob(filepath.Join(rootDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}

			if strings.HasSuffix(strings.ToLower(match), ".md") {
				segments, err := ScanMarkdown(match)
				if err != nil {
					continue
				}
				allSegments = append(allSegments, segments...)
			}
		}
	}

	// 如果 glob 没有匹配到，尝试直接扫描文件
	for _, pattern := range patterns {
		if !strings.Contains(pattern, "*") {
			filePath := filepath.Join(rootDir, pattern)
			if _, err := os.Stat(filePath); err == nil {
				segments, err := ScanMarkdown(filePath)
				if err == nil {
					// 检查是否已添加
					exists := false
					for _, s := range allSegments {
						if s.File == filePath {
							exists = true
							break
						}
					}
					if !exists {
						allSegments = append(allSegments, segments...)
					}
				}
			}
		}
	}

	return allSegments, nil
}

// FilterRelevantSegments 过滤出可能相关的段落（基于关键词预过滤）
func FilterRelevantSegments(segments []types.DocSegment, symbols []types.ChangedSymbol) []types.DocSegment {
	var relevant []types.DocSegment

	for _, seg := range segments {
		content := strings.ToLower(seg.Content + " " + seg.Heading)
		for _, sym := range symbols {
			// 检查符号名是否出现在文档中
			symLower := strings.ToLower(sym.Name)
			if strings.Contains(content, symLower) {
				relevant = append(relevant, seg)
				break
			}

			// 检查驼峰命名拆分后的词
			words := splitCamelCase(sym.Name)
			for _, word := range words {
				if len(word) > 3 && strings.Contains(content, strings.ToLower(word)) {
					relevant = append(relevant, seg)
					break
				}
			}
		}
	}

	return relevant
}

// splitCamelCase 拆分驼峰命名
func splitCamelCase(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}
