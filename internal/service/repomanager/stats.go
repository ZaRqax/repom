package repomanager

import (
	"fmt"
	"strings"

	"github.com/ZaRqax/repom/internal/domain"
)

func formatPullStats(stats domain.PullStats) string {
	if stats.UpToDate {
		return "up to date"
	}

	var parts []string
	if stats.Commits > 0 {
		word := "commits"
		if stats.Commits == 1 {
			word = "commit"
		}
		parts = append(parts, fmt.Sprintf("%d %s", stats.Commits, word))
	}
	if stats.Summary != "" {
		parts = append(parts, stats.Summary)
	}
	if len(parts) == 0 {
		return "pulled"
	}

	return strings.Join(parts, ", ")
}
