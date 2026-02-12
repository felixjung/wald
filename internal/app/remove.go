package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixjung/forest/internal/config"
	"github.com/felixjung/forest/internal/hooks"
	wt "github.com/felixjung/forest/internal/worktree"
)

// Remove removes a worktree for the given project.
func (a *App) Remove(ctx context.Context, projectName, worktree string, extraArgs []string) error {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return fmt.Errorf("project %q not found", projectName)
	}
	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	gitDir := filepath.Join(projectRoot, project.DefaultBranch)
	worktreePath := filepath.Join(projectRoot, worktree)
	if _, err := os.Stat(worktreePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("worktree path does not exist: %s", worktreePath)
		}
		return fmt.Errorf("check worktree path: %w", err)
	}
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	if err != nil {
		return fmt.Errorf("resolve worktree path: %w", err)
	}
	if project.Hooks != nil {
		worktreeWorkdirPath := workdirPath(worktreePath, project.Workdir)
		if err := hooks.RunAll(ctx, a.deps.Runner, worktreeWorkdirPath, "pre_remove", project.Hooks.PreRemove); err != nil {
			return err
		}
	}
	if _, err := wt.Remove(ctx, a.deps.Runner, gitDir, relativePath, extraArgs); err != nil {
		return err
	}
	if project.Hooks != nil {
		defaultBranchWorkdir := workdirPath(gitDir, project.Workdir)
		if err := hooks.RunAll(ctx, a.deps.Runner, defaultBranchWorkdir, "post_remove", project.Hooks.PostRemove); err != nil {
			return err
		}
	}
	return nil
}
