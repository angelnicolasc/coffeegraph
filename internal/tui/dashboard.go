package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/coffeegraph/coffeegraph/internal/graph"
)

// View renders the full dashboard frame.
func (m model) View() string {
	if m.err != nil {
		return redStyle.Render("Error: "+m.err.Error()) + "\n"
	}

	title := titleStyle.Render("☕ CoffeeGraph Dashboard — " + m.g.Agency)

	var b strings.Builder
	b.WriteString(headerStyle.Render("ACTIVE SKILLS") + "                      " + headerStyle.Render("LAST OUTPUT") + "\n")

	skills := m.g.SkillNodes()
	for i, n := range skills {
		selected := i == m.cursor
		line := formatSkillLine(n, selected)
		b.WriteString(line + "\n")
	}
	if len(skills) == 0 {
		b.WriteString(mutedStyle.Render("  No skills installed yet. Run: coffeegraph add sales-closer") + "\n")
	}

	// Detail panel: show full last output for selected skill.
	b.WriteString("\n")
	if m.cursor < len(skills) {
		sel := skills[m.cursor]
		detail := renderDetail(sel)
		b.WriteString(detail)
	}

	// Footer.
	b.WriteString("\n")
	footer := fmt.Sprintf(
		"QUEUE: %d tasks   COFFEE: off",
		m.qCount,
	)
	footerKeys := "[c] coffee  [a] add task  [v] visualize  [r] refresh  [q] quit"

	// Input mode overlay.
	var inputLine string
	switch m.mode {
	case modeAddSkill:
		inputLine = "Skill: " + m.input.View()
	case modeAddTask:
		inputLine = fmt.Sprintf("Task for %s: %s", amberStyle.Render(m.addSkill), m.input.View())
	}

	w := 78
	if m.width > 24 {
		w = min(78, m.width-4)
	}

	content := title + "\n\n" + b.String() + "\n" +
		mutedStyle.Render(footer) + "\n" +
		mutedStyle.Render(footerKeys)

	if inputLine != "" {
		content += "\n\n" + inputLine
	}

	box := borderStyle.Width(w).Render(content)
	return lipgloss.NewStyle().Padding(1, 2).Render(box)
}

func formatSkillLine(n graph.Node, selected bool) string {
	// Status indicator.
	var indicator string
	var statusLabel string
	switch n.Status {
	case graph.StatusRunning:
		indicator = "◉"
		statusLabel = runningBadge.Render("RUNNING ✦")
	case graph.StatusError:
		indicator = "✗"
		statusLabel = redStyle.Render("error")
	case graph.StatusDone:
		indicator = "✓"
		statusLabel = greenStyle.Render("done")
	default: // idle
		indicator = "○"
		statusLabel = mutedStyle.Render("idle")
	}

	// Label.
	nameStr := n.Label
	if selected {
		indicator = selectedStyle.Render("▸ " + indicator)
		nameStr = selectedStyle.Render(nameStr)
	} else {
		indicator = "  " + indicator
	}

	// Pending tasks count.
	pendStr := ""
	if n.TasksPending > 0 {
		pendStr = mutedStyle.Render(fmt.Sprintf(" (%d pending)", n.TasksPending))
	}

	// Last output preview.
	prev := n.LastOutputPreview
	if len(prev) > 40 {
		prev = prev[:37] + "..."
	}
	if prev == "" {
		prev = "—"
	}

	// Time since last run.
	ago := ""
	if n.LastRun != nil {
		d := time.Since(*n.LastRun).Truncate(time.Minute)
		switch {
		case d < time.Minute:
			ago = "just now"
		case d < time.Hour:
			ago = fmt.Sprintf("%dm ago", int(d.Minutes()))
		case d < 24*time.Hour:
			ago = fmt.Sprintf("%dh ago", int(d.Hours()))
		default:
			ago = fmt.Sprintf("%dd ago", int(d.Hours()/24))
		}
		ago = mutedStyle.Render(ago)
	}

	return fmt.Sprintf("%s %-22s %-14s%s  %-42s %s",
		indicator, nameStr, statusLabel, pendStr, mutedStyle.Render(prev), ago)
}

func renderDetail(n graph.Node) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("▾ "+n.Label+" — Detail") + "\n")

	preview := n.LastOutputPreview
	if preview == "" {
		preview = "No recent output."
	}
	// Wrap to ~70 chars.
	lines := wrapText(preview, 70)
	for _, line := range lines {
		b.WriteString("  " + line + "\n")
	}
	return detailBorder.Render(b.String())
}

// wrapText breaks s into lines of at most width characters.
func wrapText(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if len(current)+1+len(w) > width {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}
