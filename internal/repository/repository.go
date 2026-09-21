package repository

import "github.com/ZaRqax/repom/internal/domain"

// Git is the low-level interface to git operations on a repository.
type Git interface {
	CurrentBranch(repoPath string) (string, error)
	IsDirty(repoPath string) (bool, error)

	Stash(repoPath string) error
	StashPop(repoPath string) error
	Checkout(repoPath, branch string) error
	Fetch(repoPath string) error
	CheckoutTrack(repoPath, local, remote string) error
	CreateBranch(repoPath, name string) error
	CommitAll(repoPath, message string) error

	LocalBranchExists(repoPath, branch string) bool
	RemoteBranchExists(repoPath, branch string) bool

	PullWithStats(repoPath string) (domain.PullStats, error)
}

// RepoFinder discovers git repositories under a root directory.
type RepoFinder interface {
	Find(root string) ([]domain.Repo, error)
}
