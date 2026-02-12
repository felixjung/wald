package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type listModel struct {
	title    string
	root     string
	projects []ListProject
	theme    *Theme
	width    int
}

func newListModel(title, root string, projects []ListProject, theme *Theme) *listModel {
	return &listModel{
		title:    title,
		root:     root,
		projects: projects,
		theme:    theme,
		width:    80,
	}
}

func (m *listModel) Init() tea.Cmd {
	return tea.Quit
}

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowSizeMsg.Width
	}
	return m, nil
}

func (m *listModel) View() string {
	sections := make([]string, 0, 2+len(m.projects)*3)
	if m.title != "" {
		sections = append(sections, m.theme.Title.Render(m.title))
	}
	if m.root != "" {
		sections = append(sections, m.theme.Help.Render(m.root))
	}

	for i, project := range m.projects {
		header := m.theme.Label.Render(project.Name)
		sections = append(sections, header)

		switch {
		case project.Missing:
			sections = append(sections, "  "+m.theme.Error.Render("not initialized"))
		case len(project.Worktrees) == 0:
			sections = append(sections, "  "+m.theme.Help.Render("no worktrees"))
		default:
			for _, worktree := range project.Worktrees {
				sections = append(sections, renderWorktreeLine(m.theme, worktree))
			}
		}

		if i < len(m.projects)-1 {
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
	config := options{input: os.Stdin, output: os.Stdout, theme: DefaultTheme()}
	for _, opt := range opts {
		opt(&config)
	}
	model := newListModel(title, root, projects, config.theme)
	program := tea.NewProgram(model, tea.WithInput(config.input), tea.WithOutput(config.output))
	_, err := program.Run()
	return err
}
