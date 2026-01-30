package main

import (
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
	"github.com/felixjung/trees/runner"
)

type commandDeps struct {
	ConfigLoader func() (*config.Config, string, error)
	Runner       worktree.Runner
	Stdout       io.Writer
}

func newRootCommand(deps commandDeps) *cli.Command {
	if deps.ConfigLoader == nil {
		deps.ConfigLoader = config.Load
	}
	if deps.Stdout == nil {
		deps.Stdout = os.Stdout
	}
	if deps.Runner == nil {
		deps.Runner = runner.OSRunner{Stdout: os.Stdout, Stderr: os.Stderr}
	}

	return &cli.Command{
		Name:  "trees",
		Usage: "Manage git worktrees for configured projects",
		Commands: []*cli.Command{
			newAddCommand(deps),
			newRemoveCommand(deps),
		},
	}
}
