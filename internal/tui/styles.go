package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — matches the playbook visual spec.
var (
	colorAmber  = lipgloss.Color("#d4831a")
	colorGreen  = lipgloss.Color("#3a8a1a")
	colorRed    = lipgloss.Color("#bf3a3a")
	colorMuted  = lipgloss.Color("245")
	colorDimmed = lipgloss.Color("240")
	colorBright = lipgloss.Color("230")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDimmed).
			Padding(1, 2)

	mutedStyle = lipgloss.NewStyle().Foreground(colorMuted)

	redStyle = lipgloss.NewStyle().Foreground(colorRed)

	amberStyle = lipgloss.NewStyle().
			Foreground(colorAmber).
			Bold(true)

	greenStyle = lipgloss.NewStyle().Foreground(colorGreen)

	runningBadge = lipgloss.NewStyle().
			Background(colorAmber).
			Foreground(lipgloss.Color("#000")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Underline(true)

	detailBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDimmed).
			Padding(0, 1)
)
