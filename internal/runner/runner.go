// Package runner executes CoffeeGraph tasks through a shared engine used by
// CLI commands, bots, and MCP.
package runner

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
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/queue"
)

// Mode tweaks visual output while preserving execution behavior.
type Mode string

const (
	// ModeNormal runs with regular messaging.
	ModeNormal Mode = "normal"
	// ModeUrgent uses high-pressure copy.
	ModeUrgent Mode = "urgent"
	// ModeChill uses calming copy.
	ModeChill Mode = "chill"
)

// Engine coordinates queue execution.
type Engine struct {
	Root string
	Cfg  *config.Config
}

// ExecSummary captures a run outcome.
type ExecSummary struct {
	Completed int
	Errors    []string
	Paths     []string
}

// TaskResult captures one executed task output.
type TaskResult struct {
	LogPath      string
	Text         string
	InputTokens  int
	OutputTokens int
}

// ExecutePending runs up to max tasks from queue.
func (e *Engine) ExecutePending(ctx context.Context, max int, mode Mode) (ExecSummary, error) {
	items, err := queue.Read(e.Root)
	if err != nil {
		return ExecSummary{}, err
	}
	if len(items) == 0 {
		return ExecSummary{}, nil
	}
	if max <= 0 || max > len(items) {
		max = len(items)
	}
	var out ExecSummary
	for i := 0; i < max; i++ {
		if err := ctx.Err(); err != nil {
			out.Errors = append(out.Errors, err.Error())
			break
		}
		if mode == ModeUrgent {
			fmt.Printf("DEADLINE IN: %02d:%02d and counting\n", 4-i%5, 59-i*7%60)
			fmt.Println(urgentMsg(i))
		}
		if mode == ModeChill {
			fmt.Println("Your agents are vibing. No rush, you've got this.")
		}
		r, err := e.ExecuteTask(ctx, items[i])
		if err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("[%s] %v", items[i].Skill, err))
			continue
		}
		out.Completed++
		out.Paths = append(out.Paths, r.LogPath)
		if mode == ModeUrgent {
			fmt.Print("\a")
		}
		if err := queue.Remove(e.Root, items[i].ID); err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("[%s] remove queue item: %v", items[i].Skill, err))
		}
	}
	_ = e.syncGraph()
	return out, nil
}

// ExecuteTask executes a single task and writes log output.
func (e *Engine) ExecuteTask(ctx context.Context, it queue.Item) (TaskResult, error) {
	if !e.Cfg.SkillEnabled(it.Skill) {
		return TaskResult{}, fmt.Errorf("skill %s is disabled in config", it.Skill)
	}
	start := time.Now().UTC()
	skillPath := filepath.Join(e.Root, "skills", it.Skill, "SKILL.md")
	skillBody, err := os.ReadFile(skillPath)
	if err != nil {
		return TaskResult{}, fmt.Errorf("read skill: %w", err)
	}
	indexBody, err := os.ReadFile(filepath.Join(e.Root, "index.md"))
	if err != nil {
		return TaskResult{}, fmt.Errorf("read index.md: %w", err)
	}
	_ = graph.UpdateSkill(e.Root, it.Skill, func(n *graph.Node) {
		n.Status = graph.StatusRunning
		now := time.Now().UTC()
		n.LastRun = &now
	})
	_ = e.syncGraph()

	client, err := e.newClient(it.Skill)
	if err != nil {
		return TaskResult{}, err
	}
	userPrompt := fmt.Sprintf("CONTEXT:\n%s\n\nTASK:\n%s\n\nDATA:\n%s\n", string(indexBody), it.Task, it.Data)
	res, err := client.Complete(ctx, string(skillBody), userPrompt)
	if err != nil {
		_ = graph.UpdateSkill(e.Root, it.Skill, func(n *graph.Node) {
			n.Status = graph.StatusError
			n.LastOutputPreview = err.Error()
		})
		_ = e.syncGraph()
		return TaskResult{}, err
	}
	finish := time.Now().UTC()
	lp, err := logs.Write(e.Root, logs.Entry{
		Skill:      it.Skill,
		Task:       it.Task,
		Status:     "DONE",
		StartedAt:  start,
		FinishedAt: finish,
		Result:     res.Text,
	})
	if err != nil {
		return TaskResult{}, err
	}
	preview := strings.TrimSpace(res.Text)
	if len(preview) > 120 {
		preview = preview[:117] + "..."
	}
	_ = graph.UpdateSkill(e.Root, it.Skill, func(n *graph.Node) {
		n.Status = graph.StatusIdle
		n.LastOutputPreview = preview
		now := time.Now().UTC()
		n.LastRun = &now
	})
	_ = e.syncGraph()
	return TaskResult{
		LogPath:      lp,
		Text:         res.Text,
		InputTokens:  res.InputTokens,
		OutputTokens: res.OutputTokens,
	}, nil
}

func (e *Engine) syncGraph() error {
	counts, err := queue.CountBySkill(e.Root)
	if err != nil {
		return err
	}
	return graph.WriteProjectGraph(e.Root, e.Cfg, counts)
}

func (e *Engine) newClient(skill string) (*clientShim, error) {
	backend := strings.ToLower(strings.TrimSpace(e.Cfg.BackendForSkill(skill)))
	model := e.Cfg.ModelForSkill(skill)
	switch backend {
	case "ollama":
		return &clientShim{ollama: &claude.OllamaClient{Model: model}}, nil
	case "anthropic", "":
		key := e.Cfg.AnthropicKey()
		if strings.TrimSpace(key) == "" {
			return nil, claude.ErrMissingAPIKey
		}
		return &clientShim{anthropic: &claude.Client{APIKey: key, Model: model}}, nil
	default:
		return nil, fmt.Errorf("unsupported backend %q", backend)
	}
}

type clientShim struct {
	anthropic *claude.Client
	ollama    *claude.OllamaClient
}

func (c *clientShim) Complete(ctx context.Context, systemPrompt, userPrompt string) (claude.Result, error) {
	if c.anthropic != nil {
		return c.anthropic.Complete(ctx, systemPrompt, userPrompt)
	}
	return c.ollama.Complete(ctx, systemPrompt, userPrompt)
}

func urgentMsg(i int) string {
	msgs := []string{
		"NO SLACKING. CLOSE THOSE DEALS.",
		"YOUR COMPETITION IS ALREADY RUNNING.",
		"WHY IS THIS TAKING SO LONG.",
	}
	return msgs[i%len(msgs)]
}
