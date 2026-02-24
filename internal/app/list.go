package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/worktree"
)

// WorktreeInfo describes a git worktree entry.
type WorktreeInfo struct {
	Path     string
	Branch   string
	Detached bool
}

// ProjectWorktrees groups worktrees for a project.
type ProjectWorktrees struct {
	Project   config.Project
	Root      string
	Worktrees []WorktreeInfo
	Missing   bool
}

// List returns worktrees grouped by project.
func (a *App) List(ctx context.Context) (string, []ProjectWorktrees, error) {
	results := make([]ProjectWorktrees, 0, len(a.cfg.Projects))
	for _, project := range a.cfg.Projects {
		projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
		gitDir := filepath.Join(projectRoot, project.DefaultBranch)
		if _, err := os.Stat(gitDir); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				results = append(results, ProjectWorktrees{Project: project, Root: projectRoot, Missing: true})
				continue
			}
			return "", nil, fmt.Errorf("check default branch worktree %s: %w", gitDir, err)
		}

		outputRunner, ok := a.deps.Runner.(worktree.OutputRunner)
		if !ok {
			return "", nil, errors.New("runner does not support capturing command output")
		}

		entries, err := worktree.List(ctx, outputRunner, gitDir)
		if err != nil {
			return "", nil, err
		}

		worktrees := make([]WorktreeInfo, 0, len(entries))
		for _, entry := range entries {
			worktrees = append(worktrees, WorktreeInfo{
				Path:     entry.Path,
				Branch:   entry.Branch,
				Detached: entry.Detached,
			})
		}
		results = append(results, ProjectWorktrees{Project: project, Root: projectRoot, Worktrees: worktrees})
	}
	return a.cfg.WorktreeRoot, results, nil
}
