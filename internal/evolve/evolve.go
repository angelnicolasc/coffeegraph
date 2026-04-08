// Package evolve suggests SKILL.md improvements from recent execution logs.
package evolve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/fsutil"
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// Suggest generates an updated SKILL.md proposal for a given skill.
func Suggest(ctx context.Context, root string, cfg *config.Config, skill string) (outPath, proposed string, err error) {
	entry, err := latestBySkill(root, skill)
	if err != nil {
		return "", "", err
	}
	skillPath := filepath.Join(root, "skills", skill, "SKILL.md")
	cur, err := os.ReadFile(skillPath)
	if err != nil {
		return "", "", fmt.Errorf("read skill file: %w", err)
	}

	if strings.TrimSpace(cfg.AnthropicKey()) == "" {
		return "", "", claude.ErrMissingAPIKey
	}
	client := &claude.Client{APIKey: cfg.AnthropicKey(), Model: cfg.ModelForSkill(skill)}
	sys := "You improve CoffeeGraph skills. Return only the new SKILL.md markdown body."
	user := fmt.Sprintf(`Improve this SKILL.md based on latest execution output.
Keep the skill identity and structure. Add a changelog entry at bottom.

CURRENT SKILL.md:
%s

LATEST JOB:
Task: %s
Output:
%s`, string(cur), entry.Task, entry.Result)
	res, err := client.Complete(ctx, sys, user)
	if err != nil {
		return "", "", err
	}
	return skillPath, strings.TrimSpace(res.Text), nil
}

func latestBySkill(root, skill string) (*logs.Entry, error) {
	all, err := logs.List(root)
	if err != nil {
		return nil, err
	}
	for _, e := range all {
		if e.Skill == skill {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("no logs found for skill %q", skill)
}

// Apply writes a proposed SKILL.md update and appends changelog marker.
func Apply(path, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty evolved content")
	}
	if !strings.Contains(content, "## Changelog") {
		content += "\n\n## Changelog\n- Updated by coffeegraph evolve.\n"
	}
	return fsutil.AtomicWriteFile(path, []byte(content+"\n"), 0o644)
}

// DetectRoot resolves project root for evolve commands.
func DetectRoot() (string, error) {
	return project.FindRoot("")
}
