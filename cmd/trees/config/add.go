package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	internalconfig "github.com/felixjung/trees/internal/config"
)

func newAddCommand(deps Deps) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a project to the trees config",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Aliases: []string{"r"}, Usage: "git repository", Required: true},
			&cli.StringFlag{Name: "workdir", Aliases: []string{"w"}, Usage: "relative workdir within the repo"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<project name>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.StringArg("name"))
			if name == "" {
				return cli.Exit("project name is required", 1)
			}
			repo := strings.TrimSpace(cmd.String("repo"))
			if repo == "" {
				return cli.Exit("repo is required", 1)
			}
			workdir := strings.TrimSpace(cmd.String("workdir"))
			if workdir != "" && filepath.IsAbs(workdir) {
				return cli.Exit("workdir must be relative", 1)
			}

			homeDir, err := deps.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home dir: %w", err)
			}
			configPath, err := ResolvePath(deps.Getenv, homeDir, deps.Stat)
			if err != nil {
				return err
			}
			cfg, err := internalconfig.Load(configPath)
			if err != nil {
				return err
			}
			if _, ok := cfg.FindProject(name); ok {
				return fmt.Errorf("project %q already exists", name)
			}

			cfg.Projects = append(cfg.Projects, internalconfig.Project{
				Name:    name,
				Repo:    repo,
				Workdir: workdir,
			})

			content, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("render config: %w", err)
			}
			if err := deps.WriteFile(configPath, content, 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			_, _ = fmt.Fprintln(deps.Stdout, configPath)
			return nil
		},
	}
}
