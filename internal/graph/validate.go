package graph

import (
	"fmt"
	"os"
	"strings"
)

// ValidateSkillFile checks that a SKILL.md file at the given path has the
// minimum required structure. Returns a clear, actionable error if the file
// is missing, empty, or lacks required sections.
func ValidateSkillFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("SKILL.md not found at %s — every skill folder must contain a SKILL.md file", path)
		}
		return fmt.Errorf("cannot read SKILL.md: %w", err)
	}
	content := strings.TrimSpace(string(b))
	if content == "" {
		return fmt.Errorf("SKILL.md is empty at %s — add at least a '# Skill:' header and '## Identity' section", path)
	}

	hasHeader := strings.Contains(content, "# Skill:") || strings.Contains(content, "# skill:")
	hasIdentity := strings.Contains(content, "## Identity") || strings.Contains(content, "## identity")

	if !hasHeader && !hasIdentity {
		return fmt.Errorf(
			"SKILL.md at %s is missing required structure.\n"+
				"  Expected at least one of:\n"+
				"    • '# Skill: <name>' header\n"+
				"    • '## Identity' section\n"+
				"  See templates/ for examples of valid SKILL.md files",
			path,
		)
	}

	hasWorkflow := strings.Contains(content, "## Workflow") || strings.Contains(content, "## workflow")
	hasOutput := strings.Contains(content, "## Output") || strings.Contains(content, "## output")

	if !hasWorkflow && !hasOutput {
		return fmt.Errorf(
			"SKILL.md at %s is missing workflow or output sections.\n"+
				"  Expected at least one of:\n"+
				"    • '## Workflow' section\n"+
				"    • '## Output Format' section\n"+
				"  The skill needs instructions on what to do and what to produce",
			path,
		)
	}

	return nil
}
