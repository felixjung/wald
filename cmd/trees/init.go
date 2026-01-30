package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func newInitCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize worktrees for configured projects",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return app.Init(ctx)
		},
	}
}
