package main

import (
	"context"
	"fmt"
	"os"

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
		Action: func(_ context.Context, _ *cli.Command) error {
			_, err := fmt.Fprintln(os.Stdout, renderVersion())
			return err
		},
	}
}

func renderVersion() string {
	return fmt.Sprintf("wald %s (commit %s, built %s)", version, commit, date)
}
