package worktree

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type call struct {
	Dir  string
	Name string
	Args []string
}

type fakeRunner struct {
	Calls []call
	Err   error
}

func (f *fakeRunner) Run(_ context.Context, dir, name string, args ...string) error {
	f.Calls = append(f.Calls, call{Dir: dir, Name: name, Args: append([]string(nil), args...)})
	return f.Err
}

func TestAddCreatesBranch(t *testing.T) {
	gitDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(t.TempDir(), "feature")
	runner := &fakeRunner{}
	result, err := Add(context.Background(), runner, gitDir, worktreePath)
	require.NoError(t, err)
	require.Equal(t, worktreePath, result)
	require.Len(t, runner.Calls, 1)

	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", "-b", "feature", worktreePath},
	}, runner.Calls[0])
}

func TestRemoveWorktree(t *testing.T) {
	gitDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(t.TempDir(), "feature")
	require.NoError(t, os.MkdirAll(worktreePath, 0o755))

	runner := &fakeRunner{}
	result, err := Remove(context.Background(), runner, gitDir, worktreePath)
	require.NoError(t, err)
	require.Equal(t, worktreePath, result)
	require.Len(t, runner.Calls, 1)
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "remove", worktreePath},
	}, runner.Calls[0])
}
