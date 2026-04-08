package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/fsutil"
	"github.com/coffeegraph/coffeegraph/internal/graph"
	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/templates"
)

const indexTemplate = `# My AI Agency — Team Brain

## Mission
[Edit this: what does your business do, what is the primary objective]

## Active Team
- **sales-closer**: Qualifies leads and writes hyper-personalized sales emails
- **content-engine**: Monitors trends and generates viral content
- **life-os**: Daily brief, finances, schedule

## Priorities (in order)
1. New leads from CRM → sales-closer
2. Trends on X about [topic] → content-engine
3. Morning brief → life-os

## Business Context
- Industry: [fill in]
- ICP (ideal customer profile): [fill in]
- Communication tone: [formal/casual/technical]
- Main channels: [X, LinkedIn, email, etc.]

## Global Rules
- Always respond in the language of the lead/audience
- Don't mention competitors
- Primary CTA: [link or action]
`

// RunInit executes "coffeegraph init <name>".
func RunInit(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("usage: coffeegraph init <name>")
	}
	name = strings.TrimSpace(name)
	if fi, err := os.Stat(name); err == nil && fi.IsDir() {
		return fmt.Errorf("directory %q already exists", name)
	}

	// Create directory structure.
	dirs := []string{
		name,
		filepath.Join(name, "skills"),
		filepath.Join(name, ".coffee", "logs"),
		filepath.Join(name, ".coffee", "snapshots"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}

	// Write config.yaml.
	agency := strings.ReplaceAll(name, "-", " ")
	cfgPath := filepath.Join(name, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(config.DefaultConfigYAML(agency)), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Write index.md.
	if err := os.WriteFile(filepath.Join(name, "index.md"), []byte(indexTemplate), 0644); err != nil {
		return fmt.Errorf("write index.md: %w", err)
	}

	// Copy templates for reference.
	templatesDst := filepath.Join(name, "templates")
	if err := os.MkdirAll(templatesDst, 0755); err != nil {
		return err
	}
	if err := fsutil.CopyFSToDir(templates.FS, ".", templatesDst); err != nil {
		return fmt.Errorf("copy templates: %w", err)
	}

	// Generate initial graph.json (hub only).
	cfg, err := config.LoadRaw(cfgPath)
	if err != nil {
		return err
	}
	if err := graph.WriteProjectGraph(name, cfg, map[string]int{}); err != nil {
		return fmt.Errorf("write graph: %w", err)
	}

	// Create empty queue.
	if err := queue.Set(name, nil); err != nil {
		return fmt.Errorf("write queue: %w", err)
	}

	fmt.Printf(`
╭────────────────────────────────────────────────╮
│  ☕ CoffeeGraph — project initialized          │
╰────────────────────────────────────────────────╯

Created in ./%s:
  • config.yaml          — API keys & settings (edit this first)
  • index.md             — your business context
  • skills/              — empty, fill with "coffeegraph add"
  • templates/           — reference skill templates
  • graph.json           — auto-generated dependency graph
  • .coffee/queue.json   — task backlog

Next steps:
  cd %s
  coffeegraph add sales-closer
  coffeegraph dashboard
`, name, name)
	return nil
}
