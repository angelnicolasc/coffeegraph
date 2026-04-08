package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

// RunCoffee executes queued tasks with optional themed output.
func RunCoffee(urgent, chill bool) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	mode := runner.ModeNormal
	if urgent {
		mode = runner.ModeUrgent
	}
	if chill {
		mode = runner.ModeChill
	}

	ctx := context.Background()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	go func() {
		<-sig
		fmt.Println("\nGraceful shutdown requested. Current task will finish, then CoffeeGraph will stop.")
	}()

	engine := &runner.Engine{Root: root, Cfg: cfg}
	fmt.Println("Coffee mode running...")
	summary, err := engine.ExecutePending(ctx, cfg.Coffee.MaxTasksPerRun, mode)
	if err != nil {
		return err
	}
	if summary.Completed == 0 && len(summary.Errors) == 0 {
		fmt.Println("Queue is empty. Add tasks with: coffeegraph queue add")
		return nil
	}
	fmt.Printf("Done. %d task(s) executed.\n", summary.Completed)
	for _, p := range summary.Paths {
		fmt.Printf("  - %s\n", p)
	}
	if len(summary.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range summary.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
	return nil
}
