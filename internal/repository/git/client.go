package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Client runs git commands for repositories on the local filesystem.
type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) cmd(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	out, err := exec.Command("git", fullArgs...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (c *Client) CurrentBranch(repoPath string) (string, error) {
	return c.cmd(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

func (c *Client) IsDirty(repoPath string) (bool, error) {
	out, err := c.cmd(repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

func (c *Client) Stash(repoPath string) error {
	_, err := c.cmd(repoPath, "stash", "--include-untracked")
	return err
}

func (c *Client) StashPop(repoPath string) error {
	_, err := c.cmd(repoPath, "stash", "pop")
	return err
}

func (c *Client) Checkout(repoPath, branch string) error {
	out, err := c.cmd(repoPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (c *Client) Fetch(repoPath string) error {
	_, err := c.cmd(repoPath, "fetch", "origin")
	return err
}

func (c *Client) CheckoutTrack(repoPath, local, remote string) error {
	out, err := c.cmd(repoPath, "checkout", "-b", local, remote)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (c *Client) CreateBranch(repoPath, name string) error {
	out, err := c.cmd(repoPath, "checkout", "-b", name)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (c *Client) CommitAll(repoPath, message string) error {
	if _, err := c.cmd(repoPath, "add", "-A"); err != nil {
		return err
	}
	out, err := c.cmd(repoPath, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

func (c *Client) LocalBranchExists(repoPath, branch string) bool {
	_, err := c.cmd(repoPath, "rev-parse", "--verify", branch)
	return err == nil
}

func (c *Client) RemoteBranchExists(repoPath, branch string) bool {
	_, err := c.cmd(repoPath, "rev-parse", "--verify", "origin/"+branch)
	return err == nil
}
