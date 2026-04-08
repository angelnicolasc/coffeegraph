package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	root := t.TempDir()
	start := time.Now().UTC().Add(-time.Second)
	finish := time.Now().UTC()
	p, err := Write(root, Entry{
		ID:         "abc123",
		Skill:      "sales-closer",
		Task:       "follow up",
		Status:     "DONE",
		StartedAt:  start,
		FinishedAt: finish,
		Result:     "ok",
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got.ID != "abc123" || got.Skill != "sales-closer" || got.Result != "ok" {
		t.Fatalf("unexpected parsed entry: %+v", got)
	}
}

func TestReadFileFallbackPlainText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plain.md")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got.Result != "hello" {
		t.Fatalf("result = %q, want hello", got.Result)
	}
}

func TestFindByIDErrors(t *testing.T) {
	root := t.TempDir()
	if _, err := FindByID(root, "missing"); err == nil {
		t.Fatalf("FindByID() expected error")
	}
}
