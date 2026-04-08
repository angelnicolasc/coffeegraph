package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunRoast roasts the current project context.
func RunRoast() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	if cfg.AnthropicKey() == "" {
		return claude.ErrMissingAPIKey
	}
	index, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		return err
	}
	client := &claude.Client{APIKey: cfg.AnthropicKey(), Model: cfg.DefaultModel}
	res, err := client.Complete(context.Background(), "Roast mercilessly as a senior VC.", string(index))
	if err != nil {
		return err
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3b3b")).Bold(true)
	fmt.Println(style.Render(res.Text))
	return nil
}

// RunParty runs a round-robin playful conversation across skills.
func RunParty() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	ents, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return fmt.Errorf("no skills installed")
	}
	var transcript string
	msg := "Let's align on today's top growth move."
	for round := 1; round <= 5; round++ {
		fmt.Printf("Round %d\n", round)
		for _, s := range names {
			line := fmt.Sprintf("%s: %s", s, msg)
			fmt.Println(line)
			transcript += line + "\n"
			msg = "Building on that, next action?"
		}
	}
	p, _ := logs.Write(root, logs.Entry{Skill: "party", Task: "round-robin", Result: transcript, FinishedAt: time.Now().UTC(), StartedAt: time.Now().UTC()})
	fmt.Println("Party transcript saved. Config default model:", cfg.DefaultModel)
	fmt.Println(p)
	return nil
}

// RunNap pauses until a keypress while showing ascii animation.
func RunNap() error {
	frames := []string{
		" ( -_-) z",
		" ( -_-) zz",
		" ( -_-) zzz",
	}
	fmt.Println("Coffee cup nap mode. Press Enter to resume.")
	done := make(chan struct{})
	go func() {
		var in string
		_, _ = fmt.Scanln(&in)
		close(done)
	}()
	i := 0
	for {
		select {
		case <-done:
			fmt.Println("Resuming tasks.")
			return nil
		default:
			fmt.Printf("\r%s", frames[i%len(frames)])
			time.Sleep(400 * time.Millisecond)
			i++
		}
	}
}
