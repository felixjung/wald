package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/felixjung/forest/internal/tui"
	"github.com/urfave/cli/v3"
)

func newAddCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a worktree for a project",
		ArgsUsage: "<path> [-- <git worktree add args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
			&cli.StringFlag{
				Name:  "base",
				Usage: "base reference for the new worktree (for example origin/feature/foo)",
			},
			&cli.BoolFlag{Name: "no-switch", Usage: "do not switch to the new worktree"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "path", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			project := strings.TrimSpace(cmd.String("project"))
			path := strings.TrimSpace(cmd.StringArg("path"))
			startPoint := strings.TrimSpace(cmd.String("base"))
			extraArgs := cmd.Args().Slice()
			noSwitch := cmd.Bool("no-switch")

			if project == "" {
				_, groups, err := app.List(ctx)
				if err != nil {
					return err
				}
				selectedProject, _, err := resolveProjectSelection("", groups, false)
				if err != nil {
					return handleSwitchSelectionError(err)
				}
				project = selectedProject
			}

			if path == "" {
				if !tui.IsTerminal(os.Stdin) {
					return cli.Exit("worktree path is required", 1)
				}
				fields, err := tui.Prompt("Add worktree", []tui.Field{
					{ID: "path", Label: "Worktree path", Value: path, Required: true},
					{ID: "extra", Label: "Extra git args", Placeholder: "space-separated", Value: strings.Join(extraArgs, " ")},
				})
				if err != nil {
					if errors.Is(err, tui.ErrCanceled) {
						return cli.Exit("prompt canceled", 1)
					}
					return err
				}
				path = strings.TrimSpace(fieldValue(fields, "path"))
				extraArgs = splitArgs(fieldValue(fields, "extra"))
			}

			if project == "" {
				return cli.Exit("project name is required", 1)
			}
			if path == "" {
				return cli.Exit("worktree path is required", 1)
			}
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			target, err := app.AddTarget(ctx, project, path, startPoint, extraArgs)
			if err != nil {
				return err
			}
			if noSwitch {
				_, err = fmt.Fprintln(os.Stdout, target)
				return err
			}
			switchTarget, err := app.SwitchTarget(ctx, project, path, "")
			if err != nil {
				return err
			}
			return writeSwitchTarget(switchTarget)
		},
	}
}
