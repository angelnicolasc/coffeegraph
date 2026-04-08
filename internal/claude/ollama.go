package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ollamaAPI = "http://localhost:11434/api/chat"

// ErrOllamaUnavailable indicates the local Ollama server cannot be reached.
var ErrOllamaUnavailable = fmt.Errorf("ollama doesn't appear to be running. Start it with: ollama serve")

// OllamaClient calls a local Ollama chat endpoint.
type OllamaClient struct {
	Model      string
	HTTPClient *http.Client
}

type ollamaReq struct {
	Model    string `json:"model"`
	Stream   bool   `json:"stream"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type ollamaResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

// Complete generates a response from Ollama.
func (c *OllamaClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (Result, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "llama3.2"
	}
	reqBody := ollamaReq{Model: model, Stream: false}
	if sp := strings.TrimSpace(systemPrompt); sp != "" {
		reqBody.Messages = append(reqBody.Messages, struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{Role: "system", Content: sp})
	}
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "user", Content: userPrompt})

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("marshal request: %w", err)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaAPI, bytes.NewReader(raw))
	if err != nil {
		return Result{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return Result{}, ErrOllamaUnavailable
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return Result{}, fmt.Errorf("ollama api %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var parsed ollamaResp
	if err := json.Unmarshal(b, &parsed); err != nil {
		return Result{}, fmt.Errorf("invalid response: %w", err)
	}
	return Result{
		Text:         strings.TrimSpace(parsed.Message.Content),
		InputTokens:  parsed.PromptEvalCount,
		OutputTokens: parsed.EvalCount,
	}, nil
}
