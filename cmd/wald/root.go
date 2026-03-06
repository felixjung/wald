package main

import (
	configcmd "github.com/felixjung/wald/cmd/wald/config"
	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

func newRootCommand(app appAPI, themeProfile tui.ThemeProfile) *cli.Command {
	switchOptions := []tui.Option{tui.WithThemeProfile(themeProfile)}
	setSwitchThemeOptions(switchOptions...)

	return &cli.Command{
		Name:  "wald",
		Usage: "Manage git worktrees for configured projects",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Print command output"},
		},
		Commands: []*cli.Command{
			newAddCommand(app, themeProfile),
			newInitCommand(app),
			newListCommand(app, themeProfile),
			newRemoveCommand(app),
			newSwitchCommand(app),
			newVersionCommand(),
			newShellCommand(),
			configcmd.NewCommand(configcmd.Deps{
				ThemeOptions: switchOptions,
			}),
		},
	}
}
