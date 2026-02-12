package tui

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectOption is a selectable item in the interactive list.
type SelectOption struct {
	ID          string
	Label       string
	Description string
}

type selectItem struct {
	option SelectOption
}

func (i selectItem) Title() string {
	return i.option.Label
}

func (i selectItem) Description() string {
	return i.option.Description
}

func (i selectItem) FilterValue() string {
	return strings.TrimSpace(i.option.Label + " " + i.option.Description)
}

type selectModel struct {
	list     list.Model
	theme    Theme
	selected SelectOption
	hasValue bool
	canceled bool
	width    int
}

func newSelectModel(title, placeholder string, choices []SelectOption, theme Theme) selectModel {
	items := make([]list.Item, 0, len(choices))
	for _, choice := range choices {
		items = append(items, selectItem{option: choice})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = theme.Text
	delegate.Styles.SelectedTitle = theme.LabelFocused
	delegate.Styles.NormalDesc = theme.Help
	delegate.Styles.SelectedDesc = theme.Help
	delegate.Styles.DimmedTitle = theme.Placeholder
	delegate.Styles.DimmedDesc = theme.Placeholder
	delegate.Styles.FilterMatch = theme.PromptFocused

	listModel := list.New(items, delegate, 0, 0)
	listModel.Title = title
	listModel.Styles.Title = theme.Title
	listModel.Styles.FilterPrompt = theme.PromptFocused
	listModel.Styles.FilterCursor = theme.TextFocused
	listModel.Styles.DefaultFilterCharacterMatch = theme.PromptFocused
	listModel.Styles.HelpStyle = theme.Help
	listModel.Styles.NoItems = theme.Help
	listModel.SetFilteringEnabled(true)
	listModel.SetShowStatusBar(false)
	listModel.SetShowHelp(true)
	listModel.SetShowTitle(true)
	listModel.FilterInput.Prompt = "> "
	listModel.FilterInput.Placeholder = placeholder
	listModel.FilterInput.PlaceholderStyle = theme.Placeholder
	listModel.FilterInput.TextStyle = theme.TextFocused
	listModel.FilterInput.Cursor.Style = theme.TextFocused

	model := selectModel{
		list:  listModel,
		theme: theme,
		width: 80,
	}
	model.resize()
	return model
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.resize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			selected, ok := m.list.SelectedItem().(selectItem)
			if ok {
				m.selected = selected.option
				m.hasValue = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	frame := lipgloss.NewStyle().Padding(1, 2)
	return frame.Render(m.list.View())
}

func (m *selectModel) resize() {
	width := m.width - 6
	if width < 30 {
		width = 30
	}
	height := 16
	m.list.SetSize(width, height)
}

// Select shows a themed, filterable selector and returns the chosen option.
func Select(title, placeholder string, choices []SelectOption, opts ...Option) (SelectOption, error) {
	config := options{input: os.Stdin, output: os.Stdout, theme: DefaultTheme()}
	for _, opt := range opts {
		opt(&config)
	}

	if len(choices) == 0 {
		return SelectOption{}, errors.New("selection options are required")
	}

	model := newSelectModel(title, placeholder, choices, config.theme)
	program := tea.NewProgram(
		model,
		tea.WithInput(config.input),
		tea.WithOutput(config.output),
		tea.WithAltScreen(),
	)
	result, err := program.Run()
	if err != nil {
		return SelectOption{}, err
	}
	finalModel, ok := result.(selectModel)
	if !ok {
		return SelectOption{}, errors.New("unexpected select model")
	}
	if finalModel.canceled {
		return SelectOption{}, ErrCanceled
	}
	if !finalModel.hasValue {
		return SelectOption{}, errors.New("no selection made")
	}
	return finalModel.selected, nil
}
