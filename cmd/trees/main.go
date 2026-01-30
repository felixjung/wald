package main

import (
	"context"
	"os"

	"github.com/felixjung/trees/internal/app"
	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/runner"
)

type appAPI interface {
	Add(ctx context.Context, projectName, branch string) error
	Remove(ctx context.Context, projectName, branch string) error
}

func main() {
	cfg, _, err := config.Load()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	application, err := app.New(app.Deps{Runner: runner.OSRunner{Stdout: os.Stdout, Stderr: os.Stderr}, Stdout: os.Stdout}, cfg)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	api, ok := any(application).(appAPI)
	if !ok {
		_, _ = os.Stderr.WriteString("app does not implement required interface\n")
		os.Exit(1)
	}

	cmd := newRootCommand(api)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
