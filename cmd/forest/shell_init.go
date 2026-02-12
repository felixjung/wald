package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func newShellInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Print shell integration script",
		ArgsUsage: "<fish|zsh|bash>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "shell", UsageText: "<fish|zsh|bash>"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			shellName := strings.TrimSpace(cmd.StringArg("shell"))
			if shellName == "" {
				return cli.Exit("shell is required (fish|zsh|bash)", 1)
			}
			script, err := shellInitScript(shellName)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			_, err = fmt.Fprintln(os.Stdout, script)
			return err
		},
	}
}

func shellInitScript(shellName string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(shellName)) {
	case "fish":
		return fishShellInitScript, nil
	case "zsh":
		return zshShellInitScript, nil
	case "bash":
		return bashShellInitScript, nil
	default:
		return "", errors.New("unsupported shell: use fish, zsh, or bash")
	}
}

const fishShellInitScript = `function forest
    if test (count $argv) -gt 0; and contains -- "$argv[1]" switch add
        set -l __forest_target_file (mktemp -t forest-switch.XXXXXX)
        if test -z "$__forest_target_file"
            return 1
        end
        command env FOREST_SWITCH_OUT_FILE="$__forest_target_file" forest $argv
        set -l __forest_status $status
        if test $__forest_status -eq 0; and test -f "$__forest_target_file"
            set -l __forest_target (string trim (cat "$__forest_target_file"))
            if test -n "$__forest_target"
                cd "$__forest_target"
            end
        end
        rm -f "$__forest_target_file"
        if test $__forest_status -ne 0
            return $__forest_status
        end
        return 0
    end
    command forest $argv
end`

const zshShellInitScript = `forest() {
    if [[ "$1" == "switch" || "$1" == "add" ]]; then
        local __forest_target
        local __forest_target_file
        __forest_target_file="$(mktemp -t forest-switch.XXXXXX)" || return $?
        FOREST_SWITCH_OUT_FILE="$__forest_target_file" command forest "$@"
        local __forest_status=$?
        if [[ $__forest_status -eq 0 && -f "$__forest_target_file" ]]; then
            __forest_target="$(<"$__forest_target_file")"
            if [[ -n "$__forest_target" ]]; then
                cd "$__forest_target" || { rm -f "$__forest_target_file"; return $?; }
            fi
        fi
        rm -f "$__forest_target_file"
        return $__forest_status
    fi
    command forest "$@"
}`

const bashShellInitScript = `forest() {
    if [[ "$1" == "switch" || "$1" == "add" ]]; then
        local __forest_target
        local __forest_target_file
        __forest_target_file="$(mktemp -t forest-switch.XXXXXX)" || return $?
        FOREST_SWITCH_OUT_FILE="$__forest_target_file" command forest "$@"
        local __forest_status=$?
        if [[ $__forest_status -eq 0 && -f "$__forest_target_file" ]]; then
            __forest_target="$(<"$__forest_target_file")"
            if [[ -n "$__forest_target" ]]; then
                cd "$__forest_target" || { rm -f "$__forest_target_file"; return $?; }
            fi
        fi
        rm -f "$__forest_target_file"
        return $__forest_status
    fi
    command forest "$@"
}`
