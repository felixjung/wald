package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/trees/internal/tui"
)

func newAddCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a worktree for a project",
		ArgsUsage: "<path> [-- <git worktree add args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "path", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			project := strings.TrimSpace(cmd.String("project"))
			path := strings.TrimSpace(cmd.StringArg("path"))
			extraArgs := cmd.Args().Slice()

			if project == "" || path == "" {
				if !tui.IsTerminal(os.Stdin) {
					return cli.Exit("project name and worktree path are required", 1)
				}
				fields, err := tui.Prompt("Add worktree", []tui.Field{
					{ID: "project", Label: "Project name", Value: project, Required: true},
					{ID: "path", Label: "Worktree path", Value: path, Required: true},
					{ID: "extra", Label: "Extra git args", Placeholder: "space-separated", Value: strings.Join(extraArgs, " ")},
				})
				if err != nil {
					if errors.Is(err, tui.ErrCanceled) {
						return cli.Exit("prompt canceled", 1)
					}
					return err
				}
				project = strings.TrimSpace(fieldValue(fields, "project"))
				path = strings.TrimSpace(fieldValue(fields, "path"))
				extraArgs = splitArgs(fieldValue(fields, "extra"))
			}

			if project == "" {
				return cli.Exit("project name is required", 1)
			}
			if path == "" {
				return cli.Exit("worktree path is required", 1)
			}
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			return app.Add(ctx, project, path, extraArgs)
		},
	}
}
