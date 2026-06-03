package main

import (
	"context"
	"testing"

	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestAddCommandRejectsUnsupportedPositionalArg(t *testing.T) {
	fake := &fakeCLIApp{}
	cmd := newAddCommand(fake, tui.ThemeProfile{})

	err := cmd.Run(context.Background(), []string{"add", "--project", "repo", "feature", "extra"})

	require.Error(t, err)
	require.ErrorContains(t, err, `unexpected argument "extra"`)
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "OPTIONS")
	require.Zero(t, fake.addTargetCalls)
	require.Zero(t, fake.switchTargetCalls)
}

func TestAddCommandPreservesPassthroughArgs(t *testing.T) {
	fake := &fakeCLIApp{addTarget: "/tmp/repo/feature"}
	cmd := newAddCommand(fake, tui.ThemeProfile{})

	err := cmd.Run(context.Background(), []string{"add", "--project", "repo", "--no-switch", "feature", "--", "--force"})

	require.NoError(t, err)
	require.Equal(t, 1, fake.addTargetCalls)
	require.Equal(t, "repo", fake.addProject)
	require.Equal(t, "feature", fake.addBranch)
	require.Empty(t, fake.addStartPoint)
	require.Equal(t, []string{"--", "--force"}, fake.addExtraArgs)
	require.Zero(t, fake.switchTargetCalls)
}

func TestRootAddCommandAcceptsSinglePositionalArg(t *testing.T) {
	fake := &fakeCLIApp{addTarget: "/tmp/repo/feature"}
	cmd := newRootCommand(fake, tui.ThemeProfile{})

	err := cmd.Run(context.Background(), []string{"wald", "add", "--project", "repo", "--no-switch", "feature"})

	require.NoError(t, err)
	require.Equal(t, 1, fake.addTargetCalls)
	require.Equal(t, "repo", fake.addProject)
	require.Equal(t, "feature", fake.addBranch)
	require.Empty(t, fake.addExtraArgs)
}

func TestRootCommandFormatsUnknownFlagWithCommandHelp(t *testing.T) {
	fake := &fakeCLIApp{}
	cmd := newRootCommand(fake, tui.ThemeProfile{})

	err := cmd.Run(context.Background(), []string{"wald", "list", "--bogus"})

	require.Error(t, err)
	require.ErrorContains(t, err, "flag provided but not defined")
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "wald list")
	require.Zero(t, fake.listCalls)
}

func TestNoArgCommandRejectsPositionalArg(t *testing.T) {
	fake := &fakeCLIApp{}
	cmd := newRootCommand(fake, tui.ThemeProfile{})

	err := cmd.Run(context.Background(), []string{"wald", "version", "extra"})

	require.Error(t, err)
	require.ErrorContains(t, err, `unexpected argument "extra"`)
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "wald version")
}

type fakeCLIApp struct {
	addTarget         string
	addTargetCalls    int
	addProject        string
	addBranch         string
	addStartPoint     string
	addExtraArgs      []string
	initCalls         int
	listCalls         int
	removeCalls       int
	switchTarget      string
	switchTargetCalls int
}

func (f *fakeCLIApp) AddTarget(_ context.Context, projectName, branch, startPoint string, extraArgs []string) (string, error) {
	f.addTargetCalls++
	f.addProject = projectName
	f.addBranch = branch
	f.addStartPoint = startPoint
	f.addExtraArgs = append([]string(nil), extraArgs...)
	return f.addTarget, nil
}

func (f *fakeCLIApp) Init(context.Context, app.InitProgressReporter) error {
	f.initCalls++
	return nil
}

func (f *fakeCLIApp) List(context.Context) (string, []app.ProjectWorktrees, error) {
	f.listCalls++
	return "", nil, nil
}

func (f *fakeCLIApp) Remove(context.Context, string, string, []string) error {
	f.removeCalls++
	return nil
}

func (f *fakeCLIApp) SwitchTarget(context.Context, string, string, string) (string, error) {
	f.switchTargetCalls++
	return f.switchTarget, nil
}
