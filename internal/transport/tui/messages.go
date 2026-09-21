package tui

import "github.com/ZaRqax/repom/internal/domain"

type opKind int

const (
	opUpdate opKind = iota
	opBranch
)

// repoStepMsg is sent when a single repo operation finishes.
type repoStepMsg struct {
	index  int
	result domain.OpResult
}
