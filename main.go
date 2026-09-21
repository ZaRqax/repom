package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	workDir := "."
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}

	workDir, err := filepath.Abs(workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	repos, err := scanRepos(workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning repos: %v\n", err)
		os.Exit(1)
	}

	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "no git repositories found in", workDir)
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(workDir, repos), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func scanRepos(root string) ([]repoEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var repos []repoEntry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		repoPath := filepath.Join(root, e.Name())
		gitDir := filepath.Join(repoPath, ".git")
		info, statErr := os.Stat(gitDir)
		if statErr != nil || !info.IsDir() {
			continue
		}

		branch, _ := gitCurrentBranch(repoPath)
		dirty, _ := gitIsDirty(repoPath)

		repos = append(repos, repoEntry{
			path:   repoPath,
			name:   e.Name(),
			branch: branch,
			dirty:  dirty,
		})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].name < repos[j].name
	})

	return repos, nil
}
