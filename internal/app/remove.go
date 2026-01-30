package app

import (
	"context"
	"fmt"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

// Remove removes a worktree for the given project.
func (a *App) Remove(ctx context.Context, projectName, branch string) error {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	repoRoot := config.RepoPath(a.cfg.WorktreeRoot, project.Repo)
	if err := worktree.Remove(ctx, a.deps.Runner, repoRoot, project.DefaultBranch, branch); err != nil {
		return err
	}
	path := worktree.WorkdirPath(repoRoot, branch, project.Workdir)
	_, err := fmt.Fprintln(a.deps.Stdout, path)
	return err
}
