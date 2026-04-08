package queue

import (
	"os"
	"path/filepath"
	"testing"
)

func setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".coffee"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return root
}

func TestReadEmpty(t *testing.T) {
	root := setupProject(t)
	items, err := Read(root)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestWriteAndRead(t *testing.T) {
	root := setupProject(t)
	want := []Item{
		{ID: "a", Skill: "sales-closer", Task: "follow up", Priority: 1},
		{ID: "b", Skill: "content-engine", Task: "write thread", Priority: 3},
	}
	if err := Write(root, want); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := Read(root)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	// Items should be sorted by priority.
	if got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestWriteNilSlice(t *testing.T) {
	root := setupProject(t)
	if err := Write(root, nil); err != nil {
		t.Fatalf("Write(nil) error = %v", err)
	}
	got, err := Read(root)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 items after nil write, got %d", len(got))
	}
}

func TestAddGeneratesID(t *testing.T) {
	root := setupProject(t)
	pos, total, err := Add(root, Item{Skill: "sales-closer", Task: "test"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if pos != 1 || total != 1 {
		t.Fatalf("position=%d total=%d, want 1,1", pos, total)
	}
	items, _ := Read(root)
	if items[0].ID == "" {
		t.Fatal("expected auto-generated ID, got empty")
	}
}

func TestAddDefaultPriority(t *testing.T) {
	root := setupProject(t)
	_, _, err := Add(root, Item{Skill: "test", Task: "t", Priority: 0})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	items, _ := Read(root)
	if items[0].Priority != 3 {
		t.Fatalf("expected default priority 3, got %d", items[0].Priority)
	}
}

func TestAddOutOfRangePriority(t *testing.T) {
	root := setupProject(t)
	_, _, err := Add(root, Item{Skill: "test", Task: "t", Priority: 10})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	items, _ := Read(root)
	if items[0].Priority != 3 {
		t.Fatalf("expected clamped priority 3, got %d", items[0].Priority)
	}
}

func TestRemove(t *testing.T) {
	root := setupProject(t)
	_ = Write(root, []Item{
		{ID: "keep", Skill: "a", Task: "t1", Priority: 1},
		{ID: "drop", Skill: "b", Task: "t2", Priority: 2},
	})
	if err := Remove(root, "drop"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	items, _ := Read(root)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "keep" {
		t.Fatalf("wrong item kept: %v", items[0])
	}
}

func TestRemoveNonExistent(t *testing.T) {
	root := setupProject(t)
	_ = Write(root, []Item{{ID: "a", Skill: "x", Task: "t", Priority: 1}})
	if err := Remove(root, "nonexistent"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	items, _ := Read(root)
	if len(items) != 1 {
		t.Fatal("should not remove anything for unknown ID")
	}
}

func TestCountBySkill(t *testing.T) {
	root := setupProject(t)
	_ = Write(root, []Item{
		{ID: "1", Skill: "sales-closer", Task: "a", Priority: 1},
		{ID: "2", Skill: "sales-closer", Task: "b", Priority: 2},
		{ID: "3", Skill: "content-engine", Task: "c", Priority: 1},
	})
	counts, err := CountBySkill(root)
	if err != nil {
		t.Fatalf("CountBySkill() error = %v", err)
	}
	if counts["sales-closer"] != 2 {
		t.Fatalf("sales-closer count = %d, want 2", counts["sales-closer"])
	}
	if counts["content-engine"] != 1 {
		t.Fatalf("content-engine count = %d, want 1", counts["content-engine"])
	}
}

func TestPriorityOrdering(t *testing.T) {
	root := setupProject(t)
	_ = Write(root, []Item{
		{ID: "low", Skill: "a", Task: "t", Priority: 5},
		{ID: "high", Skill: "a", Task: "t", Priority: 1},
		{ID: "mid", Skill: "a", Task: "t", Priority: 3},
	})
	items, _ := Read(root)
	if items[0].ID != "high" || items[1].ID != "mid" || items[2].ID != "low" {
		t.Fatalf("unexpected priority order: %v", items)
	}
}

func TestStableOrderingSamePriority(t *testing.T) {
	root := setupProject(t)
	_ = Write(root, []Item{
		{ID: "b", Skill: "x", Task: "t", Priority: 3},
		{ID: "a", Skill: "x", Task: "t", Priority: 3},
		{ID: "c", Skill: "x", Task: "t", Priority: 3},
	})
	items, _ := Read(root)
	// Same priority should sort by ID alphabetically.
	if items[0].ID != "a" || items[1].ID != "b" || items[2].ID != "c" {
		t.Fatalf("expected stable sort by ID, got: %v", items)
	}
}
