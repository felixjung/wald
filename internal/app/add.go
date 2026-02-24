package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/hooks"
	"github.com/felixjung/wald/internal/worktree"
)

// Add adds a worktree for the given project.
func (a *App) Add(ctx context.Context, projectName, branch, startPoint string, extraArgs []string) (err error) {
	path, err := a.AddTarget(ctx, projectName, branch, startPoint, extraArgs)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(a.deps.Stdout, path)
	return err
}

// AddTarget creates a worktree and returns the resolved project workdir path.
func (a *App) AddTarget(ctx context.Context, projectName, branch, startPoint string, extraArgs []string) (string, error) {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return "", fmt.Errorf("project %q not found", projectName)
	}

	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	gitDir := filepath.Join(projectRoot, project.DefaultBranch)
	worktreePath := filepath.Join(projectRoot, branch)
	if _, err := os.Stat(worktreePath); err == nil {
		return "", fmt.Errorf("worktree path already exists: %s", worktreePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("check worktree path: %w", err)
	}
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	if err != nil {
		return "", fmt.Errorf("resolve worktree path: %w", err)
	}
	_, err = worktree.Add(ctx, a.deps.Runner, gitDir, relativePath, startPoint, extraArgs)
	if err != nil {
		return "", err
	}

	path := workdirPath(worktreePath, project.Workdir)
	if project.Hooks != nil {
		vars := hookVars(project, branch, worktreePath, path)
		if err := hooks.RunAll(ctx, a.deps.Runner, path, "post-add", project.Hooks.PostAdd, vars); err != nil {
			return "", err
		}
	}
	return path, nil
}
