package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// LoadFile reads and parses graph.json from the given path.
// Returns ErrGraphNotFound if the file does not exist.
func LoadFile(path string) (Graph, error) {
	var g Graph
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return g, fmt.Errorf("%w: %s", ErrGraphNotFound, path)
		}
		return g, fmt.Errorf("read graph: %w", err)
	}
	if err := json.Unmarshal(b, &g); err != nil {
		return g, fmt.Errorf("parse graph %s: %w", path, err)
	}
	return g, nil
}
