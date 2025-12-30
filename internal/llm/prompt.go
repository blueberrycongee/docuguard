package llm

import "fmt"

const systemPrompt = `You are a code-documentation consistency checker. Your task is to determine whether the given documentation description matches the code implementation.

You must output the result in JSON format with the following fields:
- consistent: boolean, whether they are consistent
- confidence: number (0-1), confidence level
- reason: string, explanation for the judgment
- suggestion: string, if inconsistent, provide a fix suggestion

Guidelines:
1. Focus on business logic consistency, not code style
2. Values, thresholds, and conditions must match exactly
3. If documented functionality is not implemented in code, mark as inconsistent
4. If code implements extra functionality not in docs, that's acceptable`

func buildPrompt(req AnalyzeRequest) string {
	return fmt.Sprintf(`Please check if the following documentation matches the code implementation:

## Documentation
%s

## Code Implementation
File: %s
Symbol: %s

'''go
%s
'''

Analyze and output the result in JSON format.`, req.DocContent, req.CodeFile, req.CodeSymbol, req.CodeContent)
}
