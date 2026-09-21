package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	stateList viewState = iota
	stateBranchInput
	stateRunning
	stateResults
)

type repoEntry struct {
	path     string
	name     string
	branch   string
	dirty    bool
	selected bool
}

// repoStepMsg is sent when a single repo goroutine finishes.
type repoStepMsg struct {
	index  int
	result opResult
}

type model struct {
	workDir     string
	repos       []repoEntry
	cursor      int
	state       viewState
	branchInput textinput.Model
	spinner     spinner.Model
	width       int
	height      int

	// Operation progress (parallel)
	opLabel       string
	pendingRepos  []repoEntry
	progressSlots []*opResult // nil = still running, non-nil = done
	results       []opResult  // final results for stateResults
}

func newModel(workDir string, repos []repoEntry) model {
	ti := textinput.New()
	ti.Placeholder = "feature/TRP-0000-description"
	ti.CharLimit = 100
	ti.Width = 50

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return model{
		workDir:     workDir,
		repos:       repos,
		branchInput: ti,
		spinner:     sp,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case repoStepMsg:
		m.progressSlots[msg.index] = &msg.result

		// Check if all done
		allDone := true
		for _, slot := range m.progressSlots {
			if slot == nil {
				allDone = false
				break
			}
		}
		if allDone {
			m.state = stateResults
			m.results = make([]opResult, len(m.progressSlots))
			for i, slot := range m.progressSlots {
				m.results[i] = *slot
			}
			m.refreshRepos()
		}
		return m, nil
	}

	switch m.state {
	case stateList:
		return m.updateList(msg)
	case stateBranchInput:
		return m.updateBranchInput(msg)
	case stateRunning:
		return m.updateRunning(msg)
	case stateResults:
		return m.updateResults(msg)
	}

	return m, nil
}

// runAllParallel launches all repo operations as concurrent goroutines.
func runAllParallel(repos []repoEntry, op, branch string) tea.Cmd {
	cmds := make([]tea.Cmd, len(repos))
	for i, repo := range repos {
		i, repo := i, repo
		cmds[i] = func() tea.Msg {
			var result opResult
			switch op {
			case "update":
				result = updateRepo(repo)
			case "branch":
				result = createBranchInRepo(repo, branch)
			}
			return repoStepMsg{index: i, result: result}
		}
	}
	return tea.Batch(cmds...)
}

func (m *model) refreshRepos() {
	refreshed, err := scanRepos(m.workDir)
	if err != nil {
		return
	}
	selMap := make(map[string]bool, len(m.repos))
	for _, r := range m.repos {
		selMap[r.name] = r.selected
	}
	for i := range refreshed {
		refreshed[i].selected = selMap[refreshed[i].name]
	}
	m.repos = refreshed
	if m.cursor >= len(m.repos) {
		m.cursor = len(m.repos) - 1
	}
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.repos)-1 {
			m.cursor++
		}
	case " ":
		if len(m.repos) > 0 {
			m.repos[m.cursor].selected = !m.repos[m.cursor].selected
		}
	case "a":
		allSelected := true
		for _, r := range m.repos {
			if !r.selected {
				allSelected = false
				break
			}
		}
		for i := range m.repos {
			m.repos[i].selected = !allSelected
		}
	case "u":
		selected := m.getSelected()
		if len(selected) == 0 {
			return m, nil
		}
		m.state = stateRunning
		m.opLabel = "Updating"
		m.pendingRepos = selected
		m.progressSlots = make([]*opResult, len(selected))
		return m, tea.Batch(m.spinner.Tick, runAllParallel(selected, "update", ""))
	case "b":
		selected := m.getSelected()
		if len(selected) == 0 {
			return m, nil
		}
		m.state = stateBranchInput
		m.branchInput.Focus()
		return m, textinput.Blink
	case "r":
		m.refreshRepos()
	}

	return m, nil
}

func (m model) updateBranchInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateList
			m.branchInput.SetValue("")
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.branchInput.Value())
			if name == "" {
				return m, nil
			}
			selected := m.getSelected()
			m.state = stateRunning
			m.opLabel = "Creating branches"
			m.pendingRepos = selected
			m.progressSlots = make([]*opResult, len(selected))
			m.branchInput.SetValue("")
			return m, tea.Batch(m.spinner.Tick, runAllParallel(selected, "branch", name))
		}
	}

	var cmd tea.Cmd
	m.branchInput, cmd = m.branchInput.Update(msg)
	return m, cmd
}

func (m model) updateRunning(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		default:
			m.state = stateList
			for i := range m.repos {
				m.repos[i].selected = false
			}
		}
	}
	return m, nil
}

func (m model) getSelected() []repoEntry {
	var sel []repoEntry
	for _, r := range m.repos {
		if r.selected {
			sel = append(sel, r)
		}
	}
	return sel
}
