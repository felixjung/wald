package cliusage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

// Error returns a usage error followed by help for cmd.
func Error(_ context.Context, cmd *cli.Command, err error) error {
	if err == nil {
		return nil
	}
	return messageWithHelp(cmd, err.Error())
}

// UnexpectedArgument returns a usage error for an unsupported positional argument.
func UnexpectedArgument(cmd *cli.Command, arg string) error {
	return messageWithHelp(cmd, fmt.Sprintf("unexpected argument %q", arg))
}

// ValidateNoArgs rejects positional arguments for commands that do not accept them.
func ValidateNoArgs(cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return nil
	}
	return UnexpectedArgument(cmd, args[0])
}

func messageWithHelp(cmd *cli.Command, message string) error {
	message = strings.TrimSpace(message)
	help := strings.TrimSpace(renderHelp(cmd))
	if help == "" {
		return errors.New(message)
	}
	return fmt.Errorf("%s\n\n%s", message, help)
}

func renderHelp(cmd *cli.Command) string {
	var output bytes.Buffer
	template := cmd.CustomHelpTemplate
	if template == "" {
		switch {
		case cmd.Root() == cmd:
			template = cli.RootCommandHelpTemplate
		case len(cmd.Commands) > 0:
			template = cli.SubcommandHelpTemplate
		default:
			template = cli.CommandHelpTemplate
		}
	}
	cli.HelpPrinter(&output, template, cmd)
	return output.String()
}
