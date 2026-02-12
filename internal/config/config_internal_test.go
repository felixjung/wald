package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePathPrefersXDG(t *testing.T) {
	temp := t.TempDir()
	home := filepath.Join(temp, "home")
	xdg := filepath.Join(temp, "xdg")
	require.NoError(t, os.MkdirAll(home, 0o755))
	require.NoError(t, os.MkdirAll(xdg, 0o755))

	xdgPath := filepath.Join(xdg, "forest", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(xdgPath), 0o755))
	require.NoError(t, os.WriteFile(xdgPath, []byte("worktree_root = \"/tmp\"\nprojects = []\n"), 0o644))

	dotPath := filepath.Join(home, ".forest.toml")
	require.NoError(t, os.WriteFile(dotPath, []byte("worktree_root = \"/tmp\"\nprojects = []\n"), 0o644))

	path, err := resolvePath(func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return xdg
		}
		return ""
	}, home, fileExists)
	require.NoError(t, err)
	require.Equal(t, xdgPath, path)
}

func TestResolvePathFallsBackToDotfile(t *testing.T) {
	temp := t.TempDir()
	home := filepath.Join(temp, "home")
	require.NoError(t, os.MkdirAll(home, 0o755))

	dotPath := filepath.Join(home, ".forest.toml")
	require.NoError(t, os.WriteFile(dotPath, []byte("worktree_root = \"/tmp\"\nprojects = []\n"), 0o644))

	path, err := resolvePath(func(string) string { return "" }, home, fileExists)
	require.NoError(t, err)
	require.Equal(t, dotPath, path)
}

func TestNormalizeRootExpandsHome(t *testing.T) {
	root, err := normalizeRoot("~/projects", "/home/test")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/home/test/projects"), root)
}

func TestLoadFromPath(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.toml")
	content := `worktree_root = "~/projects"

[[projects]]
name = "repo"
repo = "github.com/felixjung/mono"
workdir = "apps/repo"
default_branch = "develop"
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := loadFromPath(cfgPath, "/home/test")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/home/test/projects"), cfg.WorktreeRoot)
	require.Len(t, cfg.Projects, 1)
	require.Equal(t, "repo", cfg.Projects[0].Name)
	require.Equal(t, "github.com/felixjung/mono", cfg.Projects[0].Repo)
	require.Equal(t, "apps/repo", cfg.Projects[0].Workdir)
	require.Equal(t, "develop", cfg.Projects[0].DefaultBranch)
}

func TestLoadFromPathDefaults(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.toml")
	content := `worktree_root = "/tmp"

[[projects]]
name = "repo"
repo = "github.com/felixjung/mono"
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := loadFromPath(cfgPath, "/home/test")
	require.NoError(t, err)
	require.Equal(t, defaultWorkdir, cfg.Projects[0].Workdir)
	require.Equal(t, defaultBranch, cfg.Projects[0].DefaultBranch)
	require.Nil(t, cfg.Projects[0].Hooks)
}

func TestLoadFromPathWithHooks(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.toml")
	content := `worktree_root = "/tmp"

[[projects]]
name = "repo"
repo = "github.com/felixjung/mono"

[projects.hooks]
post_add = ["  npm ci  "]
pre_remove = ["echo before remove"]
post_remove = ["echo after remove"]
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := loadFromPath(cfgPath, "/home/test")
	require.NoError(t, err)
	require.NotNil(t, cfg.Projects[0].Hooks)
	require.Equal(t, []string{"npm ci"}, cfg.Projects[0].Hooks.PostAdd)
	require.Equal(t, []string{"echo before remove"}, cfg.Projects[0].Hooks.PreRemove)
	require.Equal(t, []string{"echo after remove"}, cfg.Projects[0].Hooks.PostRemove)
}

func TestLoadFromPathRejectsEmptyHookCommand(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.toml")
	content := `worktree_root = "/tmp"

[[projects]]
name = "repo"
repo = "github.com/felixjung/mono"

[projects.hooks]
post_add = ["", "npm ci"]
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	_, err := loadFromPath(cfgPath, "/home/test")
	require.EqualError(t, err, `project "repo" hook post_add command 1 is required`)
}
