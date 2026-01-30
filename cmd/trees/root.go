package main

import (
	"github.com/urfave/cli/v3"

	configcmd "github.com/felixjung/trees/cmd/trees/config"
)

func newRootCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "trees",
		Usage: "Manage git worktrees for configured projects",
		Commands: []*cli.Command{
			newAddCommand(app),
			newInitCommand(app),
			newRemoveCommand(app),
			configcmd.NewCommand(configcmd.Deps{}),
		},
	}
}
