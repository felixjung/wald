package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/hooks"
)

// Switch prints the resolved working directory for a project worktree.
func (a *App) Switch(ctx context.Context, projectName, worktree, workingDirOverride string) error {
	path, err := a.SwitchTarget(ctx, projectName, worktree, workingDirOverride)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(a.deps.Stdout, path)
	return err
}

// SwitchTarget resolves the target working directory for project/worktree selection.
func (a *App) SwitchTarget(ctx context.Context, projectName, worktree, workingDirOverride string) (string, error) {
	project, ok := a.cfg.FindProject(projectName)
	if !ok {
		return "", fmt.Errorf("project %q not found", projectName)
	}

	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	worktreePath := filepath.Join(projectRoot, worktree)
	info, err := os.Stat(worktreePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("worktree path does not exist: %s", worktreePath)
		}
		return "", fmt.Errorf("check worktree path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("worktree path is not a directory: %s", worktreePath)
	}

	effectiveWorkdir := project.Workdir
	if strings.TrimSpace(workingDirOverride) != "" {
		effectiveWorkdir = strings.TrimSpace(workingDirOverride)
	}
	if filepath.IsAbs(effectiveWorkdir) {
		return "", fmt.Errorf("working-dir must be relative: %s", effectiveWorkdir)
	}

	target := workdirPath(worktreePath, effectiveWorkdir)
	targetInfo, err := os.Stat(target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("working directory does not exist: %s", target)
		}
		return "", fmt.Errorf("check working directory: %w", err)
	}
	if !targetInfo.IsDir() {
		return "", fmt.Errorf("working directory is not a directory: %s", target)
	}

	if a.cfg.Hooks != nil {
		vars := hookVars(project, worktree, worktreePath, target)
		if err := hooks.RunAll(ctx, a.deps.Runner, target, "post-switch", a.cfg.Hooks.PostSwitch, vars); err != nil {
			return "", err
		}
	}
	return target, nil
}
