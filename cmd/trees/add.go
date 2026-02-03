package main

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
)

func newAddCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a worktree for a project",
		ArgsUsage: "<path> [-- <git worktree add args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "path", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path := strings.TrimSpace(cmd.StringArg("path"))
			if path == "" {
				return cli.Exit("worktree path is required", 1)
			}
			extraArgs := cmd.Args().Slice()
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			return app.Add(ctx, cmd.String("project"), path, extraArgs)
		},
	}
}
