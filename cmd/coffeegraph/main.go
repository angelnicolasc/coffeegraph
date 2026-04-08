package main

import (
	"fmt"
	"os"

	"github.com/coffeegraph/coffeegraph/internal/cli"
	"github.com/spf13/cobra"
)

// version is injected at build time via ldflags:
//
//	go build -ldflags "-X main.version=0.1.0" ./cmd/coffeegraph
var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "coffeegraph",
		Short:   "CoffeeGraph — AI agency in a folder",
		Version: version,
	}

	// coffeegraph init <name>
	root.AddCommand(&cobra.Command{
		Use:   "init [name]",
		Short: "Create a new CoffeeGraph project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunInit(args[0])
		},
	})

	// coffeegraph add <skill>
	root.AddCommand(&cobra.Command{
		Use:   "add [skill]",
		Short: "Install a skill from templates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunAdd(args[0])
		},
	})

	// coffeegraph dashboard
	root.AddCommand(&cobra.Command{
		Use:   "dashboard",
		Short: "TUI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDashboard()
		},
	})

	// coffeegraph visualize
	root.AddCommand(&cobra.Command{
		Use:   "visualize",
		Short: "Interactive graph in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunVisualize()
		},
	})

	// coffeegraph coffee
	root.AddCommand(&cobra.Command{
		Use:   "coffee",
		Short: "Coffee mode — execute queued tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunCoffee()
		},
	})

	// coffeegraph queue {add,list,clear}
	queueCmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage the task queue",
	}
	var skill, task string
	var priority int
	queueAdd := &cobra.Command{
		Use:   "add",
		Short: "Add a task to the queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueAdd(skill, task, priority)
		},
	}
	queueAdd.Flags().StringVar(&skill, "skill", "", "skill name")
	queueAdd.Flags().StringVar(&task, "task", "", "task description")
	queueAdd.Flags().IntVar(&priority, "priority", 0, "priority 1-5 (default 3)")
	queueCmd.AddCommand(queueAdd)
	queueCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List queued tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueList()
		},
	})
	queueCmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueClear()
		},
	})
	root.AddCommand(queueCmd)

	// coffeegraph deploy {openclaw,hermes}
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Export skills to external platforms",
	}
	deployCmd.AddCommand(&cobra.Command{
		Use:   "openclaw",
		Short: "OpenClaw format (SOUL.md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDeployOpenClaw()
		},
	})
	deployCmd.AddCommand(&cobra.Command{
		Use:   "hermes",
		Short: "Hermes Agent format (AGENT.md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDeployHermes()
		},
	})
	root.AddCommand(deployCmd)

	// coffeegraph suggest
	root.AddCommand(&cobra.Command{
		Use:   "suggest",
		Short: "Ask Claude to suggest new skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSuggest()
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
