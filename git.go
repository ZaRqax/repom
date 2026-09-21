package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Low-level git wrappers

func gitCmd(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	out, err := exec.Command("git", fullArgs...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func gitCurrentBranch(repoPath string) (string, error) {
	return gitCmd(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

func gitIsDirty(repoPath string) (bool, error) {
	out, err := gitCmd(repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

func gitStash(repoPath string) error {
	_, err := gitCmd(repoPath, "stash", "--include-untracked")
	return err
}

func gitCheckout(repoPath, branch string) error {
	out, err := gitCmd(repoPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func gitFetch(repoPath string) error {
	_, err := gitCmd(repoPath, "fetch", "origin")
	return err
}

func gitCreateBranch(repoPath, name string) error {
	out, err := gitCmd(repoPath, "checkout", "-b", name)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func gitCommitAll(repoPath, message string) error {
	if _, err := gitCmd(repoPath, "add", "-A"); err != nil {
		return err
	}
	out, err := gitCmd(repoPath, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func gitCheckoutTrack(repoPath, local, remote string) error {
	out, err := gitCmd(repoPath, "checkout", "-b", local, remote)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

// Branch detection helpers

func gitLocalBranchExists(repoPath, branch string) bool {
	_, err := gitCmd(repoPath, "rev-parse", "--verify", branch)
	return err == nil
}

func gitRemoteBranchExists(repoPath, branch string) bool {
	_, err := gitCmd(repoPath, "rev-parse", "--verify", "origin/"+branch)
	return err == nil
}

// defaultBranch returns "develop" for most repos, "main" for repos without develop (e.g. proto).
func defaultBranch(repoPath string) string {
	if gitLocalBranchExists(repoPath, "develop") || gitRemoteBranchExists(repoPath, "develop") {
		return "develop"
	}
	return "main"
}

// Pull with stats

type pullStats struct {
	upToDate bool
	commits  int
	summary  string // e.g. "3 files changed, 10 insertions(+), 2 deletions(-)"
}

func gitPullWithStats(repoPath string) (pullStats, error) {
	oldHead, _ := gitCmd(repoPath, "rev-parse", "HEAD")

	pullOut, err := gitCmd(repoPath, "pull")
	if err != nil {
		return pullStats{}, fmt.Errorf("%s", pullOut)
	}

	var stats pullStats

	if strings.Contains(pullOut, "Already up to date") {
		stats.upToDate = true
		return stats, nil
	}

	// Count new commits between old and new HEAD
	newHead, _ := gitCmd(repoPath, "rev-parse", "HEAD")
	if oldHead != newHead && oldHead != "" {
		countStr, _ := gitCmd(repoPath, "rev-list", "--count", oldHead+".."+newHead)
		stats.commits, _ = strconv.Atoi(countStr)
	}

	// Extract file change summary from pull output
	for _, line := range strings.Split(pullOut, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "file changed") || strings.Contains(line, "files changed") {
			stats.summary = line
			break
		}
	}

	return stats, nil
}

func formatPullStats(stats pullStats) string {
	if stats.upToDate {
		return "up to date"
	}
	var parts []string
	if stats.commits > 0 {
		word := "commits"
		if stats.commits == 1 {
			word = "commit"
		}
		parts = append(parts, fmt.Sprintf("%d %s", stats.commits, word))
	}
	if stats.summary != "" {
		parts = append(parts, stats.summary)
	}
	if len(parts) == 0 {
		return "pulled"
	}
	return strings.Join(parts, ", ")
}

// High-level operations

type opResult struct {
	repoName string
	success  bool
	message  string
}

// updateRepo: stash if dirty, checkout develop (or main for repos without develop), pull.
func updateRepo(repo repoEntry) opResult {
	targetBranch := defaultBranch(repo.path)
	stashed := false

	if repo.dirty {
		if err := gitStash(repo.path); err != nil {
			return opResult{repo.name, false, fmt.Sprintf("stash: %v", err)}
		}
		stashed = true
	}

	if err := gitCheckout(repo.path, targetBranch); err != nil {
		if stashed {
			_ = gitCheckout(repo.path, repo.branch)
			_, _ = gitCmd(repo.path, "stash", "pop")
		}
		return opResult{repo.name, false, fmt.Sprintf("checkout %s: %v", targetBranch, err)}
	}

	stats, err := gitPullWithStats(repo.path)
	if err != nil {
		return opResult{repo.name, false, fmt.Sprintf("pull %s: %v", targetBranch, err)}
	}

	msg := targetBranch + " — " + formatPullStats(stats)
	if stashed {
		msg = "stashed → " + msg
	}
	return opResult{repo.name, true, msg}
}

// createBranchInRepo: temp commit if dirty, checkout develop (or main), pull, create branch.
func createBranchInRepo(repo repoEntry, branchName string) opResult {
	baseBranch := defaultBranch(repo.path)

	if repo.dirty {
		if err := gitCommitAll(repo.path, "WIP: temporary commit before branch switch"); err != nil {
			return opResult{repo.name, false, fmt.Sprintf("temp commit: %v", err)}
		}
	}

	if err := gitCheckout(repo.path, baseBranch); err != nil {
		_ = gitFetch(repo.path)
		if err2 := gitCheckoutTrack(repo.path, baseBranch, "origin/"+baseBranch); err2 != nil {
			return opResult{repo.name, false, fmt.Sprintf("checkout %s: %v", baseBranch, err2)}
		}
	}

	stats, err := gitPullWithStats(repo.path)
	if err != nil {
		return opResult{repo.name, false, fmt.Sprintf("pull %s: %v", baseBranch, err)}
	}

	if err := gitCreateBranch(repo.path, branchName); err != nil {
		return opResult{repo.name, false, fmt.Sprintf("create branch: %v", err)}
	}

	return opResult{repo.name, true, fmt.Sprintf("%s (%s) → %s", baseBranch, formatPullStats(stats), branchName)}
}
