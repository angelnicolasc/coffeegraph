package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/graph"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/gen2brain/beeep"
)

// RunCoffee executes coffee mode: processes tasks from the queue using
// Claude, saves outputs, and optionally sends a desktop notification.
func RunCoffee() error {
	ctx := context.Background()

	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}

	items, err := queue.Read(root)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("☕ Queue is empty. Add tasks with: coffeegraph queue add")
		return nil
	}

	apiKey := cfg.AnthropicKey()
	if strings.TrimSpace(apiKey) == "" {
		return claude.ErrMissingAPIKey
	}

	max := cfg.Coffee.MaxTasksPerRun
	if max <= 0 {
		max = 3
	}
	toRun := items
	if len(toRun) > max {
		toRun = toRun[:max]
	}

	fmt.Printf("☕ Coffee mode — executing %d task(s)...\n\n", len(toRun))

	client := &claude.Client{APIKey: apiKey}
	completed := 0
	var errs []string

	for i, it := range toRun {
		client.Model = cfg.ModelForSkill(it.Skill)

		skillPath := filepath.Join(root, "skills", it.Skill, "SKILL.md")
		skillBody, err := os.ReadFile(skillPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s] skill not found: %v", it.Skill, err))
			continue
		}
		indexBody, err := os.ReadFile(filepath.Join(root, "index.md"))
		if err != nil {
			return fmt.Errorf("read index.md: %w", err)
		}

		// Snapshot SKILL.md before execution.
		snapshotDir := filepath.Join(root, ".coffee", "snapshots")
		_ = os.MkdirAll(snapshotDir, 0755)
		snapName := fmt.Sprintf("%s-%s.md", time.Now().Format("2006-01-02-15-04"), it.Skill)
		_ = os.WriteFile(filepath.Join(snapshotDir, snapName), skillBody, 0644)

		// Mark as running in graph.
		now := time.Now().UTC()
		_ = graph.UpdateSkill(root, it.Skill, func(n *graph.Node) {
			n.Status = graph.StatusRunning
			n.LastRun = &now
		})
		_ = syncGraph(root, cfg)

		fmt.Printf("  [%d/%d] %s — %s\n", i+1, len(toRun), it.Skill, truncate(it.Task, 60))

		// Call Claude API.
		userPrompt := fmt.Sprintf("CONTEXT:\n%s\n\nTASK:\n%s\n\nDATA:\n%s\n", string(indexBody), it.Task, it.Data)
		result, err := client.Complete(ctx, string(skillBody), userPrompt)
		if err != nil {
			_ = graph.UpdateSkill(root, it.Skill, func(n *graph.Node) {
				n.Status = graph.StatusError
				n.LastOutputPreview = err.Error()
			})
			_ = syncGraph(root, cfg)
			errs = append(errs, fmt.Sprintf("[%s] %v", it.Skill, err))
			continue // don't abort — process remaining tasks
		}

		// Save output log.
		ts := time.Now().Format("2006-01-02-15-04")
		logName := fmt.Sprintf("%s-%s.md", ts, it.Skill)
		logDir := filepath.Join(root, ".coffee", "logs")
		_ = os.MkdirAll(logDir, 0755)
		_ = os.WriteFile(filepath.Join(logDir, logName), []byte(result.Text), 0644)

		// Update graph with result preview.
		preview := truncate(strings.TrimSpace(result.Text), 120)
		t2 := time.Now().UTC()
		_ = graph.UpdateSkill(root, it.Skill, func(n *graph.Node) {
			n.Status = graph.StatusIdle
			n.LastRun = &t2
			n.LastOutputPreview = preview
		})

		// Remove completed task from queue.
		if rerr := queue.Remove(root, it.ID); rerr != nil {
			errs = append(errs, fmt.Sprintf("[%s] remove from queue: %v", it.Skill, rerr))
		}
		_ = syncGraph(root, cfg)
		completed++

		fmt.Printf("    ✓ done (%d input + %d output tokens)\n", result.InputTokens, result.OutputTokens)
	}

	// Desktop notification.
	if cfg.Coffee.NotifyOnComplete {
		msg := fmt.Sprintf("Done. %d task(s) completed.", completed)
		if len(errs) > 0 {
			msg += fmt.Sprintf(" %d error(s).", len(errs))
		}
		_ = beeep.Notify("☕ CoffeeGraph", msg, "")
	}

	// Summary.
	fmt.Printf("\n☕ Done. %d task(s) executed. Logs in .coffee/logs/\n", completed)
	if len(errs) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range errs {
			fmt.Printf("  • %s\n", e)
		}
	}
	return nil
}

func syncGraph(root string, cfg *config.Config) error {
	counts, err := queue.CountBySkill(root)
	if err != nil {
		return err
	}
	return graph.WriteProjectGraph(root, cfg, counts)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
