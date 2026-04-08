package main

import (
	"fmt"
	"os"

	"github.com/coffeegraph/coffeegraph/internal/cli"
	"github.com/spf13/cobra"
)

// version is injected at build time via ldflags.
var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "coffeegraph",
		Short:   "CoffeeGraph - AI agency in a folder",
		Version: version,
	}

	root.AddCommand(newInitCmd(), newAddCmd(), newDashboardCmd(), newVisualizeCmd(), newCoffeeCmd(), newQueueCmd(),
		newDeployCmd(), newSuggestCmd(), newBotCmd(), newShareCmd(), newLogCmd(), newDoctorCmd(),
		newEvolveCmd(), newMCPCmd(), newSkillCmd(), newRoastCmd(), newPartyCmd(), newNapCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "init [name]",
		Short:   "Create a new CoffeeGraph project",
		Example: "coffeegraph init my-agency",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunInit(args[0])
		},
	}
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add [skill]",
		Short:   "Install a skill from templates",
		Example: "coffeegraph add sales-closer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunAdd(args[0])
		},
	}
}

func newDashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "dashboard",
		Short:   "Open live terminal dashboard",
		Example: "coffeegraph dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDashboard()
		},
	}
}

func newVisualizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "visualize",
		Short:   "Open interactive graph in browser",
		Example: "coffeegraph visualize",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunVisualize()
		},
	}
}

func newCoffeeCmd() *cobra.Command {
	var urgent bool
	var chill bool
	cmd := &cobra.Command{
		Use:     "coffee",
		Short:   "Execute queued tasks",
		Example: "coffeegraph coffee --urgent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunCoffee(urgent, chill)
		},
	}
	cmd.Flags().BoolVar(&urgent, "urgent", false, "run in meme high-pressure visual mode")
	cmd.Flags().BoolVar(&chill, "chill", false, "run in calm visual mode")
	return cmd
}

func newQueueCmd() *cobra.Command {
	var skill, task string
	var priority int
	var urgent bool
	queueCmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage task queue",
	}
	queueAdd := &cobra.Command{
		Use:     "add",
		Short:   "Add task to queue",
		Example: "coffeegraph queue add --skill sales-closer --task \"Follow up this week's leads\" --priority 1 --urgent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueAdd(skill, task, priority, urgent)
		},
	}
	queueAdd.Flags().StringVar(&skill, "skill", "", "skill name")
	queueAdd.Flags().StringVar(&task, "task", "", "task description")
	queueAdd.Flags().IntVar(&priority, "priority", 0, "priority 1-5 (default 3)")
	queueAdd.Flags().BoolVar(&urgent, "urgent", false, "tag this task as urgent")
	queueCmd.AddCommand(queueAdd)
	queueCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List queued tasks",
		Example: "coffeegraph queue list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueList()
		},
	})
	queueCmd.AddCommand(&cobra.Command{
		Use:     "clear",
		Short:   "Clear queue",
		Example: "coffeegraph queue clear",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunQueueClear()
		},
	})
	return queueCmd
}

func newDeployCmd() *cobra.Command {
	deploy := &cobra.Command{Use: "deploy", Short: "Export skills to external runtimes"}
	deploy.AddCommand(&cobra.Command{
		Use:     "openclaw",
		Short:   "Export to OpenClaw SOUL.md format",
		Example: "coffeegraph deploy openclaw",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDeployOpenClaw()
		},
	})
	deploy.AddCommand(&cobra.Command{
		Use:     "hermes",
		Short:   "Export to Hermes AGENT.md format",
		Example: "coffeegraph deploy hermes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDeployHermes()
		},
	})
	return deploy
}

func newSuggestCmd() *cobra.Command {
	var obsidian string
	var deep bool
	cmd := &cobra.Command{
		Use:     "suggest",
		Short:   "Suggest new skills from project context",
		Example: "coffeegraph suggest --from-obsidian /path/to/vault --deep",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSuggest(obsidian, deep)
		},
	}
	cmd.Flags().StringVar(&obsidian, "from-obsidian", "", "path to Obsidian vault or markdown folder")
	cmd.Flags().BoolVar(&deep, "deep", false, "read markdown files recursively")
	return cmd
}

func newBotCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "bot",
		Short:   "Run Telegram bot adapter",
		Example: "coffeegraph bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunBot()
		},
	}
}

func newShareCmd() *cobra.Command {
	var pub bool
	cmd := &cobra.Command{
		Use:     "share [job-id]",
		Short:   "Share a completed job output",
		Example: "coffeegraph share 123abc --public",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := ""
			if len(args) == 1 {
				jobID = args[0]
			}
			return cli.RunShare(jobID, pub)
		},
	}
	cmd.Flags().BoolVar(&pub, "public", false, "create a public gist instead of secret gist")
	return cmd
}

func newLogCmd() *cobra.Command {
	var pretty bool
	var last int
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "List completed job logs",
		Example: "coffeegraph log --pretty --last 1",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunLog(pretty, last)
		},
	}
	cmd.Flags().BoolVar(&pretty, "pretty", false, "render logs as receipt-style cards")
	cmd.Flags().IntVar(&last, "last", 0, "show only the last N logs")
	return cmd
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Check dependencies and project health",
		Example: "coffeegraph doctor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDoctor()
		},
	}
}

func newEvolveCmd() *cobra.Command {
	var auto bool
	cmd := &cobra.Command{
		Use:     "evolve [skill]",
		Short:   "Suggest SKILL.md improvements based on recent logs",
		Example: "coffeegraph evolve sales-closer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunEvolve(args[0], auto)
		},
	}
	cmd.Flags().BoolVar(&auto, "auto", false, "apply automatically without confirmation")
	return cmd
}

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "mcp",
		Short:   "Run CoffeeGraph MCP server over stdio",
		Example: "coffeegraph mcp",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunMCP()
		},
	}
}

func newSkillCmd() *cobra.Command {
	skill := &cobra.Command{Use: "skill", Short: "Community skill registry commands"}
	skill.AddCommand(&cobra.Command{
		Use:     "install [name]",
		Short:   "Install a skill from community registry",
		Example: "coffeegraph skill install sales-closer-viral",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSkillInstall(args[0])
		},
	})
	skill.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List community skills",
		Example: "coffeegraph skill list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSkillList()
		},
	})
	skill.AddCommand(&cobra.Command{
		Use:     "publish",
		Short:   "Prepare publication to community registry",
		Example: "coffeegraph skill publish",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSkillPublish()
		},
	})
	return skill
}

func newRoastCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "roast",
		Short:   "Roast your business context like a tough VC",
		Example: "coffeegraph roast",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunRoast()
		},
	}
}

func newPartyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "party",
		Short:   "Run a round-robin conversation between skills",
		Example: "coffeegraph party",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunParty()
		},
	}
}

func newNapCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "nap",
		Short:   "Pause with a playful coffee-cup animation",
		Example: "coffeegraph nap",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunNap()
		},
	}
}
