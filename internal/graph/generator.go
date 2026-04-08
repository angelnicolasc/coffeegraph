package graph

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/config"
)

// Well-known skill metadata. Custom skills installed by the user that are
// not in this map get sensible defaults (category "default", label = id).
var skillCategories = map[string]string{
	"sales-closer":   "sales",
	"content-engine": "content",
	"lead-nurture":   "leads",
	"life-os":        "personal",
	"creator-stack":  "creator",
}

var skillLabels = map[string]string{
	"sales-closer":   "Sales Closer",
	"content-engine": "Content Engine",
	"lead-nurture":   "Lead Gen Autopilot",
	"life-os":        "Life OS",
	"creator-stack":  "Creator Economy Stack",
}

// Generate builds a Graph by scanning the skills/ directory, merging
// status from the previous graph.json (if any), and incorporating task
// counts from the queue. cfg may be nil for a bare-bones graph.
func Generate(projectRoot string, cfg *config.Config, tasksPending map[string]int) (Graph, error) {
	// Carry over runtime state from the previous graph so that a
	// regeneration triggered by a file-watcher event doesn't wipe
	// status/last_run metadata.
	prev, _ := LoadFile(filepath.Join(projectRoot, "graph.json"))
	prevByID := make(map[string]Node, len(prev.Nodes))
	for _, n := range prev.Nodes {
		prevByID[n.ID] = n
	}

	agency := "my-agency"
	if cfg != nil && cfg.AgencyName != "" {
		agency = cfg.AgencyName
	}

	nodes := []Node{
		{
			ID:          "index",
			Label:       "Brain",
			Type:        NodeTypeHub,
			Status:      StatusActive,
			Description: "Main orchestrator",
		},
	}
	edges := []Edge{}

	skillsDir := filepath.Join(projectRoot, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil && !os.IsNotExist(err) {
		return Graph{}, err
	}

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		name := e.Name()
		skillPath := filepath.Join(skillsDir, name, "SKILL.md")
		if st, serr := os.Stat(skillPath); serr != nil || st.IsDir() {
			continue
		}

		cat := skillCategories[name]
		if cat == "" {
			cat = "default"
		}
		label := skillLabels[name]
		if label == "" {
			// For custom skills: convert kebab-case to Title Case.
			parts := strings.Split(name, "-")
			for i, p := range parts {
				if len(p) > 0 {
					parts[i] = strings.ToUpper(p[:1]) + p[1:]
				}
			}
			label = strings.Join(parts, " ")
		}

		enabled := true
		if cfg != nil {
			enabled = cfg.SkillEnabled(name)
		}

		tpending := 0
		if tasksPending != nil {
			tpending = tasksPending[name]
		}

		// Preserve dynamic state from the previous graph.
		status := StatusIdle
		if prev, ok := prevByID[name]; ok && prev.Status != "" {
			status = prev.Status
		}

		node := Node{
			ID:                name,
			Label:             label,
			Type:              NodeTypeSkill,
			Category:          cat,
			Status:            status,
			LastRun:           prevByID[name].LastRun,
			LastOutputPreview: prevByID[name].LastOutputPreview,
			Enabled:           &enabled,
			TasksPending:      tpending,
		}
		nodes = append(nodes, node)
		edges = append(edges, Edge{From: "index", To: name, Label: "orchestrates"})
	}

	return Graph{
		GeneratedAt: time.Now().UTC(),
		Agency:      agency,
		Nodes:       nodes,
		Edges:       edges,
	}, nil
}

// WriteProjectGraph regenerates and writes graph.json in the project root.
func WriteProjectGraph(projectRoot string, cfg *config.Config, tasksPending map[string]int) error {
	g, err := Generate(projectRoot, cfg, tasksPending)
	if err != nil {
		return err
	}
	return WriteFile(filepath.Join(projectRoot, "graph.json"), g)
}
