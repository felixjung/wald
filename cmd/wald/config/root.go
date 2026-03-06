package config

import (
	"io"
	"os"

	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

type Deps struct {
	Stdout       io.Writer
	Stderr       io.Writer
	Getenv       func(string) string
	UserHomeDir  func() (string, error)
	Stat         func(string) (os.FileInfo, error)
	MkdirAll     func(string, os.FileMode) error
	WriteFile    func(string, []byte, os.FileMode) error
	ThemeOptions []tui.Option
}

func NewCommand(deps Deps) *cli.Command {
	if deps.Stdout == nil {
		deps.Stdout = os.Stdout
	}
	if deps.Stderr == nil {
		deps.Stderr = os.Stderr
	}
	if deps.Getenv == nil {
		deps.Getenv = os.Getenv
	}
	if deps.UserHomeDir == nil {
		deps.UserHomeDir = os.UserHomeDir
	}
	if deps.Stat == nil {
		deps.Stat = os.Stat
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.WriteFile == nil {
		deps.WriteFile = os.WriteFile
	}

	return &cli.Command{
		Name:  "config",
		Usage: "Manage wald configuration",
		Commands: []*cli.Command{
			newAddCommand(deps),
			newInitCommand(deps),
		},
	}
}
