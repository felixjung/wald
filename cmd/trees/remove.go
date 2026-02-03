package main

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
)

func newRemoveCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a worktree for a project",
		ArgsUsage: "<worktree> [-- <git worktree remove args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "worktree", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			worktree := strings.TrimSpace(cmd.StringArg("worktree"))
			if worktree == "" {
				return cli.Exit("worktree path is required", 1)
			}
			extraArgs := cmd.Args().Slice()
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			return app.Remove(ctx, cmd.String("project"), worktree, extraArgs)
		},
	}
}
