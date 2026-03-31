package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestParseRemoveArgs(t *testing.T) {
	t.Run("with explicit worktree", func(t *testing.T) {
		worktree, extraArgs := parseRemoveArgs([]string{"feature", "--force"})
		require.Equal(t, "feature", worktree)
		require.Equal(t, []string{"--force"}, extraArgs)
	})

	t.Run("with inferred worktree and passthrough args", func(t *testing.T) {
		worktree, extraArgs := parseRemoveArgs([]string{"--force"})
		require.Empty(t, worktree)
		require.Equal(t, []string{"--force"}, extraArgs)
	})
}

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

func TestWriteRemoveSwitchTargetSkipsWhenNoOutFile(t *testing.T) {
	fake := &fakeRemoveApp{}

	err := writeRemoveSwitchTarget(context.Background(), fake, "repo", "feature", nil)
	require.NoError(t, err)
	require.Zero(t, fake.listCalls)
	require.Zero(t, fake.switchTargetCalls)
}

func TestWriteRemoveSwitchTargetWritesDefaultWorktreeTarget(t *testing.T) {
	targetFile := filepath.Join(t.TempDir(), "target")
	require.NoError(t, os.WriteFile(targetFile, nil, 0o600))
	t.Setenv("WALD_SWITCH_OUT_FILE", targetFile)

	groups := []app.ProjectWorktrees{
		{Project: config.Project{Name: "repo", DefaultBranch: "trunk"}},
	}
	fake := &fakeRemoveApp{
		switchTarget: "/tmp/repo/trunk/apps/repo",
	}

	err := writeRemoveSwitchTarget(context.Background(), fake, "repo", "feature", groups)
	require.NoError(t, err)
	require.Equal(t, 1, fake.switchTargetCalls)
	require.Equal(t, "repo", fake.switchTargetProject)
	require.Equal(t, "trunk", fake.switchTargetWorktree)

	content, readErr := os.ReadFile(targetFile)
	require.NoError(t, readErr)
	require.Equal(t, fake.switchTarget, string(content))
}

func TestWriteRemoveSwitchTargetSkipsWhenRemovingDefaultWorktree(t *testing.T) {
	targetFile := filepath.Join(t.TempDir(), "target")
	t.Setenv("WALD_SWITCH_OUT_FILE", targetFile)

	groups := []app.ProjectWorktrees{
		{Project: config.Project{Name: "repo", DefaultBranch: "main"}},
	}
	fake := &fakeRemoveApp{}

	err := writeRemoveSwitchTarget(context.Background(), fake, "repo", "main", groups)
	require.NoError(t, err)
	require.Zero(t, fake.switchTargetCalls)
}

func TestRemoveCommandForwardsForceFlag(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(_ string, _ string, _ []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		return tui.SelectOption{ID: "feature"}, nil
	})()

	root := t.TempDir()
	cwd := filepath.Join(root, "repo", "feature")
	require.NoError(t, os.MkdirAll(cwd, 0o755))
	t.Chdir(cwd)

	fake := &fakeRemoveApp{
		groups: []app.ProjectWorktrees{
			{
				Project: config.Project{Name: "repo"},
				Root:    filepath.Join(root, "repo"),
				Worktrees: []app.WorktreeInfo{
					{Path: filepath.Join(root, "repo", "main"), Branch: "main"},
					{Path: filepath.Join(root, "repo", "feature"), Branch: "feature"},
				},
			},
		},
	}

	cmd := newRemoveCommand(fake)
	err := cmd.Run(context.Background(), []string{"remove", "--force"})
	require.NoError(t, err)
	require.Equal(t, 1, fake.removeCalls)
	require.Equal(t, "repo", fake.removeProject)
	require.Equal(t, "feature", fake.removeWorktree)
	require.Equal(t, []string{"--", "--force"}, fake.removeExtraArgs)
}

func TestRemoveCommandForwardsPassthroughArgsWithoutExplicitWorktree(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withSwitchSelector(func(_ string, _ string, _ []tui.SelectOption, _ ...tui.Option) (tui.SelectOption, error) {
		return tui.SelectOption{ID: "feature"}, nil
	})()

	root := t.TempDir()
	cwd := filepath.Join(root, "repo", "feature")
	require.NoError(t, os.MkdirAll(cwd, 0o755))
	t.Chdir(cwd)

	fake := &fakeRemoveApp{
		groups: []app.ProjectWorktrees{
			{
				Project: config.Project{Name: "repo"},
				Root:    filepath.Join(root, "repo"),
				Worktrees: []app.WorktreeInfo{
					{Path: filepath.Join(root, "repo", "main"), Branch: "main"},
					{Path: filepath.Join(root, "repo", "feature"), Branch: "feature"},
				},
			},
		},
	}

	cmd := newRemoveCommand(fake)
	err := cmd.Run(context.Background(), []string{"remove", "--", "--force"})
	require.NoError(t, err)
	require.Equal(t, 1, fake.removeCalls)
	require.Equal(t, "repo", fake.removeProject)
	require.Equal(t, "feature", fake.removeWorktree)
	require.Equal(t, []string{"--", "--force"}, fake.removeExtraArgs)
}

type fakeRemoveApp struct {
	groups               []app.ProjectWorktrees
	listCalls            int
	removeCalls          int
	removeProject        string
	removeWorktree       string
	removeExtraArgs      []string
	switchTarget         string
	switchTargetCalls    int
	switchTargetProject  string
	switchTargetWorktree string
}

func (*fakeRemoveApp) AddTarget(context.Context, string, string, string, []string) (string, error) {
	return "", nil
}

func (*fakeRemoveApp) Init(context.Context, app.InitProgressReporter) error {
	return nil
}

func (f *fakeRemoveApp) List(context.Context) (string, []app.ProjectWorktrees, error) {
	f.listCalls++
	return "", f.groups, nil
}

func (f *fakeRemoveApp) Remove(_ context.Context, projectName, worktree string, extraArgs []string) error {
	f.removeCalls++
	f.removeProject = projectName
	f.removeWorktree = worktree
	f.removeExtraArgs = append([]string(nil), extraArgs...)
	return nil
}

func (f *fakeRemoveApp) SwitchTarget(_ context.Context, projectName, worktree, _ string) (string, error) {
	f.switchTargetCalls++
	f.switchTargetProject = projectName
	f.switchTargetWorktree = worktree
	return f.switchTarget, nil
}
