package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSkillFileOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := "# Skill: test-skill\n\n## Identity\nYou are a test.\n\n## Workflow\n1. Step\n\n## Output Format\nText\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSkillFile(path); err != nil {
		t.Fatalf("ValidateSkillFile() error = %v", err)
	}
}

func TestValidateSkillFileNotFound(t *testing.T) {
	err := ValidateSkillFile(filepath.Join(t.TempDir(), "SKILL.md"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSkillFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	err := ValidateSkillFile(path)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error, got: %v", err)
	}
}

func TestValidateSkillFileMissingIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := "Some random content without any headers or sections."
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	err := ValidateSkillFile(path)
	if err == nil {
		t.Fatal("expected error for missing structure")
	}
	if !strings.Contains(err.Error(), "missing required structure") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSkillFileMissingWorkflow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := "# Skill: test\n\n## Identity\nYou are a test.\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	err := ValidateSkillFile(path)
	if err == nil {
		t.Fatal("expected error for missing workflow/output sections")
	}
	if !strings.Contains(err.Error(), "missing workflow or output") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSkillFileCaseInsensitive(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"lowercase headers", "# skill: test\n\n## identity\ntest\n\n## workflow\nstep\n\n## output format\ntext\n", false},
		{"mixed case", "# Skill: test\n\n## Identity\ntest\n\n## Workflow\nstep\n\n## Output Format\ntext\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "SKILL.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			err := ValidateSkillFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}
