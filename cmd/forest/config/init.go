package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	internalconfig "github.com/felixjung/forest/internal/config"
	"github.com/felixjung/forest/internal/tui"
)

func newInitCommand(deps Deps) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a forest config file",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "worktree-root", Aliases: []string{"r"}, Usage: "root folder for worktrees"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			root := strings.TrimSpace(cmd.String("worktree-root"))
			if root == "" {
				if !tui.IsTerminal(os.Stdin) {
					return cli.Exit("worktree-root is required", 1)
				}
				fields, err := tui.Prompt("Initialize config", []tui.Field{
					{ID: "root", Label: "Worktree root", Value: root, Required: true},
				})
				if err != nil {
					if errors.Is(err, tui.ErrCanceled) {
						return cli.Exit("prompt canceled", 1)
					}
					return err
				}
				root = strings.TrimSpace(fieldValue(fields, "root"))
			}
			if root == "" {
				return cli.Exit("worktree-root is required", 1)
			}
			homeDir, err := deps.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home dir: %w", err)
			}
			xdgPath, dotPath := Paths(deps.Getenv, homeDir)
			if fileExists(deps.Stat, xdgPath) {
				return fmt.Errorf("config already exists at %s", xdgPath)
			}
			if fileExists(deps.Stat, dotPath) {
				return fmt.Errorf("config already exists at %s", dotPath)
			}
			if err = deps.MkdirAll(filepath.Dir(xdgPath), 0o755); err != nil {
				return fmt.Errorf("create config directory: %w", err)
			}
			content, err := yaml.Marshal(internalconfig.Config{
				WorktreeRoot: root,
				Projects:     []internalconfig.Project{},
			})
			if err != nil {
				return fmt.Errorf("render config: %w", err)
			}
			if err = deps.WriteFile(xdgPath, content, 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			_, _ = fmt.Fprintln(deps.Stdout, xdgPath)
			return nil
		},
	}
}
