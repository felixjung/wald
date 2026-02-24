package app

import (
	"errors"
	"io"

	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/worktree"
)

// Deps wires dependencies for the app logic.
type Deps struct {
	Runner worktree.Runner
	Stdout io.Writer
}

// App holds application dependencies and configuration.
type App struct {
	deps Deps
	cfg  *config.Config
}

// New constructs a new App instance.
func New(deps Deps, cfg *config.Config) (*App, error) {
	if deps.Runner == nil {
		return nil, errors.New("runner is required")
	}
	if deps.Stdout == nil {
		return nil, errors.New("stdout is required")
	}
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	return &App{deps: deps, cfg: cfg}, nil
}
