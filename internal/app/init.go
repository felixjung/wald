package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/gh"
)

// Init initializes worktrees for all configured projects.
func (a *App) Init(ctx context.Context, reporter InitProgressReporter) error {
	reportInitProgress(reporter, InitProgressEvent{
		Type: InitProgressEventPlan,
		Plan: initProgressPlan(a.cfg.Projects),
	})

	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskPrepareFolders,
		State:  InitTaskStateRunning,
	})
	for _, project := range a.cfg.Projects {
		taskID := initTaskIDForProject(initTaskPrepareFolders, project)
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: taskID,
			State:  InitTaskStateRunning,
		})
		if err := a.ensureProjectFolder(project); err != nil {
			reportInitProgress(reporter, InitProgressEvent{
				Type:   InitProgressEventTask,
				TaskID: taskID,
				State:  InitTaskStateFailed,
				Detail: err.Error(),
			})
			reportInitProgress(reporter, InitProgressEvent{
				Type:   InitProgressEventTask,
				TaskID: initTaskPrepareFolders,
				State:  InitTaskStateFailed,
				Detail: err.Error(),
			})
			return err
		}
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: taskID,
			State:  InitTaskStateDone,
		})
	}
	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskPrepareFolders,
		State:  InitTaskStateDone,
	})

	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskCheckGitHubCLI,
		State:  InitTaskStateRunning,
	})

	client := &gh.GitHubCLI{Runner: a.deps.Runner}

	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskGitHubAvailability,
		State:  InitTaskStateRunning,
	})
	if err := client.EnsureAvailable(); err != nil {
		wrapped := fmt.Errorf("cannot initialize default branches: %w", err)
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: initTaskGitHubAvailability,
			State:  InitTaskStateFailed,
			Detail: wrapped.Error(),
		})
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: initTaskCheckGitHubCLI,
			State:  InitTaskStateFailed,
			Detail: wrapped.Error(),
		})
		return wrapped
	}
	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskGitHubAvailability,
		State:  InitTaskStateDone,
	})

	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskGitHubAuth,
		State:  InitTaskStateRunning,
	})
	if err := client.EnsureAuthenticated(ctx); err != nil {
		wrapped := fmt.Errorf("cannot initialize default branches: %w", err)
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: initTaskGitHubAuth,
			State:  InitTaskStateFailed,
			Detail: wrapped.Error(),
		})
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: initTaskCheckGitHubCLI,
			State:  InitTaskStateFailed,
			Detail: wrapped.Error(),
		})
		return wrapped
	}
	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskGitHubAuth,
		State:  InitTaskStateDone,
	})
	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskCheckGitHubCLI,
		State:  InitTaskStateDone,
	})

	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskCloneDefaultBranch,
		State:  InitTaskStateRunning,
	})
	for _, project := range a.cfg.Projects {
		taskID := initTaskIDForProject(initTaskCloneDefaultBranch, project)
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: taskID,
			State:  InitTaskStateRunning,
		})
		existed, err := a.ensureDefaultBranch(ctx, client, project)
		if err != nil {
			reportInitProgress(reporter, InitProgressEvent{
				Type:   InitProgressEventTask,
				TaskID: taskID,
				State:  InitTaskStateFailed,
				Detail: err.Error(),
			})
			reportInitProgress(reporter, InitProgressEvent{
				Type:   InitProgressEventTask,
				TaskID: initTaskCloneDefaultBranch,
				State:  InitTaskStateFailed,
				Detail: err.Error(),
			})
			return err
		}
		detail := ""
		if existed {
			detail = "already present"
		}
		reportInitProgress(reporter, InitProgressEvent{
			Type:   InitProgressEventTask,
			TaskID: taskID,
			State:  InitTaskStateDone,
			Detail: detail,
		})
	}
	reportInitProgress(reporter, InitProgressEvent{
		Type:   InitProgressEventTask,
		TaskID: initTaskCloneDefaultBranch,
		State:  InitTaskStateDone,
	})
	return nil
}

func (a *App) ensureProjectFolder(project config.Project) error {
	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		return fmt.Errorf("create project folder %s: %w", projectRoot, err)
	}
	return nil
}

func (a *App) ensureDefaultBranch(ctx context.Context, client *gh.GitHubCLI, project config.Project) (bool, error) {
	projectRoot := config.ProjectPath(a.cfg.WorktreeRoot, project.Name)
	branchPath := filepath.Join(projectRoot, project.DefaultBranch)
	if _, err := os.Stat(branchPath); err == nil {
		return true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("check branch folder %s: %w", branchPath, err)
	}

	if err := client.Clone(ctx, project.Repo, branchPath); err != nil {
		return false, err
	}

	return false, nil
}
