// Package config handles loading, validating, and writing the
// project-level config.yaml that controls CoffeeGraph behaviour.
package config

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Config mirrors the config.yaml that lives in every CoffeeGraph project.
type Config struct {
	AgencyName   string                `yaml:"agency_name"`
	DefaultModel string                `yaml:"default_model"`
	APIKeys      APIKeysConfig         `yaml:"api_keys"`
	Skills       map[string]SkillEntry `yaml:"skills"`
	Coffee       CoffeeConfig          `yaml:"coffee_mode"`
}

// APIKeysConfig holds sensitive credentials.
type APIKeysConfig struct {
	Anthropic  string `yaml:"anthropic"`
	N8nWebhook string `yaml:"n8n_webhook"`
}

// CoffeeConfig controls coffee mode behaviour.
type CoffeeConfig struct {
	MaxTasksPerRun    int  `yaml:"max_tasks_per_run"`
	NotifyOnComplete  bool `yaml:"notify_on_complete"`
	NotificationSound bool `yaml:"notification_sound"`
}

// SkillEntry is the per-skill configuration inside config.yaml.
type SkillEntry struct {
	Enabled    bool   `yaml:"enabled"`
	Model      string `yaml:"model"`
	N8nWebhook string `yaml:"n8n_webhook"`
}

// Parse decodes raw YAML bytes into a Config, applying defaults.
func Parse(b []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	c.applyDefaults()
	return &c, nil
}

// Write serializes config to the given path.
func Write(path string, c *Config) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, b, 0644)
}

// LoadRaw reads config.yaml without applying environment variable
// overrides. Use this when you intend to modify and write back.
func LoadRaw(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return Parse(b)
}

// Load reads config.yaml and overlays environment variables.
// Environment variables take precedence over file values:
//   - ANTHROPIC_API_KEY
//   - COFFEEGRAPH_N8N_WEBHOOK
//   - COFFEEGRAPH_MODEL
func Load(path string) (*Config, error) {
	c, err := LoadRaw(path)
	if err != nil {
		return nil, err
	}
	if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
		c.APIKeys.Anthropic = k
	}
	if w := os.Getenv("COFFEEGRAPH_N8N_WEBHOOK"); w != "" {
		c.APIKeys.N8nWebhook = w
	}
	if m := os.Getenv("COFFEEGRAPH_MODEL"); m != "" {
		c.DefaultModel = m
	}
	c.applyDefaults()
	return c, nil
}

func (c *Config) applyDefaults() {
	if c.DefaultModel == "" {
		c.DefaultModel = "claude-sonnet-4-6"
	}
	if c.Coffee.MaxTasksPerRun <= 0 {
		c.Coffee.MaxTasksPerRun = 3
	}
}

// ModelForSkill returns the model to use for a given skill, falling back
// to the project default if no per-skill override is configured.
func (c *Config) ModelForSkill(skillName string) string {
	if s, ok := c.Skills[skillName]; ok && s.Model != "" {
		return s.Model
	}
	return c.DefaultModel
}

// SkillEnabled reports whether the named skill is enabled. Skills not
// explicitly listed in config are considered enabled by default.
func (c *Config) SkillEnabled(skillName string) bool {
	s, ok := c.Skills[skillName]
	if !ok {
		return true
	}
	return s.Enabled
}

// AnthropicKey returns the API key, preferring the env var.
func (c *Config) AnthropicKey() string {
	if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
		return k
	}
	return c.APIKeys.Anthropic
}

// Validate performs structural checks on the config.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.AgencyName) == "" {
		return fmt.Errorf("config: agency_name is required")
	}
	return nil
}

// DefaultConfigYAML renders the initial config.yaml for a new project.
// Uses text/template instead of string concatenation for correctness.
func DefaultConfigYAML(agencyName string) string {
	const tmpl = `# CoffeeGraph config
agency_name: "{{.Name}}"
default_model: "claude-sonnet-4-6"

api_keys:
  anthropic: ""                           # or set ANTHROPIC_API_KEY env var
  n8n_webhook: ""                         # base URL of your n8n instance

skills:
  sales-closer:
    enabled: true
    model: "claude-sonnet-4-6"
    n8n_webhook: ""                       # per-skill override
  content-engine:
    enabled: true
    model: "claude-sonnet-4-6"
  lead-nurture:
    enabled: false
  life-os:
    enabled: true
  creator-stack:
    enabled: false

coffee_mode:
  max_tasks_per_run: 3                    # tasks per coffee run
  notify_on_complete: true
  notification_sound: true
`
	t := template.Must(template.New("config").Parse(tmpl))
	var b strings.Builder
	_ = t.Execute(&b, struct{ Name string }{Name: agencyName})
	return b.String()
}
