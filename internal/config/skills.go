package config

// KnownSkills lists the built-in template skills shipped with CoffeeGraph.
var KnownSkills = []string{
	"sales-closer",
	"content-engine",
	"lead-nurture",
	"life-os",
	"creator-stack",
}

// IsKnownSkill reports whether name matches a built-in template.
func IsKnownSkill(name string) bool {
	for _, s := range KnownSkills {
		if s == name {
			return true
		}
	}
	return false
}

// EnableSkill marks a skill as enabled in config, applying the project
// default model if no per-skill model is set.
func EnableSkill(c *Config, name string) {
	if c.Skills == nil {
		c.Skills = map[string]SkillEntry{}
	}
	e := c.Skills[name]
	e.Enabled = true
	if e.Model == "" {
		e.Model = c.DefaultModel
		if e.Model == "" {
			e.Model = "claude-sonnet-4-6"
		}
	}
	c.Skills[name] = e
}
