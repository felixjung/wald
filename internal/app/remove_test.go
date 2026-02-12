package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/felixjung/forest/internal/config"
)

func TestRemoveRunsPreAndPostHooks(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	worktreePath := filepath.Join(projectRoot, "feature")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(worktreePath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       "apps/repo",
				DefaultBranch: "main",
				Hooks: &config.ProjectHooks{
					PreRemove:  []string{"echo before remove"},
					PostRemove: []string{"echo after remove"},
				},
			},
		},
	}

	runner := &fakeRunner{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	require.NoError(t, application.Remove(context.Background(), projectName, "feature", []string{"--", "--force"}))

	require.Equal(t, []runnerCall{
		{
			Dir:  filepath.Join(worktreePath, "apps/repo"),
			Name: "sh",
			Args: []string{"-c", "echo before remove"},
		},
		{
			Dir:  gitDir,
			Name: "git",
			Args: []string{"worktree", "remove", "--force", "../feature"},
		},
		{
			Dir:  filepath.Join(gitDir, "apps/repo"),
			Name: "sh",
			Args: []string{"-c", "echo after remove"},
		},
	}, runner.calls)
}

func TestRemoveStopsWhenPreRemoveHookFails(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	worktreePath := filepath.Join(projectRoot, "feature")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(worktreePath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
				Hooks: &config.ProjectHooks{
					PreRemove: []string{"echo before remove"},
				},
			},
		},
	}

	runner := &fakeRunner{
		errs: map[int]error{
			0: errors.New("pre failed"),
		},
	}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Remove(context.Background(), projectName, "feature", nil)
	require.EqualError(t, err, "pre_remove hook command 1 failed: pre failed")
	require.Len(t, runner.calls, 1)
}

func TestRemoveReturnsPostRemoveHookError(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	worktreePath := filepath.Join(projectRoot, "feature")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(worktreePath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
				Hooks: &config.ProjectHooks{
					PreRemove:  []string{"echo before remove"},
					PostRemove: []string{"echo after remove"},
				},
			},
		},
	}

	runner := &fakeRunner{
		errs: map[int]error{
			2: errors.New("post failed"),
		},
	}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Remove(context.Background(), projectName, "feature", nil)
	require.EqualError(t, err, "post_remove hook command 1 failed: post failed")
	require.Len(t, runner.calls, 3)
}
