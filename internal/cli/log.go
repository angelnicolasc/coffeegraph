package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coffeegraph/coffeegraph/internal/logs"
	"github.com/coffeegraph/coffeegraph/internal/project"
)

// RunLog prints historical execution logs.
func RunLog(pretty bool, last int) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	all, err := logs.List(root)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("No completed logs found in .coffee/logs/")
		return nil
	}
	if last > 0 && len(all) > last {
		all = all[:last]
	}
	if !pretty {
		for _, e := range all {
			fmt.Printf("%s  %-16s %-6s %s\n", e.FinishedAt.Format("2006-01-02 15:04:05"), e.Skill, e.Status, e.ID)
		}
		return nil
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("#b98f5d"))
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7a5230"))
	statusDone := lipgloss.NewStyle().Foreground(lipgloss.Color("#2e8a3b")).Bold(true)
	statusErr := lipgloss.NewStyle().Foreground(lipgloss.Color("#c23b3b")).Bold(true)

	for i, e := range all {
		status := statusDone.Render(e.Status)
		if strings.EqualFold(e.Status, "ERROR") {
			status = statusErr.Render(e.Status)
		}
		body := fmt.Sprintf("%s %s\nTask: %s\nDuration: %dms\nWhen: %s\nID: %s",
			iconForSkill(e.Skill), title.Render(e.Skill), nonEmpty(e.Task, "(task unavailable)"), e.DurationMs, e.FinishedAt.Local().Format("2006-01-02 15:04:05"), e.ID)
		if last == 1 && i == 0 {
			body += "\n\nOutput:\n" + e.Result
		} else {
			body += "\nOutput: " + truncateSimple(oneLine(e.Result), 80)
		}
		body += "\nStatus: " + status + "\nFile: " + filepath.Base(e.File)
		fmt.Println(box.Render(body))
	}
	return nil
}

func iconForSkill(s string) string {
	ls := strings.ToLower(s)
	switch {
	case strings.Contains(ls, "sales"):
		return "💼"
	case strings.Contains(ls, "content"):
		return "🧠"
	case strings.Contains(ls, "lead"):
		return "📬"
	default:
		return "☕"
	}
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func truncateSimple(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
