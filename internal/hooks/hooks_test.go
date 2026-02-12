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

	err := RunAll(context.Background(), runner, "/repo/worktree", "post-add", map[string]string{
		"01_project":  "echo {{project}}",
		"02_worktree": "echo {{ project }} {{worktree}}",
	}, Vars{"project": "repo", "worktree": "feature"})
	require.NoError(t, err)
	require.Equal(t, []call{
		{Dir: "/repo/worktree", Name: "sh", Args: []string{"-c", "echo repo"}},
		{Dir: "/repo/worktree", Name: "sh", Args: []string{"-c", "echo repo feature"}},
	}, runner.calls)
}

func TestRunAllFailsFast(t *testing.T) {
	runner := &fakeRunner{
		errs: map[int]error{
			1: errors.New("failed"),
		},
	}

	err := RunAll(context.Background(), runner, "/repo/worktree", "post-add", map[string]string{
		"01_bootstrap": "npm ci",
		"02_copy":      "cp .env.example .env",
		"03_make":      "make bootstrap",
	}, Vars{})
	require.EqualError(t, err, "post-add hook \"02_copy\" failed: failed")
	require.Len(t, runner.calls, 2)
}

func TestRunAllValidatesInputs(t *testing.T) {
	err := RunAll(context.Background(), nil, "/repo/worktree", "post-add", map[string]string{"hook": "npm ci"}, Vars{})
	require.EqualError(t, err, "runner is required")

	err = RunAll(context.Background(), &fakeRunner{}, "", "post-add", map[string]string{"hook": "npm ci"}, Vars{})
	require.EqualError(t, err, "hook directory is required")

	err = RunAll(context.Background(), &fakeRunner{}, "/repo/worktree", "post-add", map[string]string{"hook": " "}, Vars{})
	require.EqualError(t, err, `post-add hook "hook" command is required`)
}

func TestRunAllReturnsErrorForUnknownTemplateVariable(t *testing.T) {
	err := RunAll(context.Background(), &fakeRunner{}, "/repo/worktree", "post-add", map[string]string{"hook": "echo {{unknown}}"}, Vars{})
	require.EqualError(t, err, `post-add hook "hook" template error: unknown template variable "unknown"`)
}
