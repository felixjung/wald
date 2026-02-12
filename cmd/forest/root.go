package main

import (
	"github.com/urfave/cli/v3"

	configcmd "github.com/felixjung/forest/cmd/forest/config"
)

func newRootCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "forest",
		Usage: "Manage git worktrees for configured projects",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Print command output"},
		},
		Commands: []*cli.Command{
			newAddCommand(app),
			newInitCommand(app),
			newListCommand(app),
			newRemoveCommand(app),
			newSwitchCommand(app),
			newVersionCommand(),
			newShellCommand(),
			configcmd.NewCommand(configcmd.Deps{}),
		},
	}
}
