package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunDeployOpenClaw generates SOUL.md files in OpenClaw format.
func RunDeployOpenClaw() error {
	return runDeploy("openclaw")
}

// RunDeployHermes generates AGENT.md files in Hermes Agent format.
func RunDeployHermes() error {
	return runDeploy("hermes")
}

func runDeploy(kind string) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}

	outDir := filepath.Join(root, ".coffee", "deploy", kind)
	_ = os.MkdirAll(outDir, 0o755)

	n := 0
	for name, ent := range cfg.Skills {
		if !ent.Enabled {
			continue
		}
		skillPath := filepath.Join(root, "skills", name, "SKILL.md")
		body, err := os.ReadFile(skillPath)
		if err != nil {
			fmt.Printf("  ⚠ skipping %s: %v\n", name, err)
			continue
		}

		var out string
		var filename string
		switch kind {
		case "openclaw":
			out = convertToOpenClaw(name, string(body), cfg)
			filename = name + "-SOUL.md"
		case "hermes":
			out = convertToHermes(name, string(body), cfg)
			filename = name + "-AGENT.md"
		}
		if err := os.WriteFile(filepath.Join(outDir, filename), []byte(out), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
		n++
	}

	wh := strings.TrimSpace(cfg.APIKeys.N8nWebhook)
	if wh == "" {
		wh = os.Getenv("COFFEEGRAPH_N8N_WEBHOOK")
	}
	if wh != "" {
		fmt.Printf("n8n webhook configured (%s); POST integration pending in your instance.\n", wh)
	}
	fmt.Printf("✓ %d skill(s) exported to %s\n", n, outDir)
	return nil
}

// convertToOpenClaw transforms a CoffeeGraph SKILL.md into an OpenClaw
// SOUL.md format. OpenClaw expects: Purpose, Personality, Behaviour Rules,
// Knowledge, and Conversation Flow sections.
func convertToOpenClaw(name, body string, cfg *config.Config) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# SOUL — %s\n\n", name))

	// Extract sections from the SKILL.md
	identity := extractSection(body, "Identity")
	workflow := extractSection(body, "Workflow")
	constraints := extractSection(body, "Constraints")
	outputFmt := extractSection(body, "Output Format")

	b.WriteString("## Purpose\n")
	if identity != "" {
		b.WriteString(identity + "\n\n")
	} else {
		b.WriteString(fmt.Sprintf("Autonomous agent for the %s skill.\n\n", name))
	}

	b.WriteString("## Personality\n")
	b.WriteString("- Direct and execution-focused\n")
	b.WriteString("- Never asks questions; works with available information\n")
	b.WriteString("- Marks uncertainties with [FILL IN]\n\n")

	b.WriteString("## Behaviour Rules\n")
	if constraints != "" {
		b.WriteString(constraints + "\n\n")
	}

	b.WriteString("## Knowledge\n")
	b.WriteString("- Business context from index.md (injected as CONTEXT)\n")
	b.WriteString(fmt.Sprintf("- Model: %s\n\n", cfg.ModelForSkill(name)))

	b.WriteString("## Conversation Flow\n")
	if workflow != "" {
		b.WriteString(workflow + "\n\n")
	}

	if outputFmt != "" {
		b.WriteString("## Output Specification\n")
		b.WriteString(outputFmt + "\n")
	}

	return b.String()
}

// convertToHermes transforms a CoffeeGraph SKILL.md into a Hermes Agent
// AGENT.md format. Hermes expects: Agent Name, Description, Capabilities,
// Instructions, and Constraints sections.
func convertToHermes(name, body string, cfg *config.Config) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# AGENT — %s\n\n", name))

	identity := extractSection(body, "Identity")
	workflow := extractSection(body, "Workflow")
	constraints := extractSection(body, "Constraints")
	inputs := extractSection(body, "Inputs")

	b.WriteString("## Agent Name\n")
	b.WriteString(name + "\n\n")

	b.WriteString("## Description\n")
	if identity != "" {
		b.WriteString(identity + "\n\n")
	}

	b.WriteString("## Capabilities\n")
	if inputs != "" {
		b.WriteString("### Inputs\n")
		b.WriteString(inputs + "\n\n")
	}

	b.WriteString("## Instructions\n")
	if workflow != "" {
		b.WriteString(workflow + "\n\n")
	}

	b.WriteString("## Constraints\n")
	if constraints != "" {
		b.WriteString(constraints + "\n")
	}

	return b.String()
}

// extractSection finds a ## heading and returns its content until the next
// ## heading or end of document.
func extractSection(body, heading string) string {
	marker := "## " + heading
	idx := strings.Index(body, marker)
	if idx < 0 {
		return ""
	}
	start := idx + len(marker)
	// Skip the rest of the heading line.
	if nl := strings.Index(body[start:], "\n"); nl >= 0 {
		start += nl + 1
	}
	// Find the next ## heading.
	rest := body[start:]
	end := strings.Index(rest, "\n## ")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}
