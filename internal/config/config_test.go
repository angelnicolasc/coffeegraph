package config

import (
	"os"
	"testing"
)

const validYAML = `
agency_name: "test-agency"
default_model: "claude-sonnet-4-6"
backend: "anthropic"
api_keys:
  anthropic: "sk-test"
skills:
  sales-closer:
    enabled: true
    model: "claude-opus-4-6"
  content-engine:
    enabled: false
coffee_mode:
  max_tasks_per_run: 5
`

func TestParseValid(t *testing.T) {
	cfg, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.AgencyName != "test-agency" {
		t.Fatalf("AgencyName = %q, want test-agency", cfg.AgencyName)
	}
	if cfg.DefaultModel != "claude-sonnet-4-6" {
		t.Fatalf("DefaultModel = %q", cfg.DefaultModel)
	}
	if cfg.APIKeys.Anthropic != "sk-test" {
		t.Fatalf("Anthropic key = %q", cfg.APIKeys.Anthropic)
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse([]byte(`{invalid yaml`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg, err := Parse([]byte(`agency_name: test`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.DefaultModel != "claude-sonnet-4-6" {
		t.Fatalf("default model not applied: %q", cfg.DefaultModel)
	}
	if cfg.Backend != "anthropic" {
		t.Fatalf("default backend not applied: %q", cfg.Backend)
	}
	if cfg.Coffee.MaxTasksPerRun != 3 {
		t.Fatalf("default max tasks = %d, want 3", cfg.Coffee.MaxTasksPerRun)
	}
}

func TestSkillEnabled(t *testing.T) {
	cfg, _ := Parse([]byte(validYAML))
	if !cfg.SkillEnabled("sales-closer") {
		t.Fatal("sales-closer should be enabled")
	}
	if cfg.SkillEnabled("content-engine") {
		t.Fatal("content-engine should be disabled")
	}
	// Unknown skills default to enabled.
	if !cfg.SkillEnabled("unknown-skill") {
		t.Fatal("unknown skills should default to enabled")
	}
}

func TestModelForSkill(t *testing.T) {
	cfg, _ := Parse([]byte(validYAML))
	if m := cfg.ModelForSkill("sales-closer"); m != "claude-opus-4-6" {
		t.Fatalf("sales-closer model = %q, want claude-opus-4-6", m)
	}
	if m := cfg.ModelForSkill("unknown"); m != "claude-sonnet-4-6" {
		t.Fatalf("unknown skill model = %q, want default", m)
	}
}

func TestBackendForSkill(t *testing.T) {
	cfg, _ := Parse([]byte(`
agency_name: test
backend: anthropic
skills:
  local-skill:
    enabled: true
    backend: ollama
`))
	if b := cfg.BackendForSkill("local-skill"); b != "ollama" {
		t.Fatalf("local-skill backend = %q, want ollama", b)
	}
	if b := cfg.BackendForSkill("other"); b != "anthropic" {
		t.Fatalf("other backend = %q, want anthropic", b)
	}
}

func TestEnvVarOverride(t *testing.T) {
	yaml := `agency_name: test
api_keys:
  anthropic: "from-file"
`
	t.Setenv("ANTHROPIC_API_KEY", "from-env")

	// Write to temp file.
	dir := t.TempDir()
	path := dir + "/config.yaml"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AnthropicKey() != "from-env" {
		t.Fatalf("expected env override, got %q", cfg.AnthropicKey())
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{AgencyName: ""}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty agency_name")
	}
	cfg.AgencyName = "valid"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDefaultConfigYAML(t *testing.T) {
	yaml := DefaultConfigYAML("my-agency")
	if yaml == "" {
		t.Fatal("DefaultConfigYAML returned empty string")
	}
	// Verify it parses correctly.
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("generated config does not parse: %v", err)
	}
	if cfg.AgencyName != "my-agency" {
		t.Fatalf("agency_name = %q, want my-agency", cfg.AgencyName)
	}
}
