package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

// Add adds a worktree for the given project.
func (a *App) Add(ctx context.Context, projectName, branch string) (err error) {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}

	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	gitDir := filepath.Join(projectRoot, project.DefaultBranch)
	worktreePath := filepath.Join(projectRoot, branch)
	worktreePath, err = worktree.Add(ctx, a.deps.Runner, gitDir, worktreePath)
	if err != nil {
		return err
	}

	path := workdirPath(worktreePath, project.Workdir)
	_, err = fmt.Fprintln(a.deps.Stdout, path)
	return err
}
