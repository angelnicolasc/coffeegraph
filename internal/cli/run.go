package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/graph"
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

// RunRun executes a single task against a skill directly, without the queue.
// This is the primary command for ad-hoc skill execution.
func RunRun(skill, task string) error {
	if strings.TrimSpace(skill) == "" || strings.TrimSpace(task) == "" {
		return fmt.Errorf("usage: coffeegraph run <skill> <task>")
	}
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}

	// Validate SKILL.md before executing.
	skillPath := filepath.Join(root, "skills", skill, "SKILL.md")
	if err := graph.ValidateSkillFile(skillPath); err != nil {
		return err
	}

	if !cfg.SkillEnabled(skill) {
		return fmt.Errorf("skill %q is disabled in config.yaml", skill)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	go func() {
		<-sig
		fmt.Println("\nGraceful shutdown requested.")
		cancel()
	}()

	fmt.Printf("Running skill %q...\n", skill)

	engine := &runner.Engine{Root: root, Cfg: cfg}
	it := queue.Item{
		Skill:    skill,
		Task:     task,
		Priority: 1,
	}

	// Use a temporary fake claude client to handle missing API key early.
	apiKey := cfg.AnthropicKey()
	backend := strings.ToLower(strings.TrimSpace(cfg.BackendForSkill(skill)))
	if (backend == "anthropic" || backend == "") && strings.TrimSpace(apiKey) == "" {
		return claude.ErrMissingAPIKey
	}

	result, err := engine.ExecuteTask(ctx, it)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Display output.
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("Skill:  %s\n", skill)
	fmt.Printf("Tokens: %d in / %d out\n", result.InputTokens, result.OutputTokens)
	fmt.Printf("Log:    %s\n", result.LogPath)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
	fmt.Println(result.Text)

	// Log the execution.
	_, _ = logs.Write(root, logs.Entry{
		Skill:  skill,
		Task:   task,
		Status: "DONE",
		Result: result.Text,
	})

	return nil
}
