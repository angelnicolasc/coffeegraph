// Package logs manages CoffeeGraph execution log files.
package logs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/fsutil"
)

// Entry is a structured representation of a job log.
type Entry struct {
	ID         string
	File       string
	Skill      string
	Task       string
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	DurationMs int64
	Result     string
}

// Dir returns the log directory for a project.
func Dir(root string) string {
	return filepath.Join(root, ".coffee", "logs")
}

// Write creates a markdown log file with metadata front matter.
func Write(root string, e Entry) (string, error) {
	if e.ID == "" {
		e.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if e.Status == "" {
		e.Status = "DONE"
	}
	if e.StartedAt.IsZero() {
		e.StartedAt = time.Now().UTC()
	}
	if e.FinishedAt.IsZero() {
		e.FinishedAt = e.StartedAt
	}
	if e.DurationMs == 0 {
		e.DurationMs = e.FinishedAt.Sub(e.StartedAt).Milliseconds()
	}
	ts := e.FinishedAt.UTC().Format("2006-01-02-15-04-05")
	name := fmt.Sprintf("%s-%s-%s.md", ts, e.Skill, e.ID)
	path := filepath.Join(Dir(root), name)
	body := fmt.Sprintf(`---
id: %s
skill: %s
task: %s
status: %s
started_at: %s
finished_at: %s
duration_ms: %d
---
%s
`, esc(e.ID), esc(e.Skill), esc(e.Task), esc(e.Status), e.StartedAt.UTC().Format(time.RFC3339), e.FinishedAt.UTC().Format(time.RFC3339), e.DurationMs, strings.TrimSpace(e.Result))
	if err := fsutil.AtomicWriteFile(path, []byte(body), 0644); err != nil {
		return "", fmt.Errorf("write log: %w", err)
	}
	return path, nil
}

// List returns logs in reverse chronological order.
func List(root string) ([]Entry, error) {
	matches, err := filepath.Glob(filepath.Join(Dir(root), "*.md"))
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}
	entries := make([]Entry, 0, len(matches))
	for _, p := range matches {
		e, err := ReadFile(p)
		if err != nil {
			continue
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FinishedAt.After(entries[j].FinishedAt)
	})
	return entries, nil
}

// FindByID finds a log by ID; empty id returns the latest.
func FindByID(root, id string) (*Entry, error) {
	all, err := List(root)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no logs found")
	}
	if strings.TrimSpace(id) == "" {
		return &all[0], nil
	}
	for _, e := range all {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("job-id %q not found", id)
}

// ReadFile parses a log file.
func ReadFile(path string) (Entry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("read log: %w", err)
	}
	txt := string(b)
	e := Entry{
		File:       path,
		FinishedAt: fileTime(path),
		Status:     "DONE",
	}
	if strings.HasPrefix(txt, "---\n") {
		parts := strings.SplitN(txt, "\n---\n", 2)
		if len(parts) == 2 {
			meta := strings.TrimPrefix(parts[0], "---\n")
			sc := bufio.NewScanner(strings.NewReader(meta))
			for sc.Scan() {
				line := sc.Text()
				k, v, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				key := strings.TrimSpace(k)
				val := strings.TrimSpace(v)
				switch key {
				case "id":
					e.ID = val
				case "skill":
					e.Skill = val
				case "task":
					e.Task = val
				case "status":
					e.Status = val
				case "started_at":
					if t, err := time.Parse(time.RFC3339, val); err == nil {
						e.StartedAt = t
					}
				case "finished_at":
					if t, err := time.Parse(time.RFC3339, val); err == nil {
						e.FinishedAt = t
					}
				case "duration_ms":
					if n, err := strconv.ParseInt(val, 10, 64); err == nil {
						e.DurationMs = n
					}
				}
			}
			e.Result = strings.TrimSpace(parts[1])
		}
	}
	if e.ID == "" {
		e.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if e.Skill == "" {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		parts := strings.Split(base, "-")
		if len(parts) >= 2 {
			e.Skill = parts[len(parts)-2]
		}
	}
	if e.DurationMs == 0 && !e.StartedAt.IsZero() && !e.FinishedAt.IsZero() {
		e.DurationMs = e.FinishedAt.Sub(e.StartedAt).Milliseconds()
	}
	if e.Result == "" && !strings.HasPrefix(txt, "---\n") {
		e.Result = strings.TrimSpace(txt)
	}
	return e, nil
}

func fileTime(path string) time.Time {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Now().UTC()
	}
	return fi.ModTime().UTC()
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
