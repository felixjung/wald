package main

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/forest/internal/app"
	"github.com/felixjung/forest/internal/tui"
)

func newListCommand(api appAPI) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List worktrees grouped by project",
		Action: func(ctx context.Context, _ *cli.Command) error {
			root, groups, err := api.List(ctx)
			if err != nil {
				return err
			}

			projects := make([]tui.ListProject, 0, len(groups))
			for _, group := range groups {
				project := tui.ListProject{
					Name:    group.Project.Name,
					Missing: group.Missing,
				}
				project.Worktrees = append(project.Worktrees, buildListWorktrees(group)...)
				projects = append(projects, project)
			}

			return tui.List("Worktrees", root, projects)
		},
	}
}

func buildListWorktrees(group app.ProjectWorktrees) []tui.ListWorktree {
	worktrees := make([]tui.ListWorktree, 0, len(group.Worktrees))
	for _, entry := range group.Worktrees {
		label := displayWorktreeLabel(group.Root, entry.Path)
		branch := strings.TrimSpace(entry.Branch)
		isDefault := isDefaultWorktree(group.Root, group.Project.DefaultBranch, entry.Path)
		worktrees = append(worktrees, tui.ListWorktree{
			Label:    label,
			Branch:   branch,
			Detached: entry.Detached,
			Default:  isDefault,
		})
	}
	return worktrees
}

func displayWorktreeLabel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func isDefaultWorktree(root, defaultBranch, path string) bool {
	if defaultBranch == "" {
		return false
	}
	defaultPath := filepath.Join(root, defaultBranch)
	return filepath.Clean(defaultPath) == filepath.Clean(path)
}
