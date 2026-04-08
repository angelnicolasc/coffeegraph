package graph

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/fsnotify/fsnotify"
)

// WatchSkills observes the skills/ directory (and its subdirectories) for
// changes and regenerates graph.json. It coalesces rapid-fire events with
// a 200ms debounce window to avoid redundant regenerations. The returned
// cancel function stops the watcher. The onRegenerate callback is invoked
// after every successful regeneration (e.g. to broadcast via WebSocket).
func WatchSkills(
	ctx context.Context,
	projectRoot string,
	loadConfig func() (*config.Config, error),
	countTasks func() map[string]int,
	onRegenerate func(),
) (cancel func(), err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	skillsDir := filepath.Join(projectRoot, "skills")
	_ = os.MkdirAll(skillsDir, 0755) // ensure it exists
	if err := addRecursive(watcher, skillsDir); err != nil {
		_ = watcher.Close()
		return nil, err
	}

	ctx, stopFunc := context.WithCancel(ctx)

	var debounceOnce sync.Once
	var debounceTimer *time.Timer

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
					continue
				}
				// If a new directory was created, watch it too.
				if ev.Op&fsnotify.Create != 0 {
					if st, serr := os.Stat(ev.Name); serr == nil && st.IsDir() {
						_ = addRecursive(watcher, ev.Name)
					}
				}
				// Debounce: reset the timer on every event.
				debounceOnce.Do(func() {
					debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
						regenerate(projectRoot, loadConfig, countTasks, onRegenerate)
						debounceOnce = sync.Once{} // allow next batch
					})
				})
				if debounceTimer != nil {
					debounceTimer.Reset(200 * time.Millisecond)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("coffeegraph watcher: %v", err)
			}
		}
	}()

	return stopFunc, nil
}

// addRecursive walks dir and adds every subdirectory to the watcher.
func addRecursive(w *fsnotify.Watcher, dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			return w.Add(path)
		}
		return nil
	})
}

func regenerate(
	projectRoot string,
	loadConfig func() (*config.Config, error),
	countTasks func() map[string]int,
	onRegenerate func(),
) {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("coffeegraph watcher: config: %v", err)
		return
	}
	if err := WriteProjectGraph(projectRoot, cfg, countTasks()); err != nil {
		log.Printf("coffeegraph watcher: regenerate: %v", err)
		return
	}
	if onRegenerate != nil {
		onRegenerate()
	}
}
