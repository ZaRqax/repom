package finder

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZaRqax/repom/internal/domain"
	"github.com/ZaRqax/repom/internal/repository"
)

// Finder discovers git repositories as direct subdirectories of a root.
type Finder struct {
	git repository.Git
}

func New(git repository.Git) *Finder {
	return &Finder{git: git}
}

func (f *Finder) Find(root string) ([]domain.Repo, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var repos []domain.Repo
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		repoPath := filepath.Join(root, e.Name())
		info, statErr := os.Stat(filepath.Join(repoPath, ".git"))
		if statErr != nil || !info.IsDir() {
			continue
		}

		branch, _ := f.git.CurrentBranch(repoPath)
		dirty, _ := f.git.IsDirty(repoPath)

		repos = append(repos, domain.Repo{
			Name:   e.Name(),
			Path:   repoPath,
			Branch: branch,
			Dirty:  dirty,
		})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}
