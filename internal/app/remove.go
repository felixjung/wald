package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

// Remove removes a worktree for the given project.
func (a *App) Remove(ctx context.Context, projectName, branch string) error {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	gitDir := filepath.Join(projectRoot, project.DefaultBranch)
	worktreePath := filepath.Join(projectRoot, branch)
	if _, err := worktree.Remove(ctx, a.deps.Runner, gitDir, worktreePath); err != nil {
		return err
	}
	return nil
}
