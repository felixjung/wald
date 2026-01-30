package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"gopkg.in/yaml.v3"
)

const (
	defaultWorkdir       = "."
	defaultDefaultBranch = "main"
)

// Config defines the trees configuration file schema.
type Config struct {
	WorktreeRoot string    `yaml:"worktree_root"`
	Projects     []Project `yaml:"projects"`
}

// Project describes a configured project.
type Project struct {
	Name          string `yaml:"name"`
	Repo          string `yaml:"repo"`
	Workdir       string `yaml:"workdir"`
	DefaultBranch string `yaml:"default_branch"`
}

// Load reads the config from disk and validates it.
func Load() (*Config, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("resolve home dir: %w", err)
	}
	path, err := resolvePath(os.Getenv, homeDir, fileExists)
	if err != nil {
		return nil, "", err
	}
	cfg, err := loadFromPath(path, homeDir)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

// FindProject finds a project by name.
func (c *Config) FindProject(name string) (Project, bool) {
	for _, project := range c.Projects {
		if project.Name == name {
			return project, true
		}
	}
	return Project{}, false
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
	base := repoDirName(repo)
	return filepath.Join(root, base)
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.WorktreeRoot) == "" {
		return errors.New("config worktree_root is required")
	}
	if len(c.Projects) == 0 {
		return errors.New("config projects is required")
	}
	for _, project := range c.Projects {
		if project.Name == "" {
			return errors.New("project name is required")
		}
		if project.Repo == "" {
			return fmt.Errorf("project %q repo is required", project.Name)
		}
		if filepath.IsAbs(project.Workdir) {
			return fmt.Errorf("project %q workdir must be relative", project.Name)
		}
		if project.DefaultBranch == "" {
			return fmt.Errorf("project %q default_branch is required", project.Name)
		}
	}
	return nil
}

func loadFromPath(path, homeDir string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is required")
	}
	cache := altsrc.NewURISourceCache[Config](path, yaml.Unmarshal)
	cfg := cache.Get()
	if err := cfg.normalize(homeDir); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// normalize applies defaults and expands ~ in WorktreeRoot.
func (c *Config) normalize(homeDir string) error {
	root, err := normalizeRoot(c.WorktreeRoot, homeDir)
	if err != nil {
		return err
	}
	c.WorktreeRoot = root
	for i := range c.Projects {
		project := &c.Projects[i]
		project.Name = strings.TrimSpace(project.Name)
		project.Repo = strings.TrimSpace(project.Repo)
		project.Workdir = strings.TrimSpace(project.Workdir)
		project.DefaultBranch = strings.TrimSpace(project.DefaultBranch)
		if project.Workdir == "" {
			project.Workdir = defaultWorkdir
		}
		if project.DefaultBranch == "" {
			project.DefaultBranch = defaultDefaultBranch
		}
	}
	return nil
}

// resolvePath picks the config file path based on XDG or ~/.trees.yaml.
func resolvePath(getenv func(string) string, homeDir string, exists func(string) bool) (string, error) {
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

// normalizeRoot expands ~ in root and cleans it.
func normalizeRoot(root, homeDir string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("config worktree_root is required")
	}
	if strings.HasPrefix(root, "~") {
		if homeDir == "" {
			return "", errors.New("home dir unavailable for ~ expansion")
		}
		root = filepath.Join(homeDir, strings.TrimPrefix(root, "~"))
	}
	return filepath.Clean(root), nil
}

// repoDirName derives a folder name from a repo identifier.
func repoDirName(repo string) string {
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
