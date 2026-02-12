package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type call struct {
	Dir  string
	Name string
	Args []string
}

type fakeRunner struct {
	calls []call
	errs  map[int]error
}

func (f *fakeRunner) Run(_ context.Context, dir, name string, args ...string) error {
	current := len(f.calls)
	f.calls = append(f.calls, call{
		Dir:  dir,
		Name: name,
		Args: append([]string(nil), args...),
	})
	if err, ok := f.errs[current]; ok {
		return err
	}
	return nil
}

func TestRunAllRunsCommandsInOrder(t *testing.T) {
	runner := &fakeRunner{}

	err := RunAll(context.Background(), runner, "/repo/worktree", "post_add", []string{
		"npm ci",
		"cp .env.example .env",
	})
	require.NoError(t, err)
	require.Equal(t, []call{
		{Dir: "/repo/worktree", Name: "sh", Args: []string{"-c", "npm ci"}},
		{Dir: "/repo/worktree", Name: "sh", Args: []string{"-c", "cp .env.example .env"}},
	}, runner.calls)
}

func TestRunAllFailsFast(t *testing.T) {
	runner := &fakeRunner{
		errs: map[int]error{
			1: errors.New("failed"),
		},
	}

	err := RunAll(context.Background(), runner, "/repo/worktree", "post_add", []string{
		"npm ci",
		"cp .env.example .env",
		"make bootstrap",
	})
	require.EqualError(t, err, "post_add hook command 2 failed: failed")
	require.Len(t, runner.calls, 2)
}

func TestRunAllValidatesInputs(t *testing.T) {
	err := RunAll(context.Background(), nil, "/repo/worktree", "post_add", []string{"npm ci"})
	require.EqualError(t, err, "runner is required")

	err = RunAll(context.Background(), &fakeRunner{}, "", "post_add", []string{"npm ci"})
	require.EqualError(t, err, "hook directory is required")

	err = RunAll(context.Background(), &fakeRunner{}, "/repo/worktree", "post_add", []string{" "})
	require.EqualError(t, err, "post_add hook command 1 is required")
}
