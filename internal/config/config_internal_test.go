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

	xdgPath := filepath.Join(xdg, "trees", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(xdgPath), 0o755))
	require.NoError(t, os.WriteFile(xdgPath, []byte("worktree_root: /tmp\nprojects: []\n"), 0o644))

	dotPath := filepath.Join(home, ".trees.yaml")
	require.NoError(t, os.WriteFile(dotPath, []byte("worktree_root: /tmp\nprojects: []\n"), 0o644))

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

	dotPath := filepath.Join(home, ".trees.yaml")
	require.NoError(t, os.WriteFile(dotPath, []byte("worktree_root: /tmp\nprojects: []\n"), 0o644))

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
	cfgPath := filepath.Join(temp, "config.yaml")
	content := `worktree_root: ~/projects
projects:
  - name: repo
    repo: github.com/felixjung/mono
    workdir: apps/repo
    default_branch: develop
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
	cfgPath := filepath.Join(temp, "config.yaml")
	content := `worktree_root: /tmp
projects:
  - name: repo
    repo: github.com/felixjung/mono
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := loadFromPath(cfgPath, "/home/test")
	require.NoError(t, err)
	require.Equal(t, defaultWorkdir, cfg.Projects[0].Workdir)
	require.Equal(t, defaultDefaultBranch, cfg.Projects[0].DefaultBranch)
}
