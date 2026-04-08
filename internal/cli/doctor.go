package cli

import (
	"fmt"
	"path/filepath"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/doctor"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunDoctor executes environment and project checks.
func RunDoctor() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	checks, err := doctor.Run(root, cfg)
	if err != nil {
		return err
	}
	fmt.Println("CoffeeGraph doctor report")
	allOK := true
	for _, c := range checks {
		status := "OK"
		if !c.OK {
			status = "WARN"
			allOK = false
		}
		fmt.Printf("[%s] %s - %s\n", status, c.Name, c.Detail)
	}
	if allOK {
		fmt.Println("All checks passed.")
	}
	return nil
}
