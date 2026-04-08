package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coffeegraph/coffeegraph/internal/graph"
	"github.com/coffeegraph/coffeegraph/internal/queue"
)

// --- Messages ----------------------------------------------------------

type graphMsg struct {
	g   graph.Graph
	err error
}

type tickMsg time.Time

type addedTaskMsg struct {
	skill string
	err   error
}

// --- Model -------------------------------------------------------------

type inputMode int

const (
	modeNormal inputMode = iota
	modeAddSkill
	modeAddTask
)

type model struct {
	root     string
	g        graph.Graph
	qCount   int
	err      error
	width    int
	height   int
	cursor   int    // selected skill index (0-based among skill nodes)
	mode     inputMode
	input    textinput.Model
	addSkill string // skill name captured in modeAddSkill
}

// Run starts the Bubble Tea TUI dashboard.
func Run(projectRoot string) error {
	ti := textinput.New()
	ti.Placeholder = "type here..."
	ti.CharLimit = 256

	m := model{
		root:  projectRoot,
		input: ti,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// --- Init --------------------------------------------------------------

func (m model) Init() tea.Cmd {
	return tea.Batch(loadGraph(m.root), tick())
}

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func loadGraph(root string) tea.Cmd {
	return func() tea.Msg {
		g, err := graph.LoadFile(filepath.Join(root, "graph.json"))
		return graphMsg{g, err}
	}
}

// --- Update ------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case graphMsg:
		m.g = msg.g
		m.err = msg.err
		items, _ := queue.Read(m.root)
		m.qCount = len(items)
		// Clamp cursor to valid range.
		skills := m.g.SkillNodes()
		if m.cursor >= len(skills) && len(skills) > 0 {
			m.cursor = len(skills) - 1
		}
		return m, nil

	case tickMsg:
		return m, tea.Batch(loadGraph(m.root), tick())

	case addedTaskMsg:
		m.mode = modeNormal
		m.input.Reset()
		if msg.err != nil {
			m.err = msg.err
		}
		return m, loadGraph(m.root)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass through to text input when in input mode.
	if m.mode != modeNormal {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Escape always returns to normal mode.
	if key == "esc" {
		m.mode = modeNormal
		m.input.Reset()
		return m, nil
	}

	// Input modes: capture text.
	if m.mode == modeAddSkill {
		if key == "enter" {
			m.addSkill = strings.TrimSpace(m.input.Value())
			if m.addSkill == "" {
				m.mode = modeNormal
				return m, nil
			}
			m.input.Reset()
			m.input.Placeholder = "task description..."
			m.mode = modeAddTask
			m.input.Focus()
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if m.mode == modeAddTask {
		if key == "enter" {
			taskDesc := strings.TrimSpace(m.input.Value())
			if taskDesc == "" {
				m.mode = modeNormal
				return m, nil
			}
			skill := m.addSkill
			root := m.root
			m.mode = modeNormal
			m.input.Reset()
			return m, func() tea.Msg {
				it := queue.Item{Skill: skill, Task: taskDesc, Priority: 3}
				_, _, err := queue.Add(root, it)
				return addedTaskMsg{skill: skill, err: err}
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	// Normal mode keys.
	switch strings.ToLower(key) {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "r":
		return m, loadGraph(m.root)

	case "up", "k":
		skills := m.g.SkillNodes()
		if m.cursor > 0 && len(skills) > 0 {
			m.cursor--
		}

	case "down", "j":
		skills := m.g.SkillNodes()
		if m.cursor < len(skills)-1 {
			m.cursor++
		}

	case "a":
		m.mode = modeAddSkill
		m.input.Placeholder = "skill name (e.g. sales-closer)..."
		m.input.Focus()
		return m, textinput.Blink

	case "v":
		exe := os.Args[0]
		_ = exec.Command(exe, "visualize").Start()

	case "c":
		exe := os.Args[0]
		_ = exec.Command(exe, "coffee").Start()
	}

	return m, nil
}
