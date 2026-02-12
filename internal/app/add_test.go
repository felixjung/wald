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

func TestAddRunsPostAddHooks(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       "apps/repo",
				DefaultBranch: "main",
				Hooks: &config.ProjectHooks{
					PostAdd: map[string]string{
						"01_title": "echo {{project}} {{worktree}} {{repo}} {{default_branch}} {{project_workdir}} {{worktree_path}} {{target_path}}",
						"02_copy":  "cp .env.example .env",
					},
				},
			},
		},
	}

	runner := &fakeRunner{}
	stdout := &bytes.Buffer{}
	application, err := New(Deps{Runner: runner, Stdout: stdout}, cfg)
	require.NoError(t, err)

	require.NoError(t, application.Add(context.Background(), projectName, "feature", "", []string{"--", "--track"}))

	worktreePath := filepath.Join(projectRoot, "feature")
	workdirPath := filepath.Join(worktreePath, "apps/repo")
	require.Equal(t, []runnerCall{
		{
			Dir:  gitDir,
			Name: "git",
			Args: []string{"worktree", "add", "--track", "../feature"},
		},
		{
			Dir:  workdirPath,
			Name: "sh",
			Args: []string{"-c", "echo repo feature github.com/felixjung/mono main apps/repo " + worktreePath + " " + workdirPath},
		},
		{
			Dir:  workdirPath,
			Name: "sh",
			Args: []string{"-c", "cp .env.example .env"},
		},
	}, runner.calls)
	require.Equal(t, workdirPath+"\n", stdout.String())
}

func TestAddReturnsPostAddHookError(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
				Hooks: &config.ProjectHooks{
					PostAdd: map[string]string{"hook": "npm ci"},
				},
			},
		},
	}

	runner := &fakeRunner{
		errs: map[int]error{
			1: errors.New("hook failed"),
		},
	}
	stdout := &bytes.Buffer{}
	application, err := New(Deps{Runner: runner, Stdout: stdout}, cfg)
	require.NoError(t, err)

	err = application.Add(context.Background(), projectName, "feature", "", nil)
	require.EqualError(t, err, `post-add hook "hook" failed: hook failed`)
	require.Len(t, runner.calls, 2)
	require.Empty(t, stdout.String())
}

func TestAddTargetPassesStartPoint(t *testing.T) {
	root := t.TempDir()
	projectName := "repo"
	projectRoot := filepath.Join(root, projectName)
	gitDir := filepath.Join(projectRoot, "main")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          projectName,
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
			},
		},
	}

	runner := &fakeRunner{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	target, err := application.AddTarget(context.Background(), projectName, "feature", "origin/feature/foo", nil)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(projectRoot, "feature"), target)
	require.Equal(t, []runnerCall{
		{
			Dir:  gitDir,
			Name: "git",
			Args: []string{"worktree", "add", "../feature", "origin/feature/foo"},
		},
	}, runner.calls)
}
