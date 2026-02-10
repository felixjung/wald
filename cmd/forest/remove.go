package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/forest/internal/tui"
)

func newRemoveCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a worktree for a project",
		ArgsUsage: "<worktree> [-- <git worktree remove args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "worktree", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			project := strings.TrimSpace(cmd.String("project"))
			worktree := strings.TrimSpace(cmd.StringArg("worktree"))
			extraArgs := cmd.Args().Slice()

			if project == "" || worktree == "" {
				if !tui.IsTerminal(os.Stdin) {
					return cli.Exit("project name and worktree path are required", 1)
				}
				fields, err := tui.Prompt("Remove worktree", []tui.Field{
					{ID: "project", Label: "Project name", Value: project, Required: true},
					{ID: "worktree", Label: "Worktree path", Value: worktree, Required: true},
					{ID: "extra", Label: "Extra git args", Placeholder: "space-separated", Value: strings.Join(extraArgs, " ")},
				})
				if err != nil {
					if errors.Is(err, tui.ErrCanceled) {
						return cli.Exit("prompt canceled", 1)
					}
					return err
				}
				project = strings.TrimSpace(fieldValue(fields, "project"))
				worktree = strings.TrimSpace(fieldValue(fields, "worktree"))
				extraArgs = splitArgs(fieldValue(fields, "extra"))
			}

			if project == "" {
				return cli.Exit("project name is required", 1)
			}
			if worktree == "" {
				return cli.Exit("worktree path is required", 1)
			}
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			return app.Remove(ctx, project, worktree, extraArgs)
		},
	}
}
