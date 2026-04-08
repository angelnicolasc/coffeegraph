// Package registry integrates community skills.
package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/fsutil"
)

const baseRaw = "https://raw.githubusercontent.com/angelnicolasc/awesome-coffeegraph-skills/main"

// List fetches available skills from index.json or README fallback.
func List() ([]string, error) {
	resp, err := http.Get(baseRaw + "/index.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("registry index request failed: %s", resp.Status)
	}
	var out []string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// Install downloads a skill SKILL.md from registry.
func Install(root, name string) (string, error) {
	url := fmt.Sprintf("%s/%s/SKILL.md", baseRaw, name)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("skill %s not found in registry (%s)", name, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(root, "skills", name, "SKILL.md")
	if err := fsutil.AtomicWriteFile(dst, b, 0644); err != nil {
		return "", err
	}
	return dst, nil
}

// LocalCustomSkills returns skill names that are not in built-in templates.
func LocalCustomSkills(root string) ([]string, error) {
	ents, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	return out, nil
}

// PublishTemplateURL builds the compare URL to open a PR in registry.
func PublishTemplateURL(skillNames []string) string {
	body := "## New CoffeeGraph skills\n\nPlease review these skills:\n"
	for _, s := range skillNames {
		body += "- " + s + "\n"
	}
	return "https://github.com/angelnicolasc/awesome-coffeegraph-skills/compare/main?expand=1"
}

// ParseListFromText parses newline-separated skill names.
func ParseListFromText(raw string) []string {
	lines := strings.Split(raw, "\n")
	var out []string
	for _, l := range lines {
		v := strings.TrimSpace(l)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
