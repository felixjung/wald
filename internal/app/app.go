package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

// Deps wires dependencies for the app logic.
type Deps struct {
	Runner worktree.Runner
	Stdout io.Writer
}

// RunAdd adds a worktree for the given project.
func RunAdd(ctx context.Context, deps Deps, cfg *config.Config, projectName, branch string) error {
	if deps.Runner == nil {
		return errors.New("runner is required")
	}
	if deps.Stdout == nil {
		return errors.New("stdout is required")
	}
	if cfg == nil {
		return errors.New("config is required")
	}
	project, ok := cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	repoRoot := config.RepoPath(cfg.WorktreeRoot, project.Repo)
	if err := worktree.Add(ctx, deps.Runner, repoRoot, project.DefaultBranch, branch); err != nil {
		return err
	}
	path := worktree.WorkdirPath(repoRoot, branch, project.Workdir)
	_, err := fmt.Fprintln(deps.Stdout, path)
	return err
}

// RunRemove removes a worktree for the given project.
func RunRemove(ctx context.Context, deps Deps, cfg *config.Config, projectName, branch string) error {
	if deps.Runner == nil {
		return errors.New("runner is required")
	}
	if deps.Stdout == nil {
		return errors.New("stdout is required")
	}
	if cfg == nil {
		return errors.New("config is required")
	}
	project, ok := cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	repoRoot := config.RepoPath(cfg.WorktreeRoot, project.Repo)
	if err := worktree.Remove(ctx, deps.Runner, repoRoot, project.DefaultBranch, branch); err != nil {
		return err
	}
	path := worktree.WorkdirPath(repoRoot, branch, project.Workdir)
	_, err := fmt.Fprintln(deps.Stdout, path)
	return err
}
