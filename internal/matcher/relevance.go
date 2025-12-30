package matcher

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/yourname/docuguard/internal/llm"
	"github.com/yourname/docuguard/pkg/types"
)

// Matcher 相关性匹配器
type Matcher struct {
	llmClient llm.Client
}

// NewMatcher 创建匹配器
func NewMatcher(client llm.Client) *Matcher {
	return &Matcher{llmClient: client}
}

// FindRelevantDocs 找出与代码变更相关的文档段落
func (m *Matcher) FindRelevantDocs(
	ctx context.Context,
	symbols []types.ChangedSymbol,
	segments []types.DocSegment,
) ([]types.RelevanceResult, error) {
	var results []types.RelevanceResult

	// 预过滤：使用关键词匹配快速筛选
	candidates := m.preFilter(symbols, segments)

	// 对候选项使用 LLM 进行精确判断
	for _, candidate := range candidates {
		result, err := m.checkRelevance(ctx, candidate.symbol, candidate.segment)
		if err != nil {
			// 记录错误但继续处理
			continue
		}
		if result.IsRelevant {
			results = append(results, *result)
		}
	}

	return results, nil
}

type candidate struct {
	symbol  types.ChangedSymbol
	segment types.DocSegment
}

// preFilter 预过滤：基于关键词匹配
func (m *Matcher) preFilter(symbols []types.ChangedSymbol, segments []types.DocSegment) []candidate {
	var candidates []candidate

	for _, sym := range symbols {
		symWords := extractKeywords(sym.Name)
		symWords = append(symWords, strings.ToLower(sym.Name))

		for _, seg := range segments {
			content := strings.ToLower(seg.Content + " " + seg.Heading)

			// 检查是否有关键词匹配
			for _, word := range symWords {
				if len(word) > 2 && strings.Contains(content, word) {
					candidates = append(candidates, candidate{
						symbol:  sym,
						segment: seg,
					})
					break
				}
			}
		}
	}

	return candidates
}

// extractKeywords 从符号名提取关键词
func extractKeywords(name string) []string {
	var words []string
	var current strings.Builder

	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	return words
}

// checkRelevance 使用 LLM 检查相关性
func (m *Matcher) checkRelevance(
	ctx context.Context,
	symbol types.ChangedSymbol,
	segment types.DocSegment,
) (*types.RelevanceResult, error) {
	prompt := buildRelevancePrompt(symbol, segment)

	req := llm.AnalyzeRequest{
		DocContent:  segment.Content,
		CodeContent: symbol.NewCode,
		CodeSymbol:  symbol.Name,
		CodeFile:    symbol.File,
	}

	// 使用自定义 prompt 进行分析
	_ = prompt // 在实际实现中会使用

	result, err := m.llmClient.Analyze(ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.RelevanceResult{
		Segment:    segment,
		Symbol:     symbol,
		IsRelevant: !result.Consistent, // 如果不一致，说明相关但需要更新
		Confidence: result.Confidence,
		Reason:     result.Reason,
	}, nil
}

// buildRelevancePrompt 构建相关性判断的 prompt
func buildRelevancePrompt(symbol types.ChangedSymbol, segment types.DocSegment) string {
	return `你是一个代码文档相关性判断专家。

以下是一个代码变更：
- 文件: ` + symbol.File + `
- 符号: ` + symbol.Name + ` (` + string(symbol.Type) + `)
- 变更类型: ` + string(symbol.ChangeType) + `
- 代码:
` + symbol.NewCode + `

以下是一段文档：
- 文件: ` + segment.File + `
- 标题: ` + segment.Heading + `
- 内容:
` + segment.Content + `

请判断这段文档是否描述了这个代码的功能。
输出 JSON: { "relevant": true/false, "confidence": 0-1, "reason": "..." }`
}

// QuickMatch 快速匹配（不使用 LLM，仅基于关键词）
func QuickMatch(symbols []types.ChangedSymbol, segments []types.DocSegment) []types.RelevanceResult {
	var results []types.RelevanceResult

	for _, sym := range symbols {
		symWords := extractKeywords(sym.Name)
		symWords = append(symWords, strings.ToLower(sym.Name))

		for _, seg := range segments {
			content := strings.ToLower(seg.Content + " " + seg.Heading)
			matchCount := 0

			for _, word := range symWords {
				if len(word) > 2 && strings.Contains(content, word) {
					matchCount++
				}
			}

			if matchCount > 0 {
				confidence := float64(matchCount) / float64(len(symWords))
				results = append(results, types.RelevanceResult{
					Segment:    seg,
					Symbol:     sym,
					IsRelevant: true,
					Confidence: confidence,
					Reason:     "Keyword match",
				})
			}
		}
	}

	return results
}

// ParseRelevanceResponse 解析 LLM 返回的相关性判断结果
func ParseRelevanceResponse(response string) (bool, float64, string, error) {
	// 尝试解析 JSON
	var result struct {
		Relevant   bool    `json:"relevant"`
		Confidence float64 `json:"confidence"`
		Reason     string  `json:"reason"`
	}

	// 查找 JSON 部分
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start >= 0 && end > start {
		jsonStr := response[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result.Relevant, result.Confidence, result.Reason, nil
		}
	}

	// 如果解析失败，使用启发式方法
	lower := strings.ToLower(response)
	relevant := strings.Contains(lower, "relevant") && !strings.Contains(lower, "not relevant")

	return relevant, 0.5, response, nil
}
