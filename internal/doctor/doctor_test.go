package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coffeegraph/coffeegraph/internal/config"
)

func setupProject(t *testing.T, withAPIKey bool) (string, *config.Config) {
	t.Helper()
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills", "test-skill")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# Skill: test\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	apiKey := ""
	if withAPIKey {
		apiKey = "sk-ant-test"
	}
	cfg := &config.Config{
		AgencyName:   "test",
		DefaultModel: "test-model",
		Backend:      "anthropic",
		APIKeys: config.APIKeysConfig{
			Anthropic: apiKey,
		},
	}
	return root, cfg
}

func TestRunAllChecksPass(t *testing.T) {
	root, cfg := setupProject(t, true)
	checks, err := Run(root, cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	// API key should be OK.
	if !checks[0].OK {
		t.Errorf("Anthropic API key check should pass, got: %s", checks[0].Detail)
	}
	// Skills valid should be OK.
	last := checks[len(checks)-1]
	if last.Name != "Skills valid" || !last.OK {
		t.Errorf("Skills valid check should pass, got: %v", last)
	}
}

func TestRunMissingAPIKey(t *testing.T) {
	root, cfg := setupProject(t, false)
	// Ensure env var is not set for this test.
	t.Setenv("ANTHROPIC_API_KEY", "")
	checks, err := Run(root, cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if checks[0].OK {
		t.Error("Anthropic API key check should fail when no key is set")
	}
}

func TestRunSkillMissingSKILLMD(t *testing.T) {
	root, cfg := setupProject(t, true)
	// Create a skill directory without SKILL.md.
	badSkill := filepath.Join(root, "skills", "bad-skill")
	if err := os.MkdirAll(badSkill, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	checks, err := Run(root, cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	last := checks[len(checks)-1]
	if last.Name != "Skills valid" {
		t.Fatalf("unexpected last check: %v", last)
	}
	if last.OK {
		t.Error("Skills valid should fail when a skill folder is missing SKILL.md")
	}
}

func TestRunOllamaBackend(t *testing.T) {
	root, cfg := setupProject(t, true)
	cfg.Backend = "ollama"
	checks, err := Run(root, cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// Should include an Ollama reachable check.
	found := false
	for _, c := range checks {
		if c.Name == "Ollama reachable" {
			found = true
			// Ollama is not running in test, so this should fail.
			if c.OK {
				t.Error("Ollama check should fail when ollama is not running")
			}
		}
	}
	if !found {
		t.Error("expected Ollama reachable check for ollama backend")
	}
}
