package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	temp := t.TempDir()
	xdg := filepath.Join(temp, "xdg")
	configPath := filepath.Join(xdg, "forest", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))

	content := `worktree_root: /tmp/worktrees
projects:
  - name: repo
    repo: github.com/felixjung/mono
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o644))

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/tmp/worktrees"), cfg.WorktreeRoot)

	project, ok := cfg.FindProject("repo")
	require.True(t, ok)
	require.Equal(t, "github.com/felixjung/mono", project.Repo)
	require.Equal(t, ".", project.Workdir)
	require.Equal(t, "main", project.DefaultBranch)
}

func TestRepoPath(t *testing.T) {
	require.Equal(t, filepath.FromSlash("/root/mono"), RepoPath("/root", "github.com/felixjung/mono"))
	require.Equal(t, filepath.FromSlash("/root/repo"), RepoPath("/root", "./repo"))
	require.Equal(t, "/abs/repo", RepoPath("/root", "/abs/repo"))
}

func TestProjectPath(t *testing.T) {
	require.Equal(t, filepath.FromSlash("/root/portier-backend"), ProjectPath("/root", "portier-backend"))
	require.Equal(t, "/abs/project", ProjectPath("/root", "/abs/project"))
	require.Empty(t, ProjectPath("/root", ""))
}

func TestFindProject(t *testing.T) {
	cfg := &Config{Projects: []Project{{Name: "repo", Repo: "github.com/felixjung/mono"}}}
	project, ok := cfg.FindProject("repo")
	require.True(t, ok)
	require.Equal(t, "repo", project.Name)
}
