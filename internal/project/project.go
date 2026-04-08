// Package project provides utilities for locating the CoffeeGraph
// project root directory by walking up from the current working directory.
package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRoot searches for config.yaml starting from dir (or the current
// working directory if dir is empty), walking upward until the filesystem
// root. Returns the directory containing config.yaml, or an error if not
// found.
func FindRoot(dir string) (string, error) {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("detect working directory: %w", err)
		}
		dir = wd
	}
	dir = filepath.Clean(dir)
	for {
		cfg := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(cfg); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("not inside a CoffeeGraph project. Run first: coffeegraph init <name>")
}
