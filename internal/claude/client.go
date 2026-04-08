// Package claude provides a thin client for the Anthropic Messages API.
//
// It is intentionally minimal: no streaming, no tool-use — just
// system + user → assistant text. Designed for CoffeeGraph's
// fire-and-forget skill execution model.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	anthropicAPI     = "https://api.anthropic.com/v1/messages"
	defaultModel     = "claude-sonnet-4-20250514"
	defaultVersion   = "2023-06-01"
	defaultMaxTokens = 8192
	defaultTimeout   = 90 * time.Second
	maxRetries       = 3
)

// ErrMissingAPIKey is returned when no API key is configured.
var ErrMissingAPIKey = errors.New(
	"ANTHROPIC_API_KEY not configured.\nAdd it to config.yaml or set it as an environment variable",
)

// APIError represents a non-2xx response from the Anthropic API.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
	RetryAfter time.Duration // parsed from Retry-After header, zero if absent
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Anthropic API %s: %s", e.Status, e.Body)
}

// IsRateLimited returns true if this is a 429 Too Many Requests error.
func (e *APIError) IsRateLimited() bool { return e.StatusCode == 429 }

// IsRetryable returns true for transient server errors (5xx) or rate limits.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == 429 || e.StatusCode >= 500
}

// Client calls the Anthropic Messages API (synchronous, no streaming).
type Client struct {
	APIKey     string
	Model      string
	MaxTokens  int
	HTTPClient *http.Client
	Version    string
}

type apiBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []msgPart `json:"messages"`
}

type msgPart struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messageResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Result contains the API response text and token usage.
type Result struct {
	Text         string
	InputTokens  int
	OutputTokens int
}

// Complete sends a system prompt and user message to Claude and returns
// the concatenated text response. It retries up to 3 times on transient
// errors with exponential backoff, and respects the Retry-After header.
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (Result, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return Result{}, ErrMissingAPIKey
	}

	model := c.Model
	if model == "" {
		model = defaultModel
	}
	ver := c.Version
	if ver == "" {
		ver = defaultVersion
	}
	maxTok := c.MaxTokens
	if maxTok <= 0 {
		maxTok = defaultMaxTokens
	}

	body := apiBody{
		Model:     model,
		MaxTokens: maxTok,
		Messages:  []msgPart{{Role: "user", Content: userPrompt}},
	}
	if s := strings.TrimSpace(systemPrompt); s != "" {
		body.System = s
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return Result{}, fmt.Errorf("marshal request: %w", err)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			base := time.Duration(math.Pow(2, float64(attempt))) * time.Second

			// Use Retry-After from the previous error when it exceeds
			// the exponential backoff (common with Anthropic rate limits).
			var apiErr *APIError
			if errors.As(lastErr, &apiErr) && apiErr.RetryAfter > 0 && apiErr.RetryAfter > base {
				base = apiErr.RetryAfter
			}

			backoff := jitteredBackoff(base)
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := c.doRequest(ctx, httpClient, raw, ver)
		if err == nil {
			return result, nil
		}
		lastErr = err

		var apiErr *APIError
		if errors.As(err, &apiErr) && !apiErr.IsRetryable() {
			return Result{}, err // non-retryable API error
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return Result{}, err // caller cancelled
		}
	}
	return Result{}, fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) doRequest(ctx context.Context, client *http.Client, body []byte, ver string) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPI, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", ver)
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(b),
		}
		// Parse Retry-After into the error so the retry loop can use it
		// with context-aware sleep instead of blocking here.
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, parseErr := strconv.Atoi(ra); parseErr == nil && secs > 0 {
				apiErr.RetryAfter = time.Duration(secs) * time.Second
			}
		}
		return Result{}, apiErr
	}

	var mr messageResponse
	if err := json.Unmarshal(b, &mr); err != nil {
		return Result{}, fmt.Errorf("invalid response: %w", err)
	}

	var out strings.Builder
	for _, block := range mr.Content {
		if block.Text != "" {
			out.WriteString(block.Text)
		}
	}
	return Result{
		Text:         strings.TrimSpace(out.String()),
		InputTokens:  mr.Usage.InputTokens,
		OutputTokens: mr.Usage.OutputTokens,
	}, nil
}
