package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	altsrc "github.com/urfave/cli-altsrc/v3"
)

const (
	defaultWorkdir = "."
	defaultBranch  = "main"
)

// Config defines the forest configuration file schema.
type Config struct {
	WorktreeRoot string       `toml:"worktree_root"`
	Hooks        *GlobalHooks `toml:"hooks,omitempty"`
	Projects     []Project    `toml:"projects"`
}

// GlobalHooks describes configurable shell hooks for global command lifecycle events.
type GlobalHooks struct {
	PostSwitch map[string]string `toml:"post-switch,omitempty"`
}

// Project describes a configured project.
type Project struct {
	Name          string        `toml:"name"`
	Repo          string        `toml:"repo"`
	Workdir       string        `toml:"workdir,omitempty"`
	DefaultBranch string        `toml:"default_branch,omitempty"`
	Hooks         *ProjectHooks `toml:"hooks,omitempty"`
}

// ProjectHooks describes configurable shell hooks for project lifecycle commands.
type ProjectHooks struct {
	PostAdd    map[string]string `toml:"post-add,omitempty"`
	PreRemove  map[string]string `toml:"pre-remove,omitempty"`
	PostRemove map[string]string `toml:"post-remove,omitempty"`
}

// Load reads the config from disk and validates it.
func Load(path string) (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}
	cfg, err := loadFromPath(path, homeDir)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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

// ProjectPath resolves the project directory on disk.
func ProjectPath(root, projectName string) string {
	if projectName == "" {
		return ""
	}
	if filepath.IsAbs(projectName) {
		return filepath.Clean(projectName)
	}
	return filepath.Join(root, projectName)
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
	if c.Hooks != nil {
		if err := validateHookCommands("global", "post-switch", c.Hooks.PostSwitch); err != nil {
			return err
		}
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
		if project.Hooks != nil {
			scope := fmt.Sprintf("project %q", project.Name)
			if err := validateHookCommands(scope, "post-add", project.Hooks.PostAdd); err != nil {
				return err
			}
			if err := validateHookCommands(scope, "pre-remove", project.Hooks.PreRemove); err != nil {
				return err
			}
			if err := validateHookCommands(scope, "post-remove", project.Hooks.PostRemove); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadFromPath(path, homeDir string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is required")
	}
	cache := altsrc.NewURISourceCache[Config](path, toml.Unmarshal)
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
	if c.Hooks != nil {
		c.Hooks.PostSwitch = trimHookCommands(c.Hooks.PostSwitch)
	}
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
			project.DefaultBranch = defaultBranch
		}
		if project.Hooks != nil {
			project.Hooks.PostAdd = trimHookCommands(project.Hooks.PostAdd)
			project.Hooks.PreRemove = trimHookCommands(project.Hooks.PreRemove)
			project.Hooks.PostRemove = trimHookCommands(project.Hooks.PostRemove)
		}
	}
	return nil
}

// resolvePath picks the config file path based on XDG or ~/.forest.toml.
func resolvePath(getenv func(string) string, homeDir string, exists func(string) bool) (string, error) {
	xdg := getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(homeDir, ".config")
	}
	xdgPath := filepath.Join(xdg, "forest", "config.toml")
	dotPath := filepath.Join(homeDir, ".forest.toml")

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

func trimHookCommands(commands map[string]string) map[string]string {
	for name, command := range commands {
		trimmedName := strings.TrimSpace(name)
		trimmedCommand := strings.TrimSpace(command)
		if trimmedName != name {
			delete(commands, name)
		}
		commands[trimmedName] = trimmedCommand
	}
	return commands
}

func validateHookCommands(scope, hookName string, commands map[string]string) error {
	for name, command := range commands {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("%s hook %s name is required", scope, hookName)
		}
		if strings.TrimSpace(command) == "" {
			return fmt.Errorf("%s hook %s %q command is required", scope, hookName, name)
		}
	}
	return nil
}
