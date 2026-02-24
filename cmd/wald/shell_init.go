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

const fishShellInitScript = `function wald
    if test (count $argv) -gt 0; and contains -- "$argv[1]" switch add remove
        set -l __wald_target_file (mktemp -t wald-switch.XXXXXX)
        if test -z "$__wald_target_file"
            return 1
        end
        command env WALD_SWITCH_OUT_FILE="$__wald_target_file" wald $argv
        set -l __wald_status $status
        if test $__wald_status -eq 0; and test -f "$__wald_target_file"
            set -l __wald_target (string trim (cat "$__wald_target_file"))
            if test -n "$__wald_target"
                cd "$__wald_target"
            end
        end
        rm -f "$__wald_target_file"
        if test $__wald_status -ne 0
            return $__wald_status
        end
        return 0
    end
    command wald $argv
end`

const zshShellInitScript = `wald() {
    if [[ "$1" == "switch" || "$1" == "add" || "$1" == "remove" ]]; then
        local __wald_target
        local __wald_target_file
        __wald_target_file="$(mktemp -t wald-switch.XXXXXX)" || return $?
        WALD_SWITCH_OUT_FILE="$__wald_target_file" command wald "$@"
        local __wald_status=$?
        if [[ $__wald_status -eq 0 && -f "$__wald_target_file" ]]; then
            __wald_target="$(<"$__wald_target_file")"
            if [[ -n "$__wald_target" ]]; then
                cd "$__wald_target" || { rm -f "$__wald_target_file"; return $?; }
            fi
        fi
        rm -f "$__wald_target_file"
        return $__wald_status
    fi
    command wald "$@"
}`

const bashShellInitScript = `wald() {
    if [[ "$1" == "switch" || "$1" == "add" || "$1" == "remove" ]]; then
        local __wald_target
        local __wald_target_file
        __wald_target_file="$(mktemp -t wald-switch.XXXXXX)" || return $?
        WALD_SWITCH_OUT_FILE="$__wald_target_file" command wald "$@"
        local __wald_status=$?
        if [[ $__wald_status -eq 0 && -f "$__wald_target_file" ]]; then
            __wald_target="$(<"$__wald_target_file")"
            if [[ -n "$__wald_target" ]]; then
                cd "$__wald_target" || { rm -f "$__wald_target_file"; return $?; }
            fi
        fi
        rm -f "$__wald_target_file"
        return $__wald_status
    fi
    command wald "$@"
}`
