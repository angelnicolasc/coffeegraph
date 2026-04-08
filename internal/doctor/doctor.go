// Package doctor runs dependency and project health checks.
package doctor

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/config"
)

// CheckResult represents one doctor check.
type CheckResult struct {
	Name   string
	OK     bool
	Detail string
}

// Run validates external dependencies and local skill files.
func Run(root string, cfg *config.Config) ([]CheckResult, error) {
	out := []CheckResult{
		{Name: "Anthropic API key", OK: strings.TrimSpace(cfg.AnthropicKey()) != "", Detail: "set ANTHROPIC_API_KEY or config.api_keys.anthropic"},
		{Name: "Telegram bot token", OK: strings.TrimSpace(cfg.TelegramToken()) != "", Detail: "required only for coffeegraph bot"},
		{Name: "GitHub token", OK: strings.TrimSpace(cfg.GitHubToken()) != "", Detail: "required for gist sharing"},
	}
	if strings.EqualFold(cfg.Backend, "ollama") {
		ok, detail := ping("http://localhost:11434/api/tags")
		out = append(out, CheckResult{Name: "Ollama reachable", OK: ok, Detail: detail})
	}
	if strings.TrimSpace(cfg.APIKeys.N8nWebhook) != "" {
		ok, detail := ping(cfg.APIKeys.N8nWebhook)
		out = append(out, CheckResult{Name: "n8n webhook reachable", OK: ok, Detail: detail})
	}
	skillsDir := filepath.Join(root, "skills")
	ents, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}
	ok := true
	missing := 0
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md")); err != nil {
			ok = false
			missing++
		}
	}
	detail := "all installed skills have SKILL.md"
	if !ok {
		detail = fmt.Sprintf("%d skill folders missing SKILL.md", missing)
	}
	out = append(out, CheckResult{Name: "Skills valid", OK: ok, Detail: detail})
	return out, nil
}

func ping(url string) (ok bool, detail string) {
	c := http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return false, resp.Status
	}
	return true, resp.Status
}
