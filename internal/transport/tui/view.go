package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorFmt  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	checkFmt   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	branchFmt  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	dirtyFmt   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	successFmt = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorFmt   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	promptFmt  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
)

func (m model) View() string {
	switch m.state {
	case stateList:
		return m.viewList()
	case stateBranchInput:
		return m.viewBranchInput()
	case stateRunning:
		return m.viewRunning()
	case stateResults:
		return m.viewResults()
	}
	return ""
}

func (m model) viewList() string {
	var b strings.Builder

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render("Repo Manager"))
	selCount := 0
	for _, item := range m.repos {
		if item.selected {
			selCount++
		}
	}
	if selCount > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %d selected", selCount)))
	}
	b.WriteString("\n\n")

	maxName := 0
	for _, item := range m.repos {
		if len(item.repo.Name) > maxName {
			maxName = len(item.repo.Name)
		}
	}

	visible := m.height - 7
	if visible <= 0 {
		visible = len(m.repos)
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := start + visible
	if end > len(m.repos) {
		end = len(m.repos)
	}

	for i := start; i < end; i++ {
		item := m.repos[i]

		cursor := "  "
		if i == m.cursor {
			cursor = cursorFmt.Render("▸ ")
		}

		check := "□"
		if item.selected {
			check = checkFmt.Render("■")
		}

		name := fmt.Sprintf("%-*s", maxName, item.repo.Name)
		branch := branchFmt.Render(item.repo.Branch)

		dirty := ""
		if item.repo.Dirty {
			dirty = dirtyFmt.Render(" ●")
		}

		fmt.Fprintf(&b, "  %s%s %s  %s%s\n", cursor, check, name, branch, dirty)
	}

	if len(m.repos) > visible {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n  %d–%d of %d repos", start+1, end, len(m.repos))))
	}

	b.WriteString("\n  ")
	b.WriteString(dimStyle.Render("↑/↓ move  space select  a all  u update  b branch  r refresh  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewBranchInput() string {
	var b strings.Builder

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render("Create Branch"))
	b.WriteString("\n\n")

	selCount := 0
	for _, item := range m.repos {
		if item.selected {
			selCount++
		}
	}

	fmt.Fprintf(&b, "  Creating branch in %d repositories\n\n", selCount)
	b.WriteString("  ")
	b.WriteString(promptFmt.Render("Branch name: "))
	b.WriteString(m.branchInput.View())
	b.WriteString("\n\n  ")
	b.WriteString(dimStyle.Render("enter confirm  esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewRunning() string {
	var b strings.Builder

	total := len(m.pendingRepos)
	done := 0
	for _, slot := range m.progressSlots {
		if slot != nil {
			done++
		}
	}

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render(m.opLabel))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", done, total)))
	b.WriteString("\n\n")

	maxName := 0
	for _, item := range m.pendingRepos {
		if len(item.repo.Name) > maxName {
			maxName = len(item.repo.Name)
		}
	}

	for i, item := range m.pendingRepos {
		name := fmt.Sprintf("%-*s", maxName, item.repo.Name)
		slot := m.progressSlots[i]

		if slot != nil {
			icon := successFmt.Render("✓")
			msg := slot.Message
			if !slot.Success {
				icon = errorFmt.Render("✗")
				msg = errorFmt.Render(slot.Message)
			}
			fmt.Fprintf(&b, "  %s %s  %s\n", icon, name, msg)
		} else {
			fmt.Fprintf(&b, "  %s %s\n", m.spinner.View(), name)
		}
	}

	b.WriteString("\n  ")
	b.WriteString(dimStyle.Render("ctrl+c cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewResults() string {
	var b strings.Builder

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render("Results"))
	b.WriteString("\n\n")

	maxName := 0
	for _, r := range m.results {
		if len(r.RepoName) > maxName {
			maxName = len(r.RepoName)
		}
	}

	for _, r := range m.results {
		icon := successFmt.Render("✓")
		msg := r.Message
		if !r.Success {
			icon = errorFmt.Render("✗")
			msg = errorFmt.Render(r.Message)
		}
		name := fmt.Sprintf("%-*s", maxName, r.RepoName)
		fmt.Fprintf(&b, "  %s %s  %s\n", icon, name, msg)
	}

	b.WriteString("\n  ")
	b.WriteString(dimStyle.Render("press any key to continue  q quit"))
	b.WriteString("\n")

	return b.String()
}
