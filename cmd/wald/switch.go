package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

var (
	isTerminal         = tui.IsTerminal
	selectOption       = tui.Select
	promptFields       = tui.Prompt
	switchThemeOptions []tui.Option
)

func newSwitchCommand(api appAPI) *cli.Command {
	return &cli.Command{
		Name:  "switch",
		Usage: "Switch to an existing project worktree",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project name"},
			&cli.StringFlag{Name: "worktree", Aliases: []string{"w"}, Usage: "worktree name"},
			&cli.StringFlag{Name: "working-dir", Usage: "override project workdir for this switch"},
			&cli.BoolFlag{Name: "create", Usage: "create and switch to a new worktree when missing"},
			&cli.StringFlag{
				Name:  "base",
				Usage: "base reference for created worktree (requires --create)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			projectName := strings.TrimSpace(cmd.String("project"))
			worktreeName := strings.TrimSpace(cmd.String("worktree"))
			workingDirOverride := strings.TrimSpace(cmd.String("working-dir"))
			create := cmd.Bool("create")
			startPoint := strings.TrimSpace(cmd.String("base"))
			if startPoint != "" && !create {
				return cli.Exit("base requires --create", 1)
			}

			_, groups, err := api.List(ctx)
			if err != nil {
				return err
			}
			if projectName == "" {
				projectName = inferProjectNameFromCurrentWorktree(groups, true)
			}

			project, group, err := resolveProjectSelection(projectName, groups, true)
			if err != nil {
				return handleSwitchSelectionError(err)
			}
			worktree, createWorktree, err := resolveWorktreeSelection(group, worktreeName, create)
			if err != nil {
				return handleSwitchSelectionError(err)
			}
			if createWorktree {
				if _, err := api.AddTarget(ctx, project, worktree, startPoint, nil); err != nil {
					return err
				}
			}

			target, err := api.SwitchTarget(ctx, project, worktree, workingDirOverride)
			if err != nil {
				return err
			}
			return writeSwitchTarget(target)
		},
	}
}

func setSwitchThemeOptions(options ...tui.Option) {
	switchThemeOptions = append([]tui.Option(nil), options...)
}

func withSwitchThemeOptions(options ...tui.Option) []tui.Option {
	merged := make([]tui.Option, 0, len(switchThemeOptions)+len(options))
	merged = append(merged, options...)
	merged = append(merged, switchThemeOptions...)
	return merged
}

func handleSwitchSelectionError(err error) error {
	if errors.Is(err, tui.ErrCanceled) {
		return cli.Exit("prompt canceled", 1)
	}
	return err
}

func inferProjectNameFromCurrentWorktree(groups []app.ProjectWorktrees, requireWorktrees bool) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return ""
	}
	return inferProjectNameFromPath(absCWD, groups, requireWorktrees)
}

func inferProjectNameFromPath(path string, groups []app.ProjectWorktrees, requireWorktrees bool) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)

	matches := make(map[string]struct{})
	for _, group := range groups {
		if group.Missing {
			continue
		}
		if requireWorktrees && len(group.Worktrees) == 0 {
			continue
		}
		for _, worktree := range group.Worktrees {
			if isPathWithin(worktree.Path, path) {
				matches[group.Project.Name] = struct{}{}
				break
			}
		}
	}

	if len(matches) != 1 {
		return ""
	}
	for projectName := range matches {
		return projectName
	}
	return ""
}

