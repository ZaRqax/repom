package domain

// Repo is a git repository discovered in the workspace.
type Repo struct {
	Name   string
	Path   string
	Branch string
	Dirty  bool
}

// PullStats describes the outcome of a pull operation.
type PullStats struct {
	UpToDate bool
	Commits  int
	Summary  string
}

// OpResult is the outcome of an operation performed on a single repository.
type OpResult struct {
	RepoName string
	Success  bool
	Message  string
}
