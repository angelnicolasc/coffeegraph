package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/mcp"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

// RunMCP starts MCP server mode over stdio.
func RunMCP() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "CoffeeGraph MCP server running on stdio")
	engine := &runner.Engine{Root: root, Cfg: cfg}
	return mcp.Run(context.Background(), root, engine, os.Stdin, os.Stdout)
}
