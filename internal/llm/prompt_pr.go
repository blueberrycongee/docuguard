package llm

// PRConsistencyPrompt is the prompt template for PR consistency checking.
const PRConsistencyPrompt = `You are a code documentation consistency expert.

## Code Change Information
- File: {{.CodeFile}}
- Symbol: {{.CodeSymbol}} ({{.CodeType}})
- Change Type: {{.ChangeType}}

### Code Before Change
{{.OldCode}}

### Code After Change
{{.NewCode}}

## Related Documentation
- File: {{.DocFile}}
- Heading: {{.DocHeading}}
- Content:
{{.DocContent}}

## Task
Analyze whether the documentation is still accurate after the code change.

Focus on:
1. Numeric values (thresholds, limits, defaults)
2. Behavior descriptions
3. Parameter documentation
4. Return value documentation

## Output Format
Output JSON:
{
  "consistent": true/false,
  "confidence": 0.0-1.0,
  "reason": "Brief explanation",
  "suggestion": "If inconsistent, how to fix the documentation"
}`

// PRRelevancePrompt is the prompt template for relevance determination.
const PRRelevancePrompt = `You are a code documentation relevance expert.

## Code Change
- File: {{.CodeFile}}
- Symbol: {{.CodeSymbol}} ({{.CodeType}})
- Change Type: {{.ChangeType}}
- Code:
{{.NewCode}}

## Documentation Segment
- File: {{.DocFile}}
- Heading: {{.DocHeading}}
- Content:
{{.DocContent}}

## Task
Determine if this documentation describes the functionality of this code.

Criteria:
1. Does the documentation mention this function/struct/variable?
2. Does the documentation describe related business logic?
3. Do examples in the documentation involve this code?

## Output Format
Output JSON:
{
  "relevant": true/false,
  "confidence": 0.0-1.0,
  "reason": "Brief explanation"
}`

// PRPromptData contains data for PR prompt templates.
type PRPromptData struct {
	CodeFile   string
	CodeSymbol string
	CodeType   string
	ChangeType string
	OldCode    string
	NewCode    string
	DocFile    string
	DocHeading string
	DocContent string
}
