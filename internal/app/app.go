package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/felixjung/trees/internal/config"
	"github.com/felixjung/trees/internal/worktree"
)

// Deps wires dependencies for the CLI app.
type Deps struct {
	ConfigLoader func(projectName string) (*config.ProjectConfig, string, error)
	Runner       worktree.Runner
	Stdout       io.Writer
}

// New constructs the CLI application.
func New(deps Deps) *cli.Command {
	if deps.ConfigLoader == nil {
		deps.ConfigLoader = config.LoadProject
	}
	if deps.Stdout == nil {
		deps.Stdout = os.Stdout
	}
	if deps.Runner == nil {
		deps.Runner = worktree.OSRunner{Stdout: os.Stdout, Stderr: os.Stderr}
	}

	return &cli.Command{
		Name:  "trees",
		Usage: "Manage git worktrees for configured projects",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a worktree for a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "branch name", Required: true},
					&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runAdd(ctx, deps, cmd.String("project"), cmd.String("name"))
				},
			},
			{
				Name:  "remove",
				Usage: "Remove a worktree for a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "branch name", Required: true},
					&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runRemove(ctx, deps, cmd.String("project"), cmd.String("name"))
				},
			},
		},
	}
}

func runAdd(ctx context.Context, deps Deps, projectName, branch string) error {
	cfg, _, err := deps.ConfigLoader(projectName)
	if err != nil {
		return err
	}
	repoRoot := config.RepoPath(cfg.Root, cfg.Project.Repo)
	if err = worktree.Add(ctx, deps.Runner, repoRoot, branch); err != nil {
		return err
	}
	path := worktree.WorkdirPath(repoRoot, branch, cfg.Project.Workdir)
	_, err = fmt.Fprintln(deps.Stdout, path)
	return err
}

func runRemove(ctx context.Context, deps Deps, projectName, branch string) error {
	cfg, _, err := deps.ConfigLoader(projectName)
	if err != nil {
		return err
	}
	repoRoot := config.RepoPath(cfg.Root, cfg.Project.Repo)
	if err = worktree.Remove(ctx, deps.Runner, repoRoot, branch); err != nil {
		return err
	}
	path := worktree.WorkdirPath(repoRoot, branch, cfg.Project.Workdir)
	_, err = fmt.Fprintln(deps.Stdout, path)
	return err
}
