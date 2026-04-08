// Package graph defines the CoffeeGraph data model and persistence layer.
//
// A Graph represents the current state of a user's agency: a central "hub"
// node (the index.md orchestrator) connected to one or more "skill" nodes.
// The graph is serialized as graph.json in the project root and is the
// single source of truth for both the TUI dashboard and the browser
// visualizer.
package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/fsutil"
)

// Sentinel errors for the graph package.
var (
	ErrGraphNotFound = errors.New("graph.json not found")
	ErrCorruptGraph  = errors.New("graph.json contains invalid data")
)

// NodeType defines the kind of graph node.
const (
	NodeTypeHub   = "hub"
	NodeTypeSkill = "skill"
)

// Status constants for node state.
const (
	StatusActive  = "active"
	StatusIdle    = "idle"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusError   = "error"
)

// Node represents a single entity in the graph — either the hub
// orchestrator (index.md) or an installed skill.
type Node struct {
	ID                string     `json:"id"`
	Label             string     `json:"label"`
	Type              string     `json:"type"` // hub | skill
	Category          string     `json:"category,omitempty"`
	Status            string     `json:"status"`
	LastRun           *time.Time `json:"last_run,omitempty"`
	Description       string     `json:"description,omitempty"`
	LastOutputPreview string     `json:"last_output_preview,omitempty"`
	Enabled           *bool      `json:"enabled,omitempty"`
	TasksPending      int        `json:"tasks_pending,omitempty"`
}

// Edge is a directed relationship between two nodes.
type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

// Graph is the top-level document persisted as graph.json.
type Graph struct {
	GeneratedAt time.Time `json:"generated_at"`
	Agency      string    `json:"agency"`
	Nodes       []Node    `json:"nodes"`
	Edges       []Edge    `json:"edges"`
}

// graphAlias is the wire format used for JSON (de)serialization so that
// generated_at is always a human-readable RFC 3339 string instead of Go's
// default time.Time JSON layout.
type graphAlias struct {
	GeneratedAt string `json:"generated_at"`
	Agency      string `json:"agency"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

// MarshalJSON encodes the graph with generated_at as RFC 3339.
func (g Graph) MarshalJSON() ([]byte, error) {
	a := graphAlias{
		GeneratedAt: g.GeneratedAt.UTC().Format(time.RFC3339),
		Agency:      g.Agency,
		Nodes:       g.Nodes,
		Edges:       g.Edges,
	}
	return json.Marshal(a)
}

// UnmarshalJSON decodes the graph, parsing generated_at from RFC 3339.
func (g *Graph) UnmarshalJSON(data []byte) error {
	var a graphAlias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: %v", ErrCorruptGraph, err)
	}
	if a.GeneratedAt != "" {
		t, err := time.Parse(time.RFC3339, a.GeneratedAt)
		if err != nil {
			return fmt.Errorf("%w: bad generated_at: %v", ErrCorruptGraph, err)
		}
		g.GeneratedAt = t
	}
	g.Agency = a.Agency
	g.Nodes = a.Nodes
	g.Edges = a.Edges
	return nil
}

// Validate performs basic structural checks on the graph.
func (g *Graph) Validate() error {
	ids := make(map[string]struct{}, len(g.Nodes))
	hasHub := false
	for _, n := range g.Nodes {
		if n.ID == "" {
			return fmt.Errorf("%w: node with empty id", ErrCorruptGraph)
		}
		if _, dup := ids[n.ID]; dup {
			return fmt.Errorf("%w: duplicate node id %q", ErrCorruptGraph, n.ID)
		}
		ids[n.ID] = struct{}{}
		if n.Type == NodeTypeHub {
			hasHub = true
		}
	}
	if !hasHub {
		return fmt.Errorf("%w: no hub node", ErrCorruptGraph)
	}
	for _, e := range g.Edges {
		if _, ok := ids[e.From]; !ok {
			return fmt.Errorf("%w: edge references unknown source %q", ErrCorruptGraph, e.From)
		}
		if _, ok := ids[e.To]; !ok {
			return fmt.Errorf("%w: edge references unknown target %q", ErrCorruptGraph, e.To)
		}
	}
	return nil
}

// NodeByID returns a pointer to the node with the given id, or nil.
func (g *Graph) NodeByID(id string) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// SkillNodes returns only the skill-type nodes (excluding the hub).
func (g *Graph) SkillNodes() []Node {
	out := make([]Node, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.Type == NodeTypeSkill {
			out = append(out, n)
		}
	}
	return out
}

// WriteFile atomically writes graph.json to the given path.
func WriteFile(path string, g Graph) error {
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}
	return fsutil.AtomicWriteFile(path, append(b, '\n'), 0o644)
}
