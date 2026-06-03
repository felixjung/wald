package main

import (
	"strings"

	"github.com/felixjung/wald/internal/cliusage"
	"github.com/urfave/cli/v3"
)

func validateNoExtraArgsAfterFirst(cmd *cli.Command, first string) error {
	args := argsAfterFirst(cmd.Args().Slice(), first)
	if len(args) == 0 {
		return nil
	}
	return cliusage.UnexpectedArgument(cmd, args[0])
}

func firstArgValue(args []string, value string) string {
	value = strings.TrimSpace(value)
	if value != "" || len(args) == 0 {
		return value
	}
	if strings.HasPrefix(args[0], "-") {
		return ""
	}
	return strings.TrimSpace(args[0])
}

func argsAfterFirst(args []string, first string) []string {
	first = strings.TrimSpace(first)
	if first == "" || len(args) == 0 {
		if len(args) > 1 {
			return args[1:]
		}
		return args
	}
	if strings.TrimSpace(args[0]) == first {
		return args[1:]
	}
	if len(args) > 1 {
		return args[1:]
	}
	return args
}
