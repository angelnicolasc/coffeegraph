package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/queue"
)

func setupProject(t *testing.T) (string, *config.Config) {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{
		filepath.Join(root, "skills", "test-skill"),
		filepath.Join(root, ".coffee", "logs"),
		filepath.Join(root, ".coffee", "snapshots"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	// Write a valid SKILL.md.
	skillMD := "# Skill: test-skill\n\n## Identity\nYou are a test agent.\n\n## Workflow\n1. Do stuff\n\n## Output Format\nPlain text\n"
	if err := os.WriteFile(filepath.Join(root, "skills", "test-skill", "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Write index.md.
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte("# Test Agency\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Write config.
	cfgYAML := `agency_name: test
default_model: test-model
backend: anthropic
api_keys:
  anthropic: "sk-ant-test"
skills:
  test-skill:
    enabled: true
`
	cfgPath := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	// Write empty queue.
	if err := queue.Set(root, nil); err != nil {
		t.Fatalf("setup queue: %v", err)
	}
	// Write graph.json placeholder.
	if err := os.WriteFile(filepath.Join(root, "graph.json"), []byte(`{"generated_at":"2024-01-01T00:00:00Z","agency":"test","nodes":[{"id":"index","label":"Brain","type":"hub","status":"active"}],"edges":[]}`), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return root, cfg
}

func TestExecutePendingEmptyQueue(t *testing.T) {
	root, cfg := setupProject(t)
	engine := &Engine{Root: root, Cfg: cfg}
	summary, err := engine.ExecutePending(context.Background(), 10, ModeNormal)
	if err != nil {
		t.Fatalf("ExecutePending() error = %v", err)
	}
	if summary.Completed != 0 {
		t.Fatalf("expected 0 completed, got %d", summary.Completed)
	}
}

func TestExecutePendingCancelledContext(t *testing.T) {
	root, cfg := setupProject(t)
	// Add a task so it tries to execute.
	_, _, _ = queue.Add(root, queue.Item{Skill: "test-skill", Task: "test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	engine := &Engine{Root: root, Cfg: cfg}
	summary, err := engine.ExecutePending(ctx, 10, ModeNormal)
	if err != nil {
		t.Fatalf("ExecutePending() error = %v", err)
	}
	// Should have recorded context cancellation.
	if len(summary.Errors) == 0 {
		t.Fatal("expected errors for cancelled context")
	}
}

func TestExecuteTaskSkillDisabled(t *testing.T) {
	root, cfg := setupProject(t)
	// Mark skill as disabled by modifying the in-memory config.
	cfg.Skills["test-skill"] = config.SkillEntry{Enabled: false}

	engine := &Engine{Root: root, Cfg: cfg}
	it := queue.Item{Skill: "test-skill", Task: "test"}
	_, err := engine.ExecuteTask(context.Background(), it)
	if err == nil {
		t.Fatal("expected error for disabled skill")
	}
	if err.Error() != "skill test-skill is disabled in config" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteTaskInvalidSkillMD(t *testing.T) {
	root, cfg := setupProject(t)
	// Overwrite SKILL.md with invalid content.
	if err := os.WriteFile(filepath.Join(root, "skills", "test-skill", "SKILL.md"), []byte("just random text"), 0644); err != nil {
		t.Fatalf("write invalid skill: %v", err)
	}
	engine := &Engine{Root: root, Cfg: cfg}
	it := queue.Item{Skill: "test-skill", Task: "test"}
	_, err := engine.ExecuteTask(context.Background(), it)
	if err == nil {
		t.Fatal("expected error for invalid SKILL.md")
	}
}

func TestUrgentMsg(t *testing.T) {
	tests := []struct {
		idx  int
		want string
	}{
		{0, "NO SLACKING. CLOSE THOSE DEALS."},
		{1, "YOUR COMPETITION IS ALREADY RUNNING."},
		{2, "WHY IS THIS TAKING SO LONG."},
		{3, "NO SLACKING. CLOSE THOSE DEALS."},
	}
	for _, tt := range tests {
		got := urgentMsg(tt.idx)
		if got != tt.want {
			t.Errorf("urgentMsg(%d) = %q, want %q", tt.idx, got, tt.want)
		}
	}
}
