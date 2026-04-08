package cli

import (
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/tui"
)

// RunDashboard launches the TUI dashboard.
func RunDashboard() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	return tui.Run(root)
}
