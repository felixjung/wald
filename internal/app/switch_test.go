package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixjung/forest/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSwitchPrintsResolvedTarget(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "repo")
	worktreePath := filepath.Join(projectRoot, "feature")
	targetPath := filepath.Join(worktreePath, "apps", "repo")
	require.NoError(t, os.MkdirAll(targetPath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          "repo",
				Repo:          "github.com/felixjung/mono",
				Workdir:       "apps/repo",
				DefaultBranch: "main",
			},
		},
	}

	stdout := &bytes.Buffer{}
	application, err := New(Deps{Runner: &fakeRunner{}, Stdout: stdout}, cfg)
	require.NoError(t, err)

	require.NoError(t, application.Switch(context.Background(), "repo", "feature", ""))
	require.Equal(t, targetPath+"\n", stdout.String())
}

func TestSwitchUsesWorkingDirOverride(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "repo")
	worktreePath := filepath.Join(projectRoot, "feature")
	require.NoError(t, os.MkdirAll(filepath.Join(worktreePath, "custom"), 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          "repo",
				Repo:          "github.com/felixjung/mono",
				Workdir:       "apps/repo",
				DefaultBranch: "main",
			},
		},
	}

	application, err := New(Deps{Runner: &fakeRunner{}, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	target, err := application.SwitchTarget(context.Background(), "repo", "feature", "custom")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(worktreePath, "custom"), target)
}

func TestSwitchRejectsAbsoluteWorkingDirOverride(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "repo")
	worktreePath := filepath.Join(projectRoot, "feature")
	require.NoError(t, os.MkdirAll(worktreePath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          "repo",
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
			},
		},
	}

	application, err := New(Deps{Runner: &fakeRunner{}, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	_, err = application.SwitchTarget(context.Background(), "repo", "feature", "/tmp")
	require.EqualError(t, err, "working-dir must be relative: /tmp")
}

func TestSwitchReturnsMissingWorktreeError(t *testing.T) {
	root := t.TempDir()

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{
				Name:          "repo",
				Repo:          "github.com/felixjung/mono",
				Workdir:       ".",
				DefaultBranch: "main",
			},
		},
	}

	application, err := New(Deps{Runner: &fakeRunner{}, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	_, err = application.SwitchTarget(context.Background(), "repo", "feature", "")
	require.EqualError(t, err, "worktree path does not exist: "+filepath.Join(root, "repo", "feature"))
}

func TestSwitchRunsGlobalPostSwitchHooks(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "repo")
	worktreePath := filepath.Join(projectRoot, "feature")
	targetPath := filepath.Join(worktreePath, "apps", "repo")
	require.NoError(t, os.MkdirAll(targetPath, 0o755))

	cfg := &config.Config{
		WorktreeRoot: root,
		Hooks: &config.GlobalHooks{
			PostSwitch: map[string]string{
				"set-title": "echo {{project}} {{worktree}} {{repo}} {{default_branch}} {{project_workdir}} {{worktree_path}} {{target_path}}",
			},
		},
		Projects: []config.Project{
			{
				Name:          "repo",
				Repo:          "github.com/felixjung/mono",
				Workdir:       "apps/repo",
				DefaultBranch: "main",
			},
		},
	}

	runner := &fakeRunner{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	target, err := application.SwitchTarget(context.Background(), "repo", "feature", "")
	require.NoError(t, err)
	require.Equal(t, targetPath, target)
	require.Equal(t, []runnerCall{
		{
			Dir:  targetPath,
			Name: "sh",
			Args: []string{"-c", "echo repo feature github.com/felixjung/mono main apps/repo " + worktreePath + " " + targetPath},
		},
	}, runner.calls)
}
