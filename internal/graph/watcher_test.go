package graph

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/config"
)

func TestWatchSkillsDebounce(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping file watcher test in CI; requires filesystem events")
	}

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills", "test-skill")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# Skill: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write graph.json so Generate doesn't fail.
	if err := os.WriteFile(filepath.Join(root, "graph.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	var callCount int64
	cfg := &config.Config{AgencyName: "test", DefaultModel: "test"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop, err := WatchSkills(ctx, root,
		func() (*config.Config, error) { return cfg, nil },
		func() map[string]int { return map[string]int{} },
		func() { atomic.AddInt64(&callCount, 1) },
	)
	if err != nil {
		t.Fatalf("WatchSkills() error = %v", err)
	}
	defer stop()

	// Rapid-fire writes (should be debounced to ~1 call).
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(
			filepath.Join(skillsDir, "SKILL.md"),
			[]byte("# Skill: test\n## updated\n"),
			0o644,
		)
		time.Sleep(30 * time.Millisecond)
	}

	// Wait for debounce window + execution.
	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt64(&callCount)
	if count == 0 {
		t.Fatal("expected at least 1 regeneration call")
	}
	if count > 3 {
		t.Errorf("expected debounce to coalesce calls, got %d calls", count)
	}
}

func TestWatchSkillsCancel(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping file watcher test in CI")
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := &config.Config{AgencyName: "test", DefaultModel: "test"}

	stop, err := WatchSkills(ctx, root,
		func() (*config.Config, error) { return cfg, nil },
		func() map[string]int { return map[string]int{} },
		nil,
	)
	if err != nil {
		t.Fatalf("WatchSkills() error = %v", err)
	}

	// Should not panic.
	cancel()
	stop()
}
