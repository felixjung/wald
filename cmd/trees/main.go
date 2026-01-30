package main

import (
	"context"
	"os"

	"github.com/felixjung/trees/runner"
)

func main() {
	cmd := newRootCommand(commandDeps{
		Runner: runner.OSRunner{Stdout: os.Stdout, Stderr: os.Stderr},
		Stdout: os.Stdout,
	})
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
