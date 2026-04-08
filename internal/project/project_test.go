package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRootFound(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("agency_name: test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a nested directory to search from.
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := FindRoot(nested)
	if err != nil {
		t.Fatalf("FindRoot() error = %v", err)
	}
	if got != filepath.Clean(root) {
		t.Fatalf("FindRoot() = %q, want %q", got, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindRoot(dir)
	if err == nil {
		t.Fatal("expected error when config.yaml not found")
	}
}

func TestFindRootExactDir(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("agency_name: x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := FindRoot(root)
	if err != nil {
		t.Fatalf("FindRoot() error = %v", err)
	}
	if got != filepath.Clean(root) {
		t.Fatalf("FindRoot() = %q, want %q", got, root)
	}
}
