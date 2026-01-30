package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
)

// Project describes a configured project.
type Project struct {
	// Repo is a local path (absolute or relative to Root) or a repo identifier
	// like github.com/owner/name.
	Repo string
	// Workdir is an optional subfolder within a worktree for mono repos.
	Workdir string
}

// ProjectConfig contains config data for a single project.
type ProjectConfig struct {
	Root        string
	ProjectName string
	Project     Project
}

// LoadProject reads config and returns the requested project configuration.
func LoadProject(projectName string) (*ProjectConfig, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("resolve home dir: %w", err)
	}
	path, err := ResolvePath(os.Getenv, homeDir, fileExists)
	if err != nil {
		return nil, "", err
	}
	cfg, err := loadProjectFromPath(path, homeDir, projectName)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func loadProjectFromPath(path, homeDir, projectName string) (*ProjectConfig, error) {
	if strings.TrimSpace(projectName) == "" {
		return nil, errors.New("project name is required")
	}
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is required")
	}
	source := altsrc.StringSourcer(path)

	root, ok := lookupString(source, "root")
	if !ok || root == "" {
		return nil, errors.New("config root is required")
	}
	root, err := NormalizeRoot(root, homeDir)
	if err != nil {
		return nil, err
	}

	repoKey := fmt.Sprintf("projects.%s.repo", projectName)
	repo, ok := lookupString(source, repoKey)
	if !ok || repo == "" {
		return nil, fmt.Errorf("project %q repo is required", projectName)
	}

	workdirKey := fmt.Sprintf("projects.%s.workdir", projectName)
	workdir, _ := lookupString(source, workdirKey)
	if filepath.IsAbs(workdir) {
		return nil, fmt.Errorf("project %q workdir must be relative", projectName)
	}

	return &ProjectConfig{
		Root:        root,
		ProjectName: projectName,
		Project: Project{
			Repo:    repo,
			Workdir: workdir,
		},
	}, nil
}

func lookupString(source altsrc.Sourcer, key string) (string, bool) {
	value, ok := altsrcyaml.YAML(key, source).Lookup()
	if !ok {
		return "", false
	}
	return strings.TrimSpace(value), true
}

// ResolvePath picks the config file path based on XDG or ~/.trees.yaml.
func ResolvePath(getenv func(string) string, homeDir string, exists func(string) bool) (string, error) {
	xdg := getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(homeDir, ".config")
	}
	xdgPath := filepath.Join(xdg, "trees", "config.yaml")
	dotPath := filepath.Join(homeDir, ".trees.yaml")

	if exists(xdgPath) {
		return xdgPath, nil
	}
	if exists(dotPath) {
		return dotPath, nil
	}
	return "", fmt.Errorf("config not found; create %s or %s", xdgPath, dotPath)
}

// NormalizeRoot expands ~ in root and cleans it.
func NormalizeRoot(root, homeDir string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("root is required")
	}
	if strings.HasPrefix(root, "~") {
		if homeDir == "" {
			return "", errors.New("home dir unavailable for ~ expansion")
		}
		root = filepath.Join(homeDir, strings.TrimPrefix(root, "~"))
	}
	return filepath.Clean(root), nil
}

// RepoPath resolves the repository root directory on disk.
func RepoPath(root, repo string) string {
	if repo == "" {
		return ""
	}
	if filepath.IsAbs(repo) || strings.HasPrefix(repo, ".") {
		if filepath.IsAbs(repo) {
			return filepath.Clean(repo)
		}
		return filepath.Clean(filepath.Join(root, repo))
	}
	base := RepoDirName(repo)
	return filepath.Join(root, base)
}

// RepoDirName derives a folder name from a repo identifier.
func RepoDirName(repo string) string {
	repo = strings.TrimRight(repo, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if repo == "" {
		return ""
	}
	return path.Base(repo)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
