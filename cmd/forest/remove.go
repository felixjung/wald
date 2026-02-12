package main

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/forest/internal/app"
)

func newRemoveCommand(app appAPI) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a worktree for a project",
		ArgsUsage: "<worktree> [-- <git worktree remove args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "worktree", UsageText: "<worktree path>"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			project := strings.TrimSpace(cmd.String("project"))
			worktree := strings.TrimSpace(cmd.StringArg("worktree"))
			extraArgs := cmd.Args().Slice()

			if project == "" || worktree == "" {
				_, groups, err := app.List(ctx)
				if err != nil {
					return err
				}
				project, worktree, err = resolveRemoveSelection(project, worktree, groups)
				if err != nil {
					return handleSwitchSelectionError(err)
				}
			}

			if project == "" {
				return cli.Exit("project name is required", 1)
			}
			if worktree == "" {
				return cli.Exit("worktree path is required", 1)
			}
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			return app.Remove(ctx, project, worktree, extraArgs)
		},
	}
}

func resolveRemoveSelection(project, worktree string, groups []app.ProjectWorktrees) (string, string, error) {
	project = strings.TrimSpace(project)
	worktree = strings.TrimSpace(worktree)

	if project == "" {
		project = inferProjectNameFromCurrentWorktree(groups, true)
	}
	if project == "" || worktree == "" {
		selectedProject, group, err := resolveProjectSelection(project, groups, true)
		if err != nil {
			return "", "", err
		}
		project = selectedProject

		selectedWorktree, _, err := resolveWorktreeSelection(group, worktree, false)
		if err != nil {
			return "", "", err
		}
		worktree = selectedWorktree
	}

	return project, worktree, nil
}
