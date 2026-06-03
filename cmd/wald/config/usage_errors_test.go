package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigAddRejectsUnsupportedPositionalArg(t *testing.T) {
	deps := usageTestDeps()
	cmd := NewCommand(deps)

	err := cmd.Run(context.Background(), []string{"config", "add", "--repo", "git@example.com/repo.git", "repo", "extra"})

	require.Error(t, err)
	require.ErrorContains(t, err, `unexpected argument "extra"`)
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "config add")
}

func TestConfigCommandFormatsUnknownFlagWithCommandHelp(t *testing.T) {
	deps := usageTestDeps()
	cmd := NewCommand(deps)

	err := cmd.Run(context.Background(), []string{"config", "add", "--bogus"})

	require.Error(t, err)
	require.ErrorContains(t, err, "flag provided but not defined")
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "config add")
}

func TestConfigAddAcceptsSinglePositionalArg(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, ".wald.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("worktree_root = \"/tmp/worktrees\"\n"), 0o600))

	deps := usageTestDeps()
	deps.Getenv = func(string) string {
		return ""
	}
	deps.UserHomeDir = func() (string, error) {
		return home, nil
	}
	deps.Stat = os.Stat
	cmd := NewCommand(deps)

	err := cmd.Run(context.Background(), []string{"config", "add", "--repo", "git@example.com/repo.git", "repo"})

	require.NoError(t, err)
}

func TestConfigInitRejectsPositionalArg(t *testing.T) {
	deps := usageTestDeps()
	cmd := NewCommand(deps)

	err := cmd.Run(context.Background(), []string{"config", "init", "extra"})

	require.Error(t, err)
	require.ErrorContains(t, err, `unexpected argument "extra"`)
	require.ErrorContains(t, err, "USAGE")
	require.ErrorContains(t, err, "config init")
}

func usageTestDeps() Deps {
	return Deps{
		UserHomeDir: func() (string, error) {
			return "/tmp/home", nil
		},
		Stat: func(string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}
}
