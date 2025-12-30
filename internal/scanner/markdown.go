package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

var (
	// headingRegex matches Markdown headings.
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	// codeBlockRegex matches code block delimiters.
	codeBlockRegex = regexp.MustCompile("^```")
)

// ScanMarkdown scans a Markdown file and returns document segments.
// Each segment corresponds to a heading and its content.
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

		if codeBlockRegex.MatchString(line) {
			inCodeBlock = !inCodeBlock
			if currentSegment != nil {
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}
			continue
		}

		if inCodeBlock {
			if currentSegment != nil {
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}
			continue
		}

		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			if currentSegment != nil {
				currentSegment.Content = strings.TrimSpace(contentBuilder.String())
				currentSegment.EndLine = lineNum - 1
				if currentSegment.Content != "" {
					segments = append(segments, *currentSegment)
				}
			}

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

		if currentSegment != nil {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	if currentSegment != nil {
		currentSegment.Content = strings.TrimSpace(contentBuilder.String())
		currentSegment.EndLine = lineNum
		if currentSegment.Content != "" {
			segments = append(segments, *currentSegment)
		}
	}

	return segments, scanner.Err()
}

// ScanMarkdownDir scans a directory for Markdown files matching the patterns.
func ScanMarkdownDir(rootDir string, patterns []string) ([]types.DocSegment, error) {
	var allSegments []types.DocSegment

	for _, pattern := range patterns {
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

	// Handle non-glob patterns (direct file paths).
	for _, pattern := range patterns {
		if !strings.Contains(pattern, "*") {
			filePath := filepath.Join(rootDir, pattern)
			if _, err := os.Stat(filePath); err == nil {
				segments, err := ScanMarkdown(filePath)
				if err == nil {
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

// FilterRelevantSegments filters segments that may be relevant to the symbols.
// This is a pre-filter based on keyword matching.
func FilterRelevantSegments(segments []types.DocSegment, symbols []types.ChangedSymbol) []types.DocSegment {
	var relevant []types.DocSegment

	for _, seg := range segments {
		content := strings.ToLower(seg.Content + " " + seg.Heading)
		for _, sym := range symbols {
			symLower := strings.ToLower(sym.Name)
			if strings.Contains(content, symLower) {
				relevant = append(relevant, seg)
				break
			}

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

// splitCamelCase splits a camelCase string into words.
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
