package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/fsutil"
	"github.com/coffeegraph/coffeegraph/internal/graph"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/templates"
)

// RunAdd executes "coffeegraph add <skill>".
func RunAdd(skillName string) error {
	skillName = strings.TrimSpace(skillName)
	if skillName == "" {
		return fmt.Errorf("usage: coffeegraph add <skill>")
	}

	// Verify the template exists in the embedded filesystem.
	if _, err := fs.Stat(templates.FS, filepath.Join(skillName, "SKILL.md")); err != nil {
		return fmt.Errorf("skill %q not found.\nAvailable skills: %s",
			skillName, strings.Join(config.KnownSkills, ", "))
	}

	root, err := project.FindRoot("")
	if err != nil {
		return err
	}

	// Check if already installed.
	dst := filepath.Join(root, "skills", skillName)
	if _, serr := os.Stat(filepath.Join(dst, "SKILL.md")); serr == nil {
		return fmt.Errorf("skill %q is already installed in skills/%s/", skillName, skillName)
	}

	// Update config.yaml — enable this skill.
	cfgPath := filepath.Join(root, "config.yaml")
	cfg, err := config.LoadRaw(cfgPath)
	if err != nil {
		return err
	}
	config.EnableSkill(cfg, skillName)
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	// Copy template files to skills/<name>/.
	sub, err := fs.Sub(templates.FS, skillName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := fsutil.CopyFSToDir(sub, ".", dst); err != nil {
		return fmt.Errorf("copy skill template: %w", err)
	}

	// Regenerate graph.json.
	counts, err := queue.CountBySkill(root)
	if err != nil {
		return err
	}
	if err := graph.WriteProjectGraph(root, cfg, counts); err != nil {
		return fmt.Errorf("regenerate graph: %w", err)
	}

	fmt.Printf("✓ %s added. Edit skills/%s/SKILL.md to customize.\n", skillName, skillName)

	// Warn about n8n if needed.
	if requiresN8n(skillName) && strings.TrimSpace(cfg.APIKeys.N8nWebhook) == "" && os.Getenv("COFFEEGRAPH_N8N_WEBHOOK") == "" {
		fmt.Println("⚠  Configure n8n: set api_keys.n8n_webhook in config.yaml or COFFEEGRAPH_N8N_WEBHOOK env var.")
	}
	return nil
}

func requiresN8n(skill string) bool {
	switch skill {
	case "sales-closer", "content-engine", "lead-nurture":
		return true
	default:
		return false
	}
}
