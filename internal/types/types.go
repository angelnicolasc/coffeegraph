// Package types defines shared interfaces and data contracts used across
// feature packages.
package types

import "context"

// LLMClient executes a completion request.
type LLMClient interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (LLMResult, error)
}

// LLMResult contains a completion response payload.
type LLMResult struct {
	Text         string
	InputTokens  int
	OutputTokens int
}
