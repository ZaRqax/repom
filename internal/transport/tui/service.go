package tui

import (
	"fmt"

	"github.com/ZaRqax/repom/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

// Service is the repo manager functionality consumed by the TUI.
type Service interface {
	Repos(root string) ([]domain.Repo, error)
	Update(repo domain.Repo) domain.OpResult
	CreateBranch(repo domain.Repo, branchName string) domain.OpResult
}

// Run starts the TUI for the repositories found under workDir.
func Run(svc Service, workDir string) error {
	repos, err := svc.Repos(workDir)
	if err != nil {
		return fmt.Errorf("scan repos: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found in %s", workDir)
	}

	program := tea.NewProgram(newModel(svc, workDir, repos), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	return nil
}
