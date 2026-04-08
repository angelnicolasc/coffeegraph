package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunSuggest asks Claude to suggest new skills based on the current
// project context and installed skills.
func RunSuggest() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	apiKey := cfg.AnthropicKey()
	if strings.TrimSpace(apiKey) == "" {
		return claude.ErrMissingAPIKey
	}

	index, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		return fmt.Errorf("read index.md: %w", err)
	}

	var active []string
	for k, v := range cfg.Skills {
		if v.Enabled {
			active = append(active, k)
		}
	}

	sys := "You are an automation consultant. Respond in English, concise format."
	user := fmt.Sprintf(`Given this business context and these active CoffeeGraph skills (%s), what new skill (kebab-case name) would add the most value? Suggest exactly 3 options with: name, one-line description, and reasoning.

CONTEXT:
%s
`, strings.Join(active, ", "), string(index))

	ctx := context.Background()
	client := &claude.Client{APIKey: apiKey, Model: cfg.DefaultModel}
	result, err := client.Complete(ctx, sys, user)
	if err != nil {
		return err
	}
	fmt.Println(result.Text)
	return nil
}
