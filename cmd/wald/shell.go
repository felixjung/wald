package main

import (
	"context"

	"github.com/felixjung/wald/internal/cliusage"
	"github.com/urfave/cli/v3"
)

func newShellCommand() *cli.Command {
	return &cli.Command{
		Name:  "shell",
		Usage: "Shell integration helpers",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cliusage.ValidateNoArgs(cmd)
		},
		Commands: []*cli.Command{
			newShellInitCommand(),
		},
	}
}
