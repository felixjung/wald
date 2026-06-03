package main

import (
	"context"
	"fmt"
	"os"

	"github.com/felixjung/wald/internal/cliusage"
	"github.com/urfave/cli/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print version information",
		Action: func(_ context.Context, cmd *cli.Command) error {
			if err := cliusage.ValidateNoArgs(cmd); err != nil {
				return err
			}
			_, err := fmt.Fprintln(os.Stdout, renderVersion())
			return err
		},
	}
}

func renderVersion() string {
	return fmt.Sprintf("wald %s (commit %s, built %s)", version, commit, date)
}
