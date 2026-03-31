package main

import (
	"context"
	"testing"

	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestInitCommandUsesTaskListInInteractiveNonVerboseMode(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withInitTaskList(func(title string, tasks []tui.Task, _ <-chan tui.TaskUpdate, _ ...tui.Option) error {
		require.Equal(t, "Initializing worktrees", title)
		require.Equal(t, []tui.Task{
			{ID: "clone", Label: "Clone default branches"},
			{ID: "repo", ParentID: "clone", Label: "repo"},
		}, tasks)
		return nil
	})()

	fake := &fakeInitApp{
		initFn: func(_ context.Context, reporter app.InitProgressReporter) error {
			reporter.ReportInitProgress(app.InitProgressEvent{
				Type: app.InitProgressEventPlan,
				Plan: app.InitProgressPlan{
					Tasks: []app.InitTask{
						{ID: "clone", Label: "Clone default branches"},
						{ID: "repo", ParentID: "clone", Label: "repo"},
					},
				},
			})
			return nil
		},
	}

	cmd := newRootCommand(fake, tui.DefaultThemeProfile())
	require.NoError(t, cmd.Run(context.Background(), []string{"wald", "init"}))
	require.Equal(t, 1, fake.initCalls)
}

func TestInitCommandFallsBackToPlainTextForVerboseRuns(t *testing.T) {
	defer withSwitchTTY(true)()
	defer withInitTaskList(func(_ string, _ []tui.Task, _ <-chan tui.TaskUpdate, _ ...tui.Option) error {
		t.Fatal("task list should not run in verbose mode")
		return nil
	})()

	fake := &fakeInitApp{
		initFn: func(_ context.Context, reporter app.InitProgressReporter) error {
			_, ok := reporter.(*initPlainTextReporter)
			require.True(t, ok)
			return nil
		},
	}

	cmd := newRootCommand(fake, tui.DefaultThemeProfile())
	require.NoError(t, cmd.Run(context.Background(), []string{"wald", "--verbose", "init"}))
	require.Equal(t, 1, fake.initCalls)
}

func TestInitCommandFallsBackToPlainTextWithoutTTY(t *testing.T) {
	defer withSwitchTTY(false)()
	defer withInitTaskList(func(_ string, _ []tui.Task, _ <-chan tui.TaskUpdate, _ ...tui.Option) error {
		t.Fatal("task list should not run without a TTY")
		return nil
	})()

	fake := &fakeInitApp{
		initFn: func(_ context.Context, reporter app.InitProgressReporter) error {
			_, ok := reporter.(*initPlainTextReporter)
			require.True(t, ok)
			return nil
		},
	}

	cmd := newRootCommand(fake, tui.DefaultThemeProfile())
	require.NoError(t, cmd.Run(context.Background(), []string{"wald", "init"}))
	require.Equal(t, 1, fake.initCalls)
}

type fakeInitApp struct {
	initCalls int
	initFn    func(context.Context, app.InitProgressReporter) error
}

func (*fakeInitApp) AddTarget(context.Context, string, string, string, []string) (string, error) {
	return "", nil
}

func (f *fakeInitApp) Init(ctx context.Context, reporter app.InitProgressReporter) error {
	f.initCalls++
	if f.initFn != nil {
		return f.initFn(ctx, reporter)
	}
	return nil
}

func (*fakeInitApp) List(context.Context) (string, []app.ProjectWorktrees, error) {
	return "", nil, nil
}

func (*fakeInitApp) Remove(context.Context, string, string, []string) error {
	return nil
}

func (*fakeInitApp) SwitchTarget(context.Context, string, string, string) (string, error) {
	return "", nil
}

func withInitTaskList(
	fn func(string, []tui.Task, <-chan tui.TaskUpdate, ...tui.Option) error,
) func() {
	original := runInitTaskList
	runInitTaskList = fn
	return func() {
		runInitTaskList = original
	}
}