func isPathWithin(root, path string) bool {
	root = strings.TrimSpace(root)
	path = strings.TrimSpace(path)
	if root == "" || path == "" {
		return false
	}

	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func resolveProjectSelection(projectName string, groups []app.ProjectWorktrees, requireWorktrees bool) (string, app.ProjectWorktrees, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName != "" {
		group, ok := findProjectGroup(groups, projectName)
		if !ok {
			return "", app.ProjectWorktrees{}, fmt.Errorf("project %q not found", projectName)
		}
		if group.Missing {
			return "", app.ProjectWorktrees{}, fmt.Errorf("project %q is not initialized", projectName)
		}
		if requireWorktrees && len(group.Worktrees) == 0 {
			return "", app.ProjectWorktrees{}, fmt.Errorf("project %q has no worktrees", projectName)
		}
		return projectName, group, nil
	}

	if !isTerminal(os.Stdin) {
		return "", app.ProjectWorktrees{}, errors.New("project name is required")
	}

	projectOptions := buildProjectOptions(groups, requireWorktrees)
	if len(projectOptions) == 0 {
		if requireWorktrees {
			return "", app.ProjectWorktrees{}, errors.New("no initialized projects with worktrees found")
		}
		return "", app.ProjectWorktrees{}, errors.New("no initialized projects found")
	}
	selection, err := selectOption(
		"Select project",
		"Type to filter projects...",
		projectOptions,
		withSwitchThemeOptions(tui.WithOutput(os.Stderr))...,
	)
	if err != nil {
		return "", app.ProjectWorktrees{}, err
	}
	group, ok := findProjectGroup(groups, selection.ID)
	if !ok {
		return "", app.ProjectWorktrees{}, fmt.Errorf("project %q not found", selection.ID)
	}
	return selection.ID, group, nil
}

func resolveWorktreeSelection(group app.ProjectWorktrees, worktreeName string, create bool) (selectedWorktree string, createWorktree bool, err error) {
	worktreeOptions := buildWorktreeOptions(group)
	worktreeName = strings.TrimSpace(worktreeName)

	if worktreeName != "" {
		matched, ambiguous := matchWorktreeOption(worktreeOptions, worktreeName)
		if matched != "" {
			return matched, false, nil
		}
		if create {
			return worktreeName, true, nil
		}
		if !isTerminal(os.Stdin) {
			if ambiguous {
				return "", false, fmt.Errorf("worktree %q is ambiguous for project %q", worktreeName, group.Project.Name)
			}
			return "", false, fmt.Errorf("worktree %q not found in project %q", worktreeName, group.Project.Name)
		}
	}

	if create {
		if !isTerminal(os.Stdin) {
			return "", false, errors.New("worktree is required")
		}
		fields, err := promptFields("Create and switch worktree", []tui.Field{
			{ID: "worktree", Label: "Worktree path", Value: worktreeName, Required: true},
		}, withSwitchThemeOptions(tui.WithOutput(os.Stderr))...)
		if err != nil {
			return "", false, err
		}
		value := strings.TrimSpace(fieldValue(fields, "worktree"))
		if value == "" {
			return "", false, errors.New("worktree is required")
		}
		return value, true, nil
	}

	if len(worktreeOptions) == 0 {
		return "", false, fmt.Errorf("project %q has no worktrees", group.Project.Name)
	}
	if !isTerminal(os.Stdin) {
		return "", false, errors.New("worktree is required")
	}

	title := fmt.Sprintf("Select worktree (%s)", group.Project.Name)
	selection, err := selectOption(
		title,
		"Type to filter worktrees...",
		worktreeOptions,
		withSwitchThemeOptions(tui.WithOutput(os.Stderr))...,
	)
	if err != nil {
		return "", false, err
	}
	return selection.ID, false, nil
}

func findProjectGroup(groups []app.ProjectWorktrees, projectName string) (app.ProjectWorktrees, bool) {
	for _, group := range groups {
		if group.Project.Name == projectName {
			return group, true
		}
	}
	return app.ProjectWorktrees{}, false
}

func buildProjectOptions(groups []app.ProjectWorktrees, requireWorktrees bool) []tui.SelectOption {
	options := make([]tui.SelectOption, 0, len(groups))
	for _, group := range groups {
		if group.Missing {
			continue
		}
		if requireWorktrees && len(group.Worktrees) == 0 {
			continue
		}
		options = append(options, tui.SelectOption{
			ID:          group.Project.Name,
			Label:       group.Project.Name,
			Description: fmt.Sprintf("%d worktrees", len(group.Worktrees)),
		})
	}
	sort.Slice(options, func(i, j int) bool {
		return options[i].Label < options[j].Label
	})
	return options
}

func buildWorktreeOptions(group app.ProjectWorktrees) []tui.SelectOption {
	options := make([]tui.SelectOption, 0, len(group.Worktrees))
	for _, entry := range group.Worktrees {
		value := displayWorktreeLabel(group.Root, entry.Path)
		description := buildWorktreeDescription(group, entry)
		options = append(options, tui.SelectOption{
			ID:          value,
			Label:       value,
			Description: description,
		})
	}
	sort.Slice(options, func(i, j int) bool {
		return options[i].Label < options[j].Label
	})
	return options
}

func buildWorktreeDescription(group app.ProjectWorktrees, entry app.WorktreeInfo) string {
	parts := make([]string, 0, 3)
	branch := strings.TrimSpace(entry.Branch)
	if branch != "" {
		parts = append(parts, branch)
	}
	if isDefaultWorktree(group.Root, group.Project.DefaultBranch, entry.Path) {
		parts = append(parts, "default")
	}
	if entry.Detached {
		parts = append(parts, "detached")
	}
	return strings.Join(parts, " | ")
}

func matchWorktreeOption(options []tui.SelectOption, value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	matches := make(map[string]struct{})
	for _, option := range options {
		if option.ID == value || option.Label == value {
			matches[option.ID] = struct{}{}
			continue
		}
		if filepath.Base(option.ID) == value {
			matches[option.ID] = struct{}{}
			continue
		}
		for _, segment := range strings.Split(option.Description, "|") {
			if strings.TrimSpace(segment) == value {
				matches[option.ID] = struct{}{}
			}
		}
	}
	if len(matches) == 1 {
		for id := range matches {
			return id, false
		}
	}
	return "", len(matches) > 1
}

func writeSwitchTarget(target string) error {
	outFile, err := resolveSwitchTargetFile()
	if err != nil {
		return err
	}
	if outFile != "" {
		if err = os.WriteFile(outFile, []byte(target), 0o600); err != nil {
			return fmt.Errorf("write switch target file: %w", err)
		}
		return nil
	}
	_, err = fmt.Fprintln(os.Stdout, target)
	return err
}

func resolveSwitchTargetFile() (string, error) {
	outFile := strings.TrimSpace(os.Getenv("WALD_SWITCH_OUT_FILE"))
	if outFile == "" {
		return "", nil
	}

	outFile = filepath.Clean(outFile)
	if !filepath.IsAbs(outFile) {
		return "", errors.New("WALD_SWITCH_OUT_FILE must be an absolute path")
	}
	if !isPathWithin(os.TempDir(), outFile) {
		return "", fmt.Errorf("WALD_SWITCH_OUT_FILE must be within %s", os.TempDir())
	}

	info, err := os.Lstat(outFile)
	if err != nil {
		return "", fmt.Errorf("stat switch target file: %w", err)
	}
	if !info.Mode().IsRegular() || (info.Mode()&os.ModeSymlink) != 0 {
		return "", errors.New("WALD_SWITCH_OUT_FILE must reference a regular file")
	}

	return outFile, nil
}
