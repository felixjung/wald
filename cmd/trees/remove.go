package main

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/trees/internal/app"
)

func newRemoveCommand(deps commandDeps) *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove a worktree for a project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "branch name", Required: true},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, _, err := deps.ConfigLoader()
			if err != nil {
				return err
			}
			return app.RunRemove(ctx, app.Deps{Runner: deps.Runner, Stdout: deps.Stdout}, cfg, cmd.String("project"), cmd.String("name"))
		},
	}
}
