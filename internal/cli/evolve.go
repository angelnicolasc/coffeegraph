package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/evolve"
	"github.com/coffeegraph/coffeegraph/internal/fsutil"
)

// RunEvolve proposes improvements to a skill's SKILL.md.
func RunEvolve(skill string, auto bool) error {
	root, err := evolve.DetectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	path, proposed, err := evolve.Suggest(context.Background(), root, cfg, skill)
	if err != nil {
		return err
	}
	tmp := filepath.Join(root, ".coffee", "tmp-evolve.md")
	if err := fsutil.AtomicWriteFile(tmp, []byte(proposed), 0o644); err != nil {
		return err
	}
	defer os.Remove(tmp)

	cmd := exec.Command("git", "diff", "--no-index", "--", path, tmp)
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
	if !auto {
		fmt.Print("Apply this update? (y/N): ")
		rd := bufio.NewReader(os.Stdin)
		line, _ := rd.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(line)) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}
	if err := evolve.Apply(path, proposed); err != nil {
		return err
	}
	fmt.Println("Applied evolution to", path)
	return nil
}
