package main

import (
	configcmd "github.com/felixjung/wald/cmd/wald/config"
	"github.com/urfave/cli/v3"
)

func newRootCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:  "wald",
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
