package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ZaRqax/repom/internal/domain"
)

func (c *Client) PullWithStats(repoPath string) (domain.PullStats, error) {
	oldHead, _ := c.cmd(repoPath, "rev-parse", "HEAD")

	pullOut, err := c.cmd(repoPath, "pull")
	if err != nil {
		return domain.PullStats{}, fmt.Errorf("%s", pullOut)
	}

	if strings.Contains(pullOut, "Already up to date") {
		return domain.PullStats{UpToDate: true}, nil
	}

	var stats domain.PullStats

	newHead, _ := c.cmd(repoPath, "rev-parse", "HEAD")
	if oldHead != newHead && oldHead != "" {
		countStr, _ := c.cmd(repoPath, "rev-list", "--count", oldHead+".."+newHead)
		stats.Commits, _ = strconv.Atoi(countStr)
	}

	for _, line := range strings.Split(pullOut, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "file changed") || strings.Contains(line, "files changed") {
			stats.Summary = line
			break
		}
	}

	return stats, nil
}
