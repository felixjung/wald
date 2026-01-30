package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func newAddCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a worktree for a project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "branch name", Required: true},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return app.Add(ctx, cmd.String("project"), cmd.String("name"))
		},
	}
}
