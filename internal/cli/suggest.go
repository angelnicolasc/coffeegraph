package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/claude"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/fsutil"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunSuggest asks model for skill suggestions. If vaultPath is provided, it
// uses markdown files from that folder as context and writes ready SKILL.md files.
func RunSuggest(vaultPath string, deep bool) error {
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
	var contextChunks []string
	if strings.TrimSpace(vaultPath) != "" {
		files, err := collectMarkdown(vaultPath, deep)
		if err != nil {
			return err
		}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err == nil {
				contextChunks = append(contextChunks, fmt.Sprintf("FILE: %s\n%s", filepath.Base(f), string(b)))
			}
		}
	} else {
		index, err := os.ReadFile(filepath.Join(root, "index.md"))
		if err != nil {
			return fmt.Errorf("read index.md: %w", err)
		}
		contextChunks = append(contextChunks, string(index))
	}

	sys := "You are an automation consultant. Return exactly 3 SKILL.md files separated by '\\n===FILE===\\n'."
	user := fmt.Sprintf(`Here is the user's knowledge base. Suggest 3 new CoffeeGraph skills with full SKILL.md drafts.

%s`, strings.Join(contextChunks, "\n\n"))
	client := &claude.Client{APIKey: apiKey, Model: cfg.DefaultModel}
	result, err := client.Complete(context.Background(), sys, user)
	if err != nil {
		return err
	}
	if strings.TrimSpace(vaultPath) == "" {
		fmt.Println(result.Text)
		return nil
	}
	parts := strings.Split(result.Text, "\n===FILE===\n")
	outDir := filepath.Join(root, "skills", "suggested")
	for i, p := range parts {
		name := fmt.Sprintf("suggested-%d.md", i+1)
		if strings.HasPrefix(strings.TrimSpace(p), "# Skill:") {
			line := strings.SplitN(strings.TrimSpace(p), "\n", 2)[0]
			n := strings.TrimSpace(strings.TrimPrefix(line, "# Skill:"))
			n = strings.ReplaceAll(strings.ToLower(n), " ", "-")
			if n != "" {
				name = n + ".md"
			}
		}
		dst := filepath.Join(outDir, name)
		if err := fsutil.AtomicWriteFile(dst, []byte(strings.TrimSpace(p)+"\n"), 0644); err != nil {
			return err
		}
		fmt.Println("Wrote", dst)
	}
	return nil
}

func collectMarkdown(root string, deep bool) ([]string, error) {
	var out []string
	if !deep {
		ents, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, e := range ents {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				out = append(out, filepath.Join(root, e.Name()))
			}
		}
		return out, nil
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}
