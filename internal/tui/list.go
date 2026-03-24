package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// ListWorktree describes a worktree line item.
type ListWorktree struct {
	Label    string
	Branch   string
	Detached bool
	Default  bool
}

// ListProject groups worktrees for a project.
type ListProject struct {
	Name      string
	Missing   bool
	Worktrees []ListWorktree
}

func renderListContent(title, root string, projects []ListProject, theme *Theme) string {
	sections := make([]string, 0, 2+len(projects)*3)
	if title != "" {
		sections = append(sections, theme.Title.Render(title))
	}
	if root != "" {
		sections = append(sections, theme.Help.Render(root))
	}

	for i, project := range projects {
		header := theme.Label.Render(project.Name)
		sections = append(sections, header)

		switch {
		case project.Missing:
			sections = append(sections, "  "+theme.Error.Render("not initialized"))
		case len(project.Worktrees) == 0:
			sections = append(sections, "  "+theme.Help.Render("no worktrees"))
		default:
			for _, worktree := range project.Worktrees {
				sections = append(sections, renderWorktreeLine(theme, worktree))
			}
		}

		if i < len(projects)-1 {
			sections = append(sections, "")
		}
	}

	content := strings.Join(sections, "\n")
	frame := lipgloss.NewStyle().Padding(1, 1)
	return frame.Render(content)
}

func renderWorktreeLine(theme *Theme, worktree ListWorktree) string {
	label := worktree.Label
	var tags []string
	if worktree.Branch != "" && worktree.Branch != worktree.Label {
		tags = append(tags, worktree.Branch)
	}
	if worktree.Default {
		tags = append(tags, "default")
	}
	if worktree.Detached {
		tags = append(tags, "detached")
	}
	if len(tags) > 0 {
		label = label + " (" + strings.Join(tags, ", ") + ")"
	}
	return "  - " + theme.Text.Render(label)
}

// List renders a grouped worktree list using Bubble Tea.
func List(title, root string, projects []ListProject, opts ...Option) error {
	defaultProfile := DefaultThemeProfile()
	config := options{input: os.Stdin, output: os.Stdout, themeProfile: &defaultProfile}
	for _, opt := range opts {
		opt(&config)
	}

	theme := resolveTheme(config, true)
	content := renderListContent(title, root, projects, theme)
	_, err := io.WriteString(config.output, content)
	if err != nil {
		return fmt.Errorf("write list output: %w", err)
	}
	return nil
}
