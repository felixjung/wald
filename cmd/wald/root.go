package main

import (
	"context"

	configcmd "github.com/felixjung/wald/cmd/wald/config"
	"github.com/felixjung/wald/internal/cliusage"
	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

func newRootCommand(app appAPI, themeProfile tui.ThemeProfile) *cli.Command {
	switchOptions := []tui.Option{tui.WithThemeProfile(themeProfile)}
	setSwitchThemeOptions(switchOptions...)

	return withUsageErrors(&cli.Command{
		Name:         "wald",
		Usage:        "Manage git worktrees for configured projects",
		OnUsageError: usageError,
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cliusage.ValidateNoArgs(cmd)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Print command output"},
		},
		Commands: []*cli.Command{
			newAddCommand(app, themeProfile),
			newInitCommand(app, themeProfile),
			newListCommand(app, themeProfile),
			newRemoveCommand(app),
			newSwitchCommand(app),
			newVersionCommand(),
			newShellCommand(),
			configcmd.NewCommand(configcmd.Deps{
				ThemeOptions: switchOptions,
			}),
		},
	})
}

func usageError(ctx context.Context, cmd *cli.Command, err error, _ bool) error {
	return cliusage.Error(ctx, cmd, err)
}

func withUsageErrors(cmd *cli.Command) *cli.Command {
	if cmd.OnUsageError == nil {
		cmd.OnUsageError = usageError
	}
	for _, subcommand := range cmd.Commands {
		withUsageErrors(subcommand)
	}
	return cmd
}
