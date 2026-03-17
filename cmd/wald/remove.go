package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felixjung/wald/internal/app"
	"github.com/urfave/cli/v3"
)

func newRemoveCommand(api appAPI) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a worktree for a project",
		ArgsUsage: "<worktree> [-- <git worktree remove args>]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
			&cli.BoolFlag{Name: "force", Usage: "force removal even if the worktree is dirty"},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "args", Max: -1, UsageText: "<worktree> [-- <git worktree remove args>]"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			project := strings.TrimSpace(cmd.String("project"))
			worktree, extraArgs := parseRemoveArgs(cmd.StringArgs("args"))
			var groups []app.ProjectWorktrees

			if project == "" || worktree == "" {
				_, listedGroups, err := api.List(ctx)
				if err != nil {
					return err
				}
				groups = listedGroups
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
			if cmd.Bool("force") {
				extraArgs = append(extraArgs, "--force")
			}
			if len(extraArgs) > 0 {
				extraArgs = append([]string{"--"}, extraArgs...)
			}
			if err := api.Remove(ctx, project, worktree, extraArgs); err != nil {
				return err
			}
			return writeRemoveSwitchTarget(ctx, api, project, worktree, groups)
		},
	}
}

func parseRemoveArgs(args []string) (worktree string, extraArgs []string) {
	if len(args) == 0 {
		return "", nil
	}
	if strings.HasPrefix(args[0], "-") {
		return "", args
	}
	return strings.TrimSpace(args[0]), args[1:]
}

func resolveRemoveSelection(project, worktree string, groups []app.ProjectWorktrees) (selectedProject, selectedWorktree string, err error) {
	selectedProject = strings.TrimSpace(project)
	selectedWorktree = strings.TrimSpace(worktree)

	if selectedProject == "" {
		selectedProject = inferProjectNameFromCurrentWorktree(groups, true)
	}
	if selectedProject == "" || selectedWorktree == "" {
		var group app.ProjectWorktrees
		selectedProject, group, err = resolveProjectSelection(selectedProject, groups, true)
		if err != nil {
			return "", "", err
		}

		selectedWorktree, _, err = resolveWorktreeSelection(group, selectedWorktree, false, "")
		if err != nil {
			return "", "", err
		}
	}

	return selectedProject, selectedWorktree, nil
}

func writeRemoveSwitchTarget(ctx context.Context, app appAPI, project, worktree string, groups []app.ProjectWorktrees) error {
	if strings.TrimSpace(os.Getenv("WALD_SWITCH_OUT_FILE")) == "" {
		return nil
	}
	if len(groups) == 0 {
		_, listedGroups, err := app.List(ctx)
		if err != nil {
			return err
		}
		groups = listedGroups
	}

	group, ok := findProjectGroup(groups, project)
	if !ok {
		return fmt.Errorf("project %q not found", project)
	}
	defaultWorktree := strings.TrimSpace(group.Project.DefaultBranch)
	if defaultWorktree == "" {
		defaultWorktree = "main"
	}
	if filepath.Clean(worktree) == filepath.Clean(defaultWorktree) {
		return nil
	}

	target, err := app.SwitchTarget(ctx, project, defaultWorktree, "")
	if err != nil {
		return err
	}
	return writeSwitchTarget(target)
}
