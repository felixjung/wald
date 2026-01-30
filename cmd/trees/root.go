package main

import "github.com/urfave/cli/v3"

func newRootCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "trees",
		Usage: "Manage git worktrees for configured projects",
		Commands: []*cli.Command{
			newAddCommand(app),
			newRemoveCommand(app),
		},
	}
}
