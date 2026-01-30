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

func TestAddCreatesMainAndBranch(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	runner := &fakeRunner{}
	err := Add(context.Background(), runner, repo, "feature")
	require.NoError(t, err)
	require.Len(t, runner.Calls, 2)

	mainPath := filepath.Join(repo, "main")
	branchPath := filepath.Join(repo, "feature")

	require.Equal(t, call{
		Dir:  repo,
		Name: "git",
		Args: []string{"worktree", "add", "-B", "main", mainPath, "main"},
	}, runner.Calls[0])
	require.Equal(t, call{
		Dir:  repo,
		Name: "git",
		Args: []string{"worktree", "add", "-b", "feature", branchPath, "main"},
	}, runner.Calls[1])
}

func TestRemoveWorktree(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	branchPath := filepath.Join(repo, "feature")
	require.NoError(t, os.Mkdir(branchPath, 0o755))

	runner := &fakeRunner{}
	err := Remove(context.Background(), runner, repo, "feature")
	require.NoError(t, err)
	require.Len(t, runner.Calls, 1)
	require.Equal(t, call{
		Dir:  repo,
		Name: "git",
		Args: []string{"worktree", "remove", branchPath},
	}, runner.Calls[0])
}

func TestWorkdirPath(t *testing.T) {
	repo := filepath.FromSlash("/repo")
	require.Equal(t, filepath.FromSlash("/repo/branch"), WorkdirPath(repo, "branch", ""))
	require.Equal(t, filepath.FromSlash("/repo/branch/apps/repo"), WorkdirPath(repo, "branch", "apps/repo"))
}
