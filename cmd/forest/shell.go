package main

import "github.com/urfave/cli/v3"

func newShellCommand() *cli.Command {
	return &cli.Command{
		Name:  "shell",
		Usage: "Shell integration helpers",
		Commands: []*cli.Command{
			newShellInitCommand(),
		},
	}
}
