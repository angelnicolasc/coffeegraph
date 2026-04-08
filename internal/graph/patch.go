package graph

import (
	"fmt"
	"path/filepath"
	"time"
)

// UpdateSkill atomically loads graph.json, applies fn to the named skill
// node, updates generated_at, and writes back. Returns an error if the
// graph cannot be loaded or the skill node is not found.
func UpdateSkill(projectRoot string, skillID string, fn func(*Node)) error {
	p := filepath.Join(projectRoot, "graph.json")
	g, err := LoadFile(p)
	if err != nil {
		return fmt.Errorf("update skill %s: %w", skillID, err)
	}
	n := g.NodeByID(skillID)
	if n == nil || n.Type != NodeTypeSkill {
		return fmt.Errorf("update skill: node %q not found or not a skill", skillID)
	}
	fn(n)
	g.GeneratedAt = time.Now().UTC()
	return WriteFile(p, g)
}
