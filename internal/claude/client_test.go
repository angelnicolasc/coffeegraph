package claude

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCompleteMissingAPIKey(t *testing.T) {
	c := &Client{APIKey: ""}
	_, err := c.Complete(context.Background(), "sys", "user")
	if err == nil {
		t.Fatal("expected ErrMissingAPIKey")
	}
	if err != ErrMissingAPIKey {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompleteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"hello world"}],"usage":{"input_tokens":10,"output_tokens":5}}`))
	}))
	defer srv.Close()

	c := &Client{
		APIKey:     "sk-test",
		HTTPClient: srv.Client(),
	}
	// Override the API URL by using a custom transport.
	c.HTTPClient = &http.Client{
		Transport: &rewriteTransport{base: srv.Client().Transport, target: srv.URL},
	}

	result, err := c.Complete(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if result.Text != "hello world" {
		t.Fatalf("Text = %q, want 'hello world'", result.Text)
	}
	if result.InputTokens != 10 || result.OutputTokens != 5 {
		t.Fatalf("tokens: in=%d out=%d", result.InputTokens, result.OutputTokens)
	}
}

func TestCompleteNonRetryableError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := &Client{
		APIKey: "sk-test",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{base: srv.Client().Transport, target: srv.URL},
		},
	}

	_, err := c.Complete(context.Background(), "", "test")
	if err == nil {
		t.Fatal("expected error for 401")
	}
	// 401 is non-retryable, should only call once.
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("expected 1 call for non-retryable, got %d", n)
	}
}

func TestCompleteRetriesOnServerError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
			return
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"recovered"}],"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer srv.Close()

	c := &Client{
		APIKey: "sk-test",
		HTTPClient: &http.Client{
			Transport: &rewriteTransport{base: srv.Client().Transport, target: srv.URL},
		},
	}

	result, err := c.Complete(context.Background(), "", "test")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if result.Text != "recovered" {
		t.Fatalf("Text = %q, want recovered", result.Text)
	}
}

func TestCompleteContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := &Client{APIKey: "sk-test"}
	_, err := c.Complete(ctx, "", "test")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestBackoffHasJitter(t *testing.T) {
	// Verify that jitteredBackoff returns values in the expected range.
	base := 2 * time.Second
	seen := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		d := jitteredBackoff(base)
		seen[d] = true
		// With ±25% jitter, 2s should be between 1.5s and 2.5s.
		if d < base*3/4 || d > base*5/4 {
			t.Fatalf("jitter out of range: %v (base=%v)", d, base)
		}
	}
	// With 20 samples, we should see at least 2 distinct values.
	if len(seen) < 2 {
		t.Fatal("jitter appears to produce no variation")
	}
}

func TestAPIErrorMethods(t *testing.T) {
	e429 := &APIError{StatusCode: 429, Status: "429 Too Many Requests", Body: "rate limited"}
	if !e429.IsRateLimited() {
		t.Fatal("429 should be rate limited")
	}
	if !e429.IsRetryable() {
		t.Fatal("429 should be retryable")
	}

	e500 := &APIError{StatusCode: 500, Status: "500 Internal Server Error", Body: ""}
	if e500.IsRateLimited() {
		t.Fatal("500 should not be rate limited")
	}
	if !e500.IsRetryable() {
		t.Fatal("500 should be retryable")
	}

	e400 := &APIError{StatusCode: 400, Status: "400 Bad Request", Body: "bad"}
	if e400.IsRetryable() {
		t.Fatal("400 should not be retryable")
	}
}

// rewriteTransport redirects all requests to the test server.
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.target, "http://")
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
