package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateOK(t *testing.T) {
	g := &Graph{
		Nodes: []Node{
			{ID: "index", Type: NodeTypeHub, Status: StatusActive},
			{ID: "sales-closer", Type: NodeTypeSkill, Status: StatusIdle},
		},
		Edges: []Edge{{From: "index", To: "sales-closer"}},
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateNoHub(t *testing.T) {
	g := &Graph{Nodes: []Node{{ID: "x", Type: NodeTypeSkill}}}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error for missing hub")
	}
}

func TestValidateDuplicateID(t *testing.T) {
	g := &Graph{Nodes: []Node{
		{ID: "index", Type: NodeTypeHub},
		{ID: "index", Type: NodeTypeSkill},
	}}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestValidateEmptyID(t *testing.T) {
	g := &Graph{Nodes: []Node{{ID: "", Type: NodeTypeHub}}}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestValidateEdgeBadSource(t *testing.T) {
	g := &Graph{
		Nodes: []Node{{ID: "index", Type: NodeTypeHub}},
		Edges: []Edge{{From: "missing", To: "index"}},
	}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error for edge with unknown source")
	}
}

func TestValidateEdgeBadTarget(t *testing.T) {
	g := &Graph{
		Nodes: []Node{{ID: "index", Type: NodeTypeHub}},
		Edges: []Edge{{From: "index", To: "missing"}},
	}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error for edge with unknown target")
	}
}

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := Graph{
		GeneratedAt: now,
		Agency:      "test-agency",
		Nodes: []Node{
			{ID: "index", Label: "Brain", Type: NodeTypeHub, Status: StatusActive},
		},
		Edges: []Edge{},
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded Graph
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !decoded.GeneratedAt.Equal(original.GeneratedAt) {
		t.Fatalf("GeneratedAt mismatch: got %v, want %v", decoded.GeneratedAt, original.GeneratedAt)
	}
	if decoded.Agency != "test-agency" {
		t.Fatalf("Agency = %q, want test-agency", decoded.Agency)
	}
	if len(decoded.Nodes) != 1 || decoded.Nodes[0].ID != "index" {
		t.Fatalf("unexpected nodes: %v", decoded.Nodes)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	var g Graph
	if err := json.Unmarshal([]byte(`{invalid`), &g); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUnmarshalBadTimestamp(t *testing.T) {
	var g Graph
	if err := json.Unmarshal([]byte(`{"generated_at":"not-a-date","agency":"x","nodes":[],"edges":[]}`), &g); err == nil {
		t.Fatal("expected error for bad timestamp")
	}
}

func TestNodeByID(t *testing.T) {
	g := &Graph{Nodes: []Node{
		{ID: "index", Type: NodeTypeHub},
		{ID: "sales", Type: NodeTypeSkill},
	}}
	if n := g.NodeByID("sales"); n == nil || n.ID != "sales" {
		t.Fatal("NodeByID should find 'sales'")
	}
	if n := g.NodeByID("missing"); n != nil {
		t.Fatal("NodeByID should return nil for missing")
	}
}

func TestSkillNodes(t *testing.T) {
	g := &Graph{Nodes: []Node{
		{ID: "index", Type: NodeTypeHub},
		{ID: "a", Type: NodeTypeSkill},
		{ID: "b", Type: NodeTypeSkill},
	}}
	skills := g.SkillNodes()
	if len(skills) != 2 {
		t.Fatalf("expected 2 skill nodes, got %d", len(skills))
	}
}

func TestWriteFileAndLoadFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "graph.json")
	g := Graph{
		GeneratedAt: time.Now().UTC().Truncate(time.Second),
		Agency:      "roundtrip",
		Nodes:       []Node{{ID: "index", Type: NodeTypeHub, Status: StatusActive}},
		Edges:       []Edge{},
	}
	if err := WriteFile(p, g); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	loaded, err := LoadFile(p)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if loaded.Agency != "roundtrip" {
		t.Fatalf("Agency = %q, want roundtrip", loaded.Agency)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestGenerateWithSkills(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills", "my-skill")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# Skill: my-skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g, err := Generate(root, nil, map[string]int{"my-skill": 2})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(g.Nodes) != 2 { // hub + skill
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
	skill := g.NodeByID("my-skill")
	if skill == nil {
		t.Fatal("expected my-skill node")
	}
	if skill.TasksPending != 2 {
		t.Fatalf("TasksPending = %d, want 2", skill.TasksPending)
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
}
