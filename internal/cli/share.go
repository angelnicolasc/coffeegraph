package cli

import (
	"fmt"
	"path/filepath"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/share"
)

// RunShare publishes a completed job output.
func RunShare(jobID string, public bool) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	entry, err := logs.FindByID(root, jobID)
	if err != nil {
		return err
	}
	res, err := share.Publish(*entry, cfg.GitHubToken(), public, root)
	if err != nil {
		return err
	}
	if res.Fallback {
		fmt.Println("GitHub token not configured or gist failed. Generated local share page:")
		fmt.Println(res.Local)
		return nil
	}
	fmt.Println("Share URL:")
	fmt.Println(res.URL)
	return nil
}
