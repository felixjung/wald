package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/felixjung/forest/internal/app"
	"github.com/felixjung/forest/internal/config"
	"github.com/felixjung/forest/internal/tui"
)

func TestResolveProjectSelectionWithFlag(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    "/root/repo",
			Worktrees: []app.WorktreeInfo{
				{Path: "/root/repo/main", Branch: "main"},
			},
		},
	}

	project, group, err := resolveProjectSelection("repo", groups)
	require.NoError(t, err)
	require.Equal(t, "repo", project)
	require.Equal(t, "repo", group.Project.Name)
}

func TestResolveProjectSelectionRequiresFlagWithoutTTY(t *testing.T) {
	defer withSwitchTTY(false)()
	_, _, err := resolveProjectSelection("", nil)
	require.EqualError(t, err, "project name is required")
}

func TestResolveProjectSelectionUsesSelector(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(_ string, _ string, options []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		require.Len(t, options, 2)
		return options[1], nil
	})()

	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "api"},
			Root:    "/root/api",
			Worktrees: []app.WorktreeInfo{
				{Path: "/root/api/main", Branch: "main"},
			},
		},
		{
			Project: config.Project{Name: "web"},
			Root:    "/root/web",
			Worktrees: []app.WorktreeInfo{
				{Path: "/root/web/main", Branch: "main"},
			},
		},
	}

	project, _, err := resolveProjectSelection("", groups)
	require.NoError(t, err)
	require.Equal(t, "web", project)
}

func TestResolveWorktreeSelectionMatchesFlag(t *testing.T) {
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo", DefaultBranch: "main"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
			{Path: "/root/repo/feature/one", Branch: "feature/one"},
		},
	}

	worktree, err := resolveWorktreeSelection(group, "feature/one")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("feature/one"), worktree)
}

func TestResolveWorktreeSelectionInvalidFlagWithoutTTY(t *testing.T) {
	defer withSwitchTTY(false)()
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
		},
	}

	_, err := resolveWorktreeSelection(group, "missing")
	require.EqualError(t, err, `worktree "missing" not found in project "repo"`)
}

func TestResolveWorktreeSelectionUsesSelectorWhenMissing(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(_ string, _ string, _ []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		return tui.SelectOption{ID: "feature"}, nil
	})()
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
			{Path: "/root/repo/feature", Branch: "feature"},
		},
	}

	worktree, err := resolveWorktreeSelection(group, "")
	require.NoError(t, err)
	require.Equal(t, "feature", worktree)
}

func TestResolveWorktreeSelectionPropagatesSelectorError(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(_ string, _ string, _ []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		return tui.SelectOption{}, errors.New("select failed")
	})()
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
		},
	}

	_, err := resolveWorktreeSelection(group, "")
	require.EqualError(t, err, "select failed")
}

func TestWriteSwitchTargetToFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "target")
	t.Setenv("FOREST_SWITCH_OUT_FILE", file)

	require.NoError(t, writeSwitchTarget("/tmp/worktree"))

	content, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, "/tmp/worktree", string(content))
}

func withSwitchTTY(v bool) func() {
	original := isTerminal
	isTerminal = func(*os.File) bool {
		return v
	}
	return func() {
		isTerminal = original
	}
}

func withSwitchSelector(fn func(string, string, []tui.SelectOption, ...tui.Option) (tui.SelectOption, error)) func() {
	original := selectOption
	selectOption = fn
	return func() {
		selectOption = original
	}
}
