package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/gh"
)

// Init initializes worktrees for all configured projects.
func (a *App) Init(ctx context.Context) error {
	for _, project := range a.cfg.Projects {
		if err := a.ensureProjectFolder(project); err != nil {
			return err
		}
	}

	client, err := gh.New(ctx, a.deps.Runner)
	if err != nil {
		switch {
		case errors.Is(err, gh.ErrNotAvailable):
			return fmt.Errorf("cannot initialize default branches: %w", err)
		case errors.Is(err, gh.ErrNotAuthenticated):
			return fmt.Errorf("cannot initialize default branches: %w", err)
		default:
			return err
		}
	}

	for _, project := range a.cfg.Projects {
		if err := a.ensureDefaultBranch(ctx, client, project); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ensureProjectFolder(project config.Project) error {
	projectRoot := config.RepoPath(a.cfg.WorktreeRoot, project.Repo)
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		return fmt.Errorf("create project folder %s: %w", projectRoot, err)
	}
	return nil
}

func (a *App) ensureDefaultBranch(ctx context.Context, client *gh.GitHubCLI, project config.Project) error {
	projectRoot := config.RepoPath(a.cfg.WorktreeRoot, project.Repo)
	branchPath := filepath.Join(projectRoot, project.DefaultBranch)
	if _, err := os.Stat(branchPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check branch folder %s: %w", branchPath, err)
	}

	if err := client.Clone(ctx, project.Repo, branchPath); err != nil {
		return err
	}

	return nil
}
