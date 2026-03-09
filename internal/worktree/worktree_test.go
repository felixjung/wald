package worktree

import (
	"context"
	"errors"
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
	RunFn func(context.Context, string, string, ...string) error
}

func (f *fakeRunner) Run(ctx context.Context, dir, name string, args ...string) error {
	f.Calls = append(f.Calls, call{Dir: dir, Name: name, Args: append([]string(nil), args...)})
	if f.RunFn != nil {
		return f.RunFn(ctx, dir, name, args...)
	}
	return f.Err
}

func TestAddCreatesWorktree(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)
	runner := &fakeRunner{}
	result, err := Add(context.Background(), runner, gitDir, relativePath, "feature", "", nil)
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 1)

	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", relativePath},
	}, runner.Calls[0])
}

func TestRemoveWorktree(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)

	runner := &fakeRunner{}
	result, err := Remove(context.Background(), runner, gitDir, relativePath, nil)
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 1)
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "remove", relativePath},
	}, runner.Calls[0])
}

func TestAddForwardsExtraArgs(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)
	runner := &fakeRunner{}
	result, err := Add(context.Background(), runner, gitDir, relativePath, "feature", "", []string{"--", "--force"})
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 1)

	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", "--force", relativePath},
	}, runner.Calls[0])
}

func TestRemoveForwardsExtraArgs(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)

	runner := &fakeRunner{}
	result, err := Remove(context.Background(), runner, gitDir, relativePath, []string{"--", "--force"})
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 1)
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "remove", "--force", relativePath},
	}, runner.Calls[0])
}

func TestSplitExtraArgsRequiresSeparator(t *testing.T) {
	_, err := splitExtraArgs([]string{"--force"})
	require.Error(t, err)
}

func TestAddAcceptsStartPoint(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)
	runner := &fakeRunner{}
	result, err := Add(context.Background(), runner, gitDir, relativePath, "feature", "origin/feature/foo", nil)
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 1)

	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", "-b", "feature", relativePath, "origin/feature/foo"},
	}, runner.Calls[0])
}

func TestAddRequiresBranchWhenStartPointIsSet(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)

	_, err = Add(context.Background(), &fakeRunner{}, gitDir, relativePath, "", "origin/main", nil)
	require.EqualError(t, err, "branch is required when start point is set")
}

func TestAddUsesExistingLocalBranchWhenStartPointResolvesToBranch(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)

	runner := &fakeRunner{
		RunFn: func(_ context.Context, _ string, _ string, args ...string) error {
			if len(args) >= 4 && args[0] == "show-ref" && args[1] == "--verify" && args[2] == "--quiet" {
				return nil
			}
			return nil
		},
	}
	result, err := Add(context.Background(), runner, gitDir, relativePath, "feature", "origin/feature", nil)
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 2)
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"},
	}, runner.Calls[0])
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", relativePath},
	}, runner.Calls[1])
}

func TestAddCreatesBranchWhenStartPointBranchDoesNotExistLocally(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, "main")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(gitDir, ".git"), 0o755))

	worktreePath := filepath.Join(root, "feature")
	relativePath, err := filepath.Rel(gitDir, worktreePath)
	require.NoError(t, err)

	runner := &fakeRunner{
		RunFn: func(_ context.Context, _ string, _ string, args ...string) error {
			if len(args) >= 4 && args[0] == "show-ref" && args[1] == "--verify" && args[2] == "--quiet" {
				return errors.New("not found")
			}
			return nil
		},
	}
	result, err := Add(context.Background(), runner, gitDir, relativePath, "feature", "origin/feature", nil)
	require.NoError(t, err)
	require.Equal(t, relativePath, result)
	require.Len(t, runner.Calls, 2)
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"},
	}, runner.Calls[0])
	require.Equal(t, call{
		Dir:  gitDir,
		Name: "git",
		Args: []string{"worktree", "add", "-b", "feature", relativePath, "origin/feature"},
	}, runner.Calls[1])
}
