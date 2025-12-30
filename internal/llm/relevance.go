package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

// RelevanceRequest represents a batch relevance check request.
type RelevanceRequest struct {
	Symbol     types.ChangedSymbol
	Candidates []types.DocSegment
}

// RelevanceResponse represents the LLM response for relevance check.
type RelevanceResponse struct {
	RelevantIndices []int `json:"relevant"`
}

const relevanceSystemPrompt = `You are a code-documentation relevance expert. Your task is to determine which documentation segments are specifically describing a given code symbol.

You must output the result in JSON format:
{"relevant": [indices of relevant segments]}

Guidelines:
1. A segment is relevant if it specifically describes THIS function/struct/variable
2. A segment is relevant if it contains usage examples of THIS symbol
3. A segment is NOT relevant if it only mentions the symbol name in passing
4. A segment is NOT relevant if it describes a different but similarly named symbol
5. When in doubt, include the segment (prefer false positives over false negatives)`

func buildRelevancePrompt(req RelevanceRequest) string {
	var sb strings.Builder

	sb.WriteString("## Code Symbol\n")
	sb.WriteString("Name: " + req.Symbol.Name + "\n")
	sb.WriteString("Type: " + string(req.Symbol.Type) + "\n")
	sb.WriteString("File: " + req.Symbol.File + "\n\n")
	sb.WriteString("```go\n")
	sb.WriteString(req.Symbol.NewCode)
	sb.WriteString("\n```\n\n")

	sb.WriteString("## Candidate Documentation Segments\n\n")
	for i, seg := range req.Candidates {
		sb.WriteString(fmt.Sprintf("[%d] %s - %s\n", i, seg.File, seg.Heading))
		// Truncate content if too long
		content := seg.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Which segments (by index) are specifically describing this code symbol?\n")
	sb.WriteString("Output JSON: {\"relevant\": [list of indices]}")

	return sb.String()
}

// CheckRelevanceBatch checks relevance of multiple document segments for a symbol in one LLM call.
func (c *OpenAIClient) CheckRelevanceBatch(ctx context.Context, req RelevanceRequest) ([]int, error) {
	if len(req.Candidates) == 0 {
		return nil, nil
	}

	prompt := buildRelevancePrompt(req)

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": relevanceSystemPrompt},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&response).
		Post("/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	var result RelevanceResponse
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate indices
	validIndices := make([]int, 0)
	for _, idx := range result.RelevantIndices {
		if idx >= 0 && idx < len(req.Candidates) {
			validIndices = append(validIndices, idx)
		}
	}

	return validIndices, nil
}
