package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixjung/wald/internal/config"
	"github.com/stretchr/testify/require"
)

func TestInitReportsProgressAndClonesRepositories(t *testing.T) {
	root := t.TempDir()
	installFakeGHBinary(t)

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{Name: "repo-one", Repo: "org/repo-one", DefaultBranch: "main"},
			{Name: "repo-two", Repo: "org/repo-two", DefaultBranch: "main"},
		},
	}

	runner := &fakeRunner{}
	reporter := &recordingInitProgressReporter{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Init(context.Background(), reporter)
	require.NoError(t, err)

	require.DirExists(t, filepath.Join(root, "repo-one"))
	require.DirExists(t, filepath.Join(root, "repo-two"))
	require.Equal(t, []runnerCall{
		{
			Dir:  "",
			Name: "gh",
			Args: []string{"auth", "status", "-h", "github.com"},
		},
		{
			Dir:  "",
			Name: "gh",
			Args: []string{"repo", "clone", "org/repo-one", filepath.Join(root, "repo-one", "main")},
		},
		{
			Dir:  "",
			Name: "gh",
			Args: []string{"repo", "clone", "org/repo-two", filepath.Join(root, "repo-two", "main")},
		},
	}, runner.calls)

	require.Equal(t, InitProgressEventPlan, reporter.events[0].Type)
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskPrepareFolders))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskCheckGitHubCLI))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskCloneDefaultBranch))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskIDForProject(initTaskCloneDefaultBranch, cfg.Projects[0])))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskIDForProject(initTaskCloneDefaultBranch, cfg.Projects[1])))
}

func TestInitMarksExistingRepositoryAsAlreadyPresent(t *testing.T) {
	root := t.TempDir()
	installFakeGHBinary(t)

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{Name: "repo-one", Repo: "org/repo-one", DefaultBranch: "main"},
			{Name: "repo-two", Repo: "org/repo-two", DefaultBranch: "main"},
		},
	}
	require.NoError(t, os.MkdirAll(filepath.Join(root, "repo-one", "main"), 0o755))

	runner := &fakeRunner{}
	reporter := &recordingInitProgressReporter{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Init(context.Background(), reporter)
	require.NoError(t, err)

	require.Equal(t, []runnerCall{
		{
			Dir:  "",
			Name: "gh",
			Args: []string{"auth", "status", "-h", "github.com"},
		},
		{
			Dir:  "",
			Name: "gh",
			Args: []string{"repo", "clone", "org/repo-two", filepath.Join(root, "repo-two", "main")},
		},
	}, runner.calls)
	require.Equal(
		t,
		[]string{"", "already present"},
		reporter.details(initTaskIDForProject(initTaskCloneDefaultBranch, cfg.Projects[0])),
	)
}

func TestInitReturnsHelpfulErrorWhenGitHubCLIIsUnavailable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects:     []config.Project{{Name: "repo", Repo: "org/repo", DefaultBranch: "main"}},
	}

	runner := &fakeRunner{}
	reporter := &recordingInitProgressReporter{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Init(context.Background(), reporter)
	require.EqualError(t, err, "cannot initialize default branches: gh CLI not found in PATH")
	require.Empty(t, runner.calls)
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskCheckGitHubCLI))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskGitHubAvailability))
}

func TestInitReturnsHelpfulErrorWhenGitHubCLIIsUnauthenticated(t *testing.T) {
	root := t.TempDir()
	installFakeGHBinary(t)

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects:     []config.Project{{Name: "repo", Repo: "org/repo", DefaultBranch: "main"}},
	}

	runner := &fakeRunner{
		errs: map[int]error{0: errors.New("auth failed")},
	}
	reporter := &recordingInitProgressReporter{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Init(context.Background(), reporter)
	require.EqualError(t, err, "cannot initialize default branches: gh is not authenticated; run 'gh auth login'")
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskGitHubAuth))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskCheckGitHubCLI))
}

func TestInitStopsAfterCloneFailure(t *testing.T) {
	root := t.TempDir()
	installFakeGHBinary(t)

	cfg := &config.Config{
		WorktreeRoot: root,
		Projects: []config.Project{
			{Name: "repo-one", Repo: "org/repo-one", DefaultBranch: "main"},
			{Name: "repo-two", Repo: "org/repo-two", DefaultBranch: "main"},
		},
	}

	runner := &fakeRunner{
		errs: map[int]error{2: errors.New("clone failed")},
	}
	reporter := &recordingInitProgressReporter{}
	application, err := New(Deps{Runner: runner, Stdout: &bytes.Buffer{}}, cfg)
	require.NoError(t, err)

	err = application.Init(context.Background(), reporter)
	require.EqualError(t, err, "clone failed")
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateDone}, reporter.states(initTaskIDForProject(initTaskCloneDefaultBranch, cfg.Projects[0])))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskIDForProject(initTaskCloneDefaultBranch, cfg.Projects[1])))
	require.Equal(t, []InitTaskState{InitTaskStateRunning, InitTaskStateFailed}, reporter.states(initTaskCloneDefaultBranch))
}

type recordingInitProgressReporter struct {
	events []InitProgressEvent
}

func (r *recordingInitProgressReporter) ReportInitProgress(event InitProgressEvent) {
	r.events = append(r.events, event)
}

func (r *recordingInitProgressReporter) states(taskID string) []InitTaskState {
	states := make([]InitTaskState, 0, len(r.events))
	for _, event := range r.events {
		if event.Type != InitProgressEventTask || event.TaskID != taskID {
			continue
		}
		states = append(states, event.State)
	}
	return states
}

func (r *recordingInitProgressReporter) details(taskID string) []string {
	details := make([]string, 0, len(r.events))
	for _, event := range r.events {
		if event.Type != InitProgressEventTask || event.TaskID != taskID {
			continue
		}
		details = append(details, event.Detail)
	}
	return details
}

func installFakeGHBinary(t *testing.T) {
	t.Helper()

	binDir := t.TempDir()
	path := filepath.Join(binDir, "gh")
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755))
	t.Setenv("PATH", binDir)
}
