package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felixjung/forest/internal/app"
	"github.com/felixjung/forest/internal/config"
	"github.com/felixjung/forest/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestResolveRemoveSelectionInfersProjectAndSelectsWorktree(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(title string, _ string, _ []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		require.Equal(t, "Select worktree (repo)", title)
		return tui.SelectOption{ID: "feature"}, nil
	})()

	root := t.TempDir()
	cwd := filepath.Join(root, "repo", "feature", "apps", "repo")
	require.NoError(t, os.MkdirAll(cwd, 0o755))
	t.Chdir(cwd)

	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    filepath.Join(root, "repo"),
			Worktrees: []app.WorktreeInfo{
				{Path: filepath.Join(root, "repo", "main"), Branch: "main"},
				{Path: filepath.Join(root, "repo", "feature"), Branch: "feature"},
			},
		},
		{
			Project: config.Project{Name: "web"},
			Root:    filepath.Join(root, "web"),
			Worktrees: []app.WorktreeInfo{
				{Path: filepath.Join(root, "web", "main"), Branch: "main"},
			},
		},
	}

	project, worktree, err := resolveRemoveSelection("", "", groups)
	require.NoError(t, err)
	require.Equal(t, "repo", project)
	require.Equal(t, "feature", worktree)
}

func TestResolveRemoveSelectionInfersProjectWithoutTTYWhenWorktreeProvided(t *testing.T) {
	defer withSwitchTTY(false)()

	root := t.TempDir()
	cwd := filepath.Join(root, "repo", "feature")
	require.NoError(t, os.MkdirAll(cwd, 0o755))
	t.Chdir(cwd)

	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    filepath.Join(root, "repo"),
			Worktrees: []app.WorktreeInfo{
				{Path: filepath.Join(root, "repo", "main"), Branch: "main"},
				{Path: filepath.Join(root, "repo", "feature"), Branch: "feature"},
			},
		},
	}

	project, worktree, err := resolveRemoveSelection("", "feature", groups)
	require.NoError(t, err)
	require.Equal(t, "repo", project)
	require.Equal(t, "feature", worktree)
}

func TestResolveRemoveSelectionRequiresProjectWithoutTTYWhenPathUnknown(t *testing.T) {
	defer withSwitchTTY(false)()

	root := t.TempDir()
	unknown := filepath.Join(root, "elsewhere")
	require.NoError(t, os.MkdirAll(unknown, 0o755))
	t.Chdir(unknown)

	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    filepath.Join(root, "repo"),
			Worktrees: []app.WorktreeInfo{
				{Path: filepath.Join(root, "repo", "main"), Branch: "main"},
			},
		},
	}

	_, _, err := resolveRemoveSelection("", "main", groups)
	require.EqualError(t, err, "project name is required")
}
