package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixjung/forest/internal/app"
	"github.com/felixjung/forest/internal/config"
	"github.com/felixjung/forest/internal/tui"
	"github.com/stretchr/testify/require"
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

	project, group, err := resolveProjectSelection("repo", groups, true)
	require.NoError(t, err)
	require.Equal(t, "repo", project)
	require.Equal(t, "repo", group.Project.Name)
}

func TestResolveProjectSelectionRequiresFlagWithoutTTY(t *testing.T) {
	defer withSwitchTTY(false)()
	_, _, err := resolveProjectSelection("", nil, true)
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

	project, _, err := resolveProjectSelection("", groups, true)
	require.NoError(t, err)
	require.Equal(t, "web", project)
}

func TestResolveProjectSelectionAllowsProjectWithoutWorktreesWhenNotRequired(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project:   config.Project{Name: "repo"},
			Root:      "/root/repo",
			Worktrees: nil,
		},
	}

	project, group, err := resolveProjectSelection("repo", groups, false)
	require.NoError(t, err)
	require.Equal(t, "repo", project)
	require.Equal(t, "repo", group.Project.Name)
}

func TestInferProjectNameFromPathMatchesCurrentWorktree(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    absPath("root", "repo"),
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "repo", "main"), Branch: "main"},
				{Path: absPath("root", "repo", "feature", "abc"), Branch: "feature/abc"},
			},
		},
		{
			Project: config.Project{Name: "web"},
			Root:    absPath("root", "web"),
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "web", "main"), Branch: "main"},
			},
		},
	}

	path := absPath("root", "repo", "feature", "abc", "apps", "repo")
	require.Equal(t, "repo", inferProjectNameFromPath(path, groups, true))
}

func TestInferProjectNameFromPathReturnsEmptyWhenNoMatch(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "repo"},
			Root:    absPath("root", "repo"),
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "repo", "main"), Branch: "main"},
			},
		},
	}

	path := absPath("elsewhere")
	require.Empty(t, inferProjectNameFromPath(path, groups, true))
}

func TestInferProjectNameFromPathReturnsEmptyOnAmbiguousMatch(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "api"},
			Root:    absPath("root", "api"),
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "shared"), Branch: "main"},
			},
		},
		{
			Project: config.Project{Name: "web"},
			Root:    absPath("root", "web"),
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "shared", "nested"), Branch: "main"},
			},
		},
	}

	path := absPath("root", "shared", "nested", "apps", "web")
	require.Empty(t, inferProjectNameFromPath(path, groups, true))
}

func TestInferProjectNameFromPathSkipsMissingOrEmptyProjectsWhenRequired(t *testing.T) {
	groups := []app.ProjectWorktrees{
		{
			Project: config.Project{Name: "missing"},
			Missing: true,
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "missing", "main"), Branch: "main"},
			},
		},
		{
			Project:   config.Project{Name: "empty"},
			Worktrees: nil,
		},
		{
			Project: config.Project{Name: "repo"},
			Worktrees: []app.WorktreeInfo{
				{Path: absPath("root", "repo", "main"), Branch: "main"},
			},
		},
	}

	path := absPath("root", "repo", "main")
	require.Equal(t, "repo", inferProjectNameFromPath(path, groups, true))
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

	worktree, create, err := resolveWorktreeSelection(group, "feature/one", false)
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("feature/one"), worktree)
	require.False(t, create)
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

	_, _, err := resolveWorktreeSelection(group, "missing", false)
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

	worktree, create, err := resolveWorktreeSelection(group, "", false)
	require.NoError(t, err)
	require.Equal(t, "feature", worktree)
	require.False(t, create)
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

	_, _, err := resolveWorktreeSelection(group, "", false)
	require.EqualError(t, err, "select failed")
}

func TestResolveWorktreeSelectionCreateUsesProvidedValueWhenMissing(t *testing.T) {
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
		},
	}

	worktree, create, err := resolveWorktreeSelection(group, "feature/new", true)
	require.NoError(t, err)
	require.Equal(t, "feature/new", worktree)
	require.True(t, create)
}

func TestResolveWorktreeSelectionCreatePromptsWhenMissing(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchPrompt(func(_ string, _ []tui.Field, _ ...tui.Option) ([]tui.Field, error) {
		return []tui.Field{{ID: "worktree", Value: "feature/new"}}, nil
	})()
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
		Worktrees: []app.WorktreeInfo{
			{Path: "/root/repo/main", Branch: "main"},
		},
	}

	worktree, create, err := resolveWorktreeSelection(group, "", true)
	require.NoError(t, err)
	require.Equal(t, "feature/new", worktree)
	require.True(t, create)
}

func TestResolveWorktreeSelectionCreateRequiresWorktreeWithoutTTY(t *testing.T) {
	defer withSwitchTTY(false)()
	group := app.ProjectWorktrees{
		Project: config.Project{Name: "repo"},
		Root:    "/root/repo",
	}

	_, _, err := resolveWorktreeSelection(group, "", true)
	require.EqualError(t, err, "worktree is required")
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

func withSwitchPrompt(fn func(string, []tui.Field, ...tui.Option) ([]tui.Field, error)) func() {
	original := promptFields
	promptFields = fn
	return func() {
		promptFields = original
	}
}

func absPath(parts ...string) string {
	allParts := append([]string{string(filepath.Separator)}, parts...)
	return filepath.Join(allParts...)
}
