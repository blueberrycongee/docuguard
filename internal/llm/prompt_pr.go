package llm

// PRConsistencyPrompt PR 模式下的一致性检查 prompt 模板
const PRConsistencyPrompt = `你是一个代码文档一致性检查专家。

## 代码变更信息
- 文件: {{.CodeFile}}
- 符号: {{.CodeSymbol}} ({{.CodeType}})
- 变更类型: {{.ChangeType}}

### 变更前代码
{{.OldCode}}

### 变更后代码
{{.NewCode}}

## 相关文档
- 文件: {{.DocFile}}
- 标题: {{.DocHeading}}
- 内容:
{{.DocContent}}

## 任务
请分析代码变更后，文档描述是否仍然准确。

重点检查：
1. 数值是否一致（如阈值、限制、默认值等）
2. 行为描述是否准确
3. 参数说明是否正确
4. 返回值描述是否正确

## 输出格式
请输出 JSON:
{
  "consistent": true/false,
  "confidence": 0.0-1.0,
  "reason": "简要说明判断理由",
  "suggestion": "如果不一致，建议如何修改文档"
}`

// PRRelevancePrompt PR 模式下的相关性判断 prompt 模板
const PRRelevancePrompt = `你是一个代码文档相关性判断专家。

## 代码变更
- 文件: {{.CodeFile}}
- 符号: {{.CodeSymbol}} ({{.CodeType}})
- 变更类型: {{.ChangeType}}
- 代码:
{{.NewCode}}

## 文档段落
- 文件: {{.DocFile}}
- 标题: {{.DocHeading}}
- 内容:
{{.DocContent}}

## 任务
请判断这段文档是否描述了这个代码的功能或行为。

判断标准：
1. 文档是否提到了这个函数/结构体/变量
2. 文档是否描述了相关的业务逻辑
3. 文档中的示例是否涉及这段代码

## 输出格式
请输出 JSON:
{
  "relevant": true/false,
  "confidence": 0.0-1.0,
  "reason": "简要说明判断理由"
}`

// PRPromptData PR prompt 数据
type PRPromptData struct {
	// 代码信息
	CodeFile   string
	CodeSymbol string
	CodeType   string
	ChangeType string
	OldCode    string
	NewCode    string

	// 文档信息
	DocFile    string
	DocHeading string
	DocContent string
}
