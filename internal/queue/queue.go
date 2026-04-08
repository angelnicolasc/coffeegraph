// Package queue manages the .coffee/queue.json task backlog.
package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Item is a single entry in .coffee/queue.json.
type Item struct {
	ID       string `json:"id"`
	Skill    string `json:"skill"`
	Task     string `json:"task"`
	Priority int    `json:"priority"`
	Data     string `json:"data,omitempty"`
}

// Path returns the absolute path to queue.json for the given project.
func Path(projectRoot string) string {
	return filepath.Join(projectRoot, ".coffee", "queue.json")
}

// Read loads the queue sorted by priority (lower number = more urgent),
// with stable ordering by ID within the same priority level.
func Read(projectRoot string) ([]Item, error) {
	p := Path(projectRoot)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("read queue: %w", err)
	}
	content := strings.TrimSpace(string(b))
	if content == "" || content == "[]" {
		return []Item{}, nil
	}
	var items []Item
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, fmt.Errorf("parse queue: %w", err)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority == items[j].Priority {
			return items[i].ID < items[j].ID
		}
		return items[i].Priority < items[j].Priority
	})
	return items, nil
}

// Write persists the complete queue to disk. A nil slice is written as [].
func Write(projectRoot string, items []Item) error {
	if items == nil {
		items = []Item{}
	}
	p := Path(projectRoot)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create queue dir: %w", err)
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal queue: %w", err)
	}
	return os.WriteFile(p, append(b, '\n'), 0644)
}

// Add appends an item and returns the position (1-based) among items
// for the same skill, and the total queue size.
func Add(projectRoot string, it Item) (position int, total int, err error) {
	if it.ID == "" {
		it.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if it.Priority < 1 || it.Priority > 5 {
		it.Priority = 3
	}
	items, err := Read(projectRoot)
	if err != nil {
		return 0, 0, err
	}
	items = append(items, it)
	sameSkill := 0
	for _, x := range items {
		if x.Skill == it.Skill {
			sameSkill++
		}
	}
	if err := Write(projectRoot, items); err != nil {
		return 0, 0, err
	}
	return sameSkill, len(items), nil
}

// Remove deletes the item with the given ID from the queue.
func Remove(projectRoot string, id string) error {
	items, err := Read(projectRoot)
	if err != nil {
		return err
	}
	out := make([]Item, 0, len(items))
	for _, x := range items {
		if x.ID != id {
			out = append(out, x)
		}
	}
	return Write(projectRoot, out)
}

// Set replaces the entire queue (typically after executing tasks).
func Set(projectRoot string, items []Item) error {
	return Write(projectRoot, items)
}

// CountBySkill counts pending tasks grouped by skill name.
func CountBySkill(projectRoot string) (map[string]int, error) {
	items, err := Read(projectRoot)
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(items))
	for _, it := range items {
		m[it.Skill]++
	}
	return m, nil
}
