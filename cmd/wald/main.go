package main

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"

	configcmd "github.com/felixjung/wald/cmd/wald/config"
	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/runner"
)

type appAPI interface {
	AddTarget(ctx context.Context, projectName, branch, startPoint string, extraArgs []string) (string, error)
	Init(ctx context.Context) error
	List(ctx context.Context) (string, []app.ProjectWorktrees, error)
	Remove(ctx context.Context, projectName, branch string, extraArgs []string) error
	SwitchTarget(ctx context.Context, projectName, worktree, workingDirOverride string) (string, error)
}

func main() {
	verbose := hasVerboseFlag(os.Args[1:])
	quietOut := io.Discard
	quietErr := io.Discard
	if verbose {
		quietOut = os.Stdout
		quietErr = os.Stderr
	}

	cfg, err := loadConfigIfNeeded(os.Args[1:])
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	application, err := app.New(app.Deps{Runner: runner.OSRunner{Stdout: quietOut, Stderr: quietErr}, Stdout: os.Stdout}, cfg)
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
	if command == "config" || command == "shell" || command == "" {
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

func hasVerboseFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "-v" || arg == "--verbose" {
			return true
		}
		if strings.HasPrefix(arg, "--verbose=") {
			value := strings.TrimPrefix(arg, "--verbose=")
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return true
			}
			return parsed
		}
		if strings.HasPrefix(arg, "-v=") {
			value := strings.TrimPrefix(arg, "-v=")
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return true
			}
			return parsed
		}
	}
	return false
}
