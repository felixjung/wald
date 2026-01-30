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
	require.NoError(t, os.WriteFile(xdgPath, []byte("root: /tmp\nprojects: {}\n"), 0o644))

	dotPath := filepath.Join(home, ".trees.yaml")
	require.NoError(t, os.WriteFile(dotPath, []byte("root: /tmp\nprojects: {}\n"), 0o644))

	path, err := ResolvePath(func(key string) string {
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
	require.NoError(t, os.WriteFile(dotPath, []byte("root: /tmp\nprojects: {}\n"), 0o644))

	path, err := ResolvePath(func(string) string { return "" }, home, fileExists)
	require.NoError(t, err)
	require.Equal(t, dotPath, path)
}

func TestNormalizeRootExpandsHome(t *testing.T) {
	root, err := NormalizeRoot("~/projects", "/home/test")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/home/test/projects"), root)
}

func TestRepoPath(t *testing.T) {
	require.Equal(t, filepath.FromSlash("/root/mono"), RepoPath("/root", "github.com/felixjung/mono"))
	require.Equal(t, filepath.FromSlash("/root/repo"), RepoPath("/root", "./repo"))
	require.Equal(t, "/abs/repo", RepoPath("/root", "/abs/repo"))
}

func TestLoadProjectFromPath(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.yaml")
	content := `root: ~/projects
projects:
  repo:
    repo: github.com/felixjung/mono
    workdir: apps/repo
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := loadProjectFromPath(cfgPath, "/home/test", "repo")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/home/test/projects"), cfg.Root)
	require.Equal(t, "github.com/felixjung/mono", cfg.Project.Repo)
	require.Equal(t, "apps/repo", cfg.Project.Workdir)
}

func TestLoadProjectFromPathMissingProject(t *testing.T) {
	temp := t.TempDir()
	cfgPath := filepath.Join(temp, "config.yaml")
	content := `root: /tmp
projects:
  repo:
    repo: github.com/felixjung/mono
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	_, err := loadProjectFromPath(cfgPath, "/home/test", "missing")
	require.Error(t, err)
}
