package repomanager

import (
	"fmt"

	"github.com/ZaRqax/repom/internal/domain"
	"github.com/ZaRqax/repom/internal/repository"
)

const tempCommitMessage = "WIP: temporary commit before branch switch"

// Service orchestrates repository operations and default-branch resolution.
type Service struct {
	git    repository.Git
	finder repository.RepoFinder
}

func New(git repository.Git, finder repository.RepoFinder) *Service {
	return &Service{git: git, finder: finder}
}

func (s *Service) Repos(root string) ([]domain.Repo, error) {
	return s.finder.Find(root)
}

// Update stashes local changes, checks out the default branch and pulls.
func (s *Service) Update(repo domain.Repo) domain.OpResult {
	targetBranch := s.defaultBranch(repo.Path)
	stashed := false

	if repo.Dirty {
		if err := s.git.Stash(repo.Path); err != nil {
			return domain.OpResult{RepoName: repo.Name, Message: fmt.Sprintf("stash: %v", err)}
		}
		stashed = true
	}

	if err := s.git.Checkout(repo.Path, targetBranch); err != nil {
		if stashed {
			_ = s.git.Checkout(repo.Path, repo.Branch)
			_ = s.git.StashPop(repo.Path)
		}
		return domain.OpResult{
			RepoName: repo.Name,
			Message:  fmt.Sprintf("checkout %s: %v", targetBranch, err),
		}
	}

	stats, err := s.git.PullWithStats(repo.Path)
	if err != nil {
		return domain.OpResult{
			RepoName: repo.Name,
			Message:  fmt.Sprintf("pull %s: %v", targetBranch, err),
		}
	}

	message := targetBranch + " — " + formatPullStats(stats)
	if stashed {
		message = "stashed → " + message
	}

	return domain.OpResult{RepoName: repo.Name, Success: true, Message: message}
}

// CreateBranch commits local changes if needed, updates the default branch and
// creates a new branch off it.
func (s *Service) CreateBranch(repo domain.Repo, branchName string) domain.OpResult {
	baseBranch := s.defaultBranch(repo.Path)

	if repo.Dirty {
		if err := s.git.CommitAll(repo.Path, tempCommitMessage); err != nil {
			return domain.OpResult{RepoName: repo.Name, Message: fmt.Sprintf("temp commit: %v", err)}
		}
	}

	if err := s.git.Checkout(repo.Path, baseBranch); err != nil {
		_ = s.git.Fetch(repo.Path)
		if trackErr := s.git.CheckoutTrack(repo.Path, baseBranch, "origin/"+baseBranch); trackErr != nil {
			return domain.OpResult{
				RepoName: repo.Name,
				Message:  fmt.Sprintf("checkout %s: %v", baseBranch, trackErr),
			}
		}
	}

	stats, err := s.git.PullWithStats(repo.Path)
	if err != nil {
		return domain.OpResult{
			RepoName: repo.Name,
			Message:  fmt.Sprintf("pull %s: %v", baseBranch, err),
		}
	}

	if err := s.git.CreateBranch(repo.Path, branchName); err != nil {
		return domain.OpResult{RepoName: repo.Name, Message: fmt.Sprintf("create branch: %v", err)}
	}

	return domain.OpResult{
		RepoName: repo.Name,
		Success:  true,
		Message:  fmt.Sprintf("%s (%s) → %s", baseBranch, formatPullStats(stats), branchName),
	}
}

func (s *Service) defaultBranch(repoPath string) string {
	if s.git.LocalBranchExists(repoPath, "develop") || s.git.RemoteBranchExists(repoPath, "develop") {
		return "develop"
	}
	return "main"
}
