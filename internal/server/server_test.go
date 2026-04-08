package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupGraphFile(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	graph := `{"generated_at":"2024-01-01T00:00:00Z","agency":"test","nodes":[{"id":"index","label":"Brain","type":"hub","status":"active"}],"edges":[]}`
	if err := os.WriteFile(filepath.Join(root, "graph.json"), []byte(graph), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestServeGraphJSONOK(t *testing.T) {
	root := setupGraphFile(t)
	h := NewGraphHandler(root)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/graph.json", http.NoBody)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	// Verify it's valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["agency"] != "test" {
		t.Errorf("agency = %v, want test", parsed["agency"])
	}
}

func TestServeGraphJSONMissingFile(t *testing.T) {
	root := t.TempDir() // No graph.json here.
	h := NewGraphHandler(root)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/graph.json", http.NoBody)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (fallback), got %d", w.Code)
	}
	// Should return a valid empty graph.
	var parsed map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("fallback should be valid JSON: %v", err)
	}
}

func TestNoCacheHeader(t *testing.T) {
	root := setupGraphFile(t)
	h := NewGraphHandler(root)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/graph.json", http.NoBody)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	cc := w.Header().Get("Cache-Control")
	if cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", cc)
	}
}
