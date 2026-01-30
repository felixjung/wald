package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

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
