package main

import (
	"context"
	"os"

	configcmd "github.com/felixjung/trees/cmd/trees/config"
	"github.com/felixjung/trees/internal/app"
	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/runner"
)

type appAPI interface {
	Add(ctx context.Context, projectName, branch string) error
	Remove(ctx context.Context, projectName, branch string) error
}

func main() {
	cfg, err := loadConfigIfNeeded(os.Args[1:])
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

func loadConfigIfNeeded(args []string) (*config.Config, error) {
	command := firstNonFlagArg(args)
	if command == "config" || command == "" {
		return &config.Config{}, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath, err := configcmd.ResolvePath(os.Getenv, homeDir, os.Stat)
	if err != nil {
		return nil, err
	}
	return config.Load(configPath)
}

func firstNonFlagArg(args []string) string {
	for _, arg := range args {
		if arg == "--" {
			return ""
		}
		if arg == "" || arg[0] == '-' {
			continue
		}
		return arg
	}
	return ""
}
