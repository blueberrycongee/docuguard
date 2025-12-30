package llm

import "fmt"

const systemPrompt = `You are a code-documentation consistency checker. Your task is to determine whether the given documentation description matches the code implementation.

IMPORTANT: First determine if the documentation is actually describing THIS SPECIFIC code/function.

You must output the result in JSON format with the following fields:
- related: boolean, whether the documentation is specifically about this code/function
- consistent: boolean, whether they are consistent (set to true if not related)
- confidence: number (0-1), confidence level
- reason: string, explanation for the judgment
- suggestion: string, if inconsistent, provide a fix suggestion

Guidelines:
1. FIRST check if the documentation is describing THIS specific function/symbol
2. If the documentation is about a DIFFERENT function or topic, set related=false and consistent=true
3. Only mark as inconsistent when the doc IS about this code but describes it incorrectly
4. Focus on business logic consistency, not code style
5. Values, thresholds, and conditions must match exactly
6. If documented functionality is not implemented in code, mark as inconsistent
7. If code implements extra functionality not in docs, that's acceptable`

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

STEP 1: First determine if this documentation is specifically describing the function/symbol "%s".
STEP 2: If NOT related (doc is about something else), output: {"related": false, "consistent": true, ...}
STEP 3: If related, check if the description matches the actual implementation.

Output the result in JSON format.`, req.DocContent, req.CodeFile, req.CodeSymbol, req.CodeContent, req.CodeSymbol)
}
