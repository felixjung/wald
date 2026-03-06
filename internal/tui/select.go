package tui

import (
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	list         list.Model
	delegate     list.DefaultDelegate
	theme        *Theme
	themeProfile *ThemeProfile
	isDarkBG     bool
	selected     SelectOption
	hasValue     bool
	canceled     bool
	width        int
}

func newSelectModel(
	title, placeholder string,
	choices []SelectOption,
	theme *Theme,
	themeProfile *ThemeProfile,
	isDarkBG bool,
) *selectModel {
	items := make([]list.Item, 0, len(choices))
	for _, choice := range choices {
		items = append(items, selectItem{option: choice})
	}

	delegate := list.NewDefaultDelegate()

	listModel := list.New(items, delegate, 0, 0)
	listModel.Title = title
	listModel.SetFilteringEnabled(true)
	listModel.SetShowStatusBar(false)
	listModel.SetShowHelp(true)
	listModel.SetShowTitle(true)
	listModel.FilterInput.Prompt = "> "
	listModel.FilterInput.Placeholder = placeholder

	model := &selectModel{
		list:         listModel,
		delegate:     delegate,
		theme:        theme,
		themeProfile: themeProfile,
		isDarkBG:     isDarkBG,
		width:        80,
	}
	model.applyTheme(theme)
	model.resize()
	return model
}

func (m *selectModel) Init() tea.Cmd {
	if !m.shouldTrackBackground() {
		return nil
	}
	return tea.Batch(requestBackgroundColorCmd(), scheduleBackgroundPollCmd())
}

func (m *selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.resize()
	case tea.BackgroundColorMsg:
		m.updateThemeForBackground(msg.IsDark())
	case backgroundPollMsg:
		if m.shouldTrackBackground() {
			return m, tea.Batch(requestBackgroundColorCmd(), scheduleBackgroundPollCmd())
		}
	case tea.KeyPressMsg:
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

func (m *selectModel) View() tea.View {
	frame := lipgloss.NewStyle().Padding(1, 2)
	view := tea.NewView(frame.Render(m.list.View()))
	view.AltScreen = true
	return view
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
	defaultProfile := DefaultThemeProfile()
	config := options{input: os.Stdin, output: os.Stdout, themeProfile: &defaultProfile}
	for _, opt := range opts {
		opt(&config)
	}

	if len(choices) == 0 {
		return SelectOption{}, errors.New("selection options are required")
	}

	initialDark := true
	if dark, ok := detectDarkBackground(config.input, config.output); ok {
		initialDark = dark
	}
	model := newSelectModel(
		title,
		placeholder,
		choices,
		resolveTheme(config, initialDark),
		config.themeProfile,
		initialDark,
	)
	program := tea.NewProgram(
		model,
		tea.WithInput(config.input),
		tea.WithOutput(config.output),
	)
	result, err := program.Run()
	if err != nil {
		return SelectOption{}, err
	}
	finalModel, ok := result.(*selectModel)
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

func (m *selectModel) shouldTrackBackground() bool {
	return m.themeProfile != nil && m.themeProfile.IsAuto()
}

func (m *selectModel) updateThemeForBackground(isDark bool) {
	if !m.shouldTrackBackground() || m.isDarkBG == isDark {
		return
	}
	m.isDarkBG = isDark
	m.applyTheme(m.themeProfile.Theme(isDark))
}

func (m *selectModel) applyTheme(theme *Theme) {
	m.theme = theme

	m.delegate.Styles.NormalTitle = theme.Text
	m.delegate.Styles.SelectedTitle = theme.LabelFocused
	m.delegate.Styles.NormalDesc = theme.Help
	m.delegate.Styles.SelectedDesc = theme.Help
	m.delegate.Styles.DimmedTitle = theme.Placeholder
	m.delegate.Styles.DimmedDesc = theme.Placeholder
	m.delegate.Styles.FilterMatch = theme.PromptFocused
	m.list.SetDelegate(m.delegate)

	m.list.Styles.Title = theme.Title
	m.list.Styles.DefaultFilterCharacterMatch = theme.PromptFocused
	m.list.Styles.HelpStyle = theme.Help
	m.list.Styles.NoItems = theme.Help

	filterStyles := m.list.FilterInput.Styles()
	filterStyles.Focused.Prompt = theme.PromptFocused
	filterStyles.Focused.Text = theme.TextFocused
	filterStyles.Focused.Placeholder = theme.Placeholder
	filterStyles.Blurred.Prompt = theme.Prompt
	filterStyles.Blurred.Text = theme.Text
	filterStyles.Blurred.Placeholder = theme.Placeholder
	m.list.FilterInput.SetStyles(filterStyles)
	m.list.Styles.Filter = filterStyles
}
