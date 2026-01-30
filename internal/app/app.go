package app

import (
	"io"

	"github.com/felixjung/trees/internal/worktree"
)

// Deps wires dependencies for the app logic.
type Deps struct {
	Runner worktree.Runner
	Stdout io.Writer
}
