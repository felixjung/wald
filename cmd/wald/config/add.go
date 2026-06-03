package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felixjung/wald/internal/cliusage"
	internalconfig "github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

func newAddCommand(deps Deps) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a project to the wald config",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Aliases: []string{"r"}, Usage: "git repository"},
			&cli.StringFlag{Name: "workdir", Aliases: []string{"w"}, Usage: "relative workdir within the repo"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<project name>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			rawArgs := cmd.Args().Slice()
			name := firstArgValue(rawArgs, cmd.StringArg("name"))
			if err := validateNoExtraArgsAfterFirst(cmd, name); err != nil {
				return err
			}
			repo := strings.TrimSpace(cmd.String("repo"))
			workdir := strings.TrimSpace(cmd.String("workdir"))

			if name == "" || repo == "" {
				if !tui.IsTerminal(os.Stdin) {
					return cli.Exit("project name and repo are required", 1)
				}
				fields, err := tui.Prompt("Add project", []tui.Field{
					{ID: "name", Label: "Project name", Value: name, Required: true},
					{ID: "repo", Label: "Repository", Value: repo, Required: true},
					{ID: "workdir", Label: "Workdir", Placeholder: "relative path", Value: workdir, Default: "."},
				}, withThemeOptions(deps.ThemeOptions)...)
				if err != nil {
					if errors.Is(err, tui.ErrCanceled) {
						return cli.Exit("prompt canceled", 1)
					}
					return err
				}
				name = strings.TrimSpace(fieldValue(fields, "name"))
				repo = strings.TrimSpace(fieldValue(fields, "repo"))
				if field, ok := fieldByID(fields, "workdir"); ok {
					if field.UsedDefault {
						workdir = ""
					} else {
						workdir = strings.TrimSpace(field.Value)
					}
				}
			}

			if name == "" {
				return cli.Exit("project name is required", 1)
			}
			if repo == "" {
				return cli.Exit("repo is required", 1)
			}
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

			content, err := marshalConfig(cfg)
			if err != nil {
				return err
			}
			if err := deps.WriteFile(configPath, content, 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			_, _ = fmt.Fprintln(deps.Stdout, configPath)
			return nil
		},
	}
}

func validateNoExtraArgsAfterFirst(cmd *cli.Command, first string) error {
	args := argsAfterFirst(cmd.Args().Slice(), first)
	if len(args) == 0 {
		return nil
	}
	return cliusage.UnexpectedArgument(cmd, args[0])
}

func firstArgValue(args []string, value string) string {
	value = strings.TrimSpace(value)
	if value != "" || len(args) == 0 {
		return value
	}
	if strings.HasPrefix(args[0], "-") {
		return ""
	}
	return strings.TrimSpace(args[0])
}

func argsAfterFirst(args []string, first string) []string {
	first = strings.TrimSpace(first)
	if first == "" || len(args) == 0 {
		if len(args) > 1 {
			return args[1:]
		}
		return args
	}
	if strings.TrimSpace(args[0]) == first {
		return args[1:]
	}
	if len(args) > 1 {
		return args[1:]
	}
	return args
}
