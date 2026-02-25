package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type fieldModel struct {
	title          string
	field          Field
	input          textinput.Model
	err            error
	canceled       bool
	width          int
	theme          *Theme
	defaultValue   string
	displayDefault string
	showDefault    bool
}

func newFieldModel(title string, field Field, theme *Theme) *fieldModel {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = field.Placeholder
	input.SetValue("")
	input.CharLimit = 0
	input.SetWidth(48)

	defaultValue := field.Value
	if defaultValue == "" {
		defaultValue = field.Default
	}
	showDefault := defaultValue != ""
	displayDefault := defaultValue
	if displayDefault == "" && !field.Required {
		displayDefault = "none"
		showDefault = true
	}

	model := &fieldModel{
		title:          title,
		field:          field,
		input:          input,
		width:          80,
		theme:          theme,
		defaultValue:   defaultValue,
		displayDefault: displayDefault,
		showDefault:    showDefault,
	}
	model.applyStyles()
	return model
}

func (m *fieldModel) Init() tea.Cmd {
	return m.input.Focus()
}

func (m *fieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.resizeInput()
	case tea.KeyPressMsg:
		if msg.String() != "enter" && msg.String() != "ctrl+c" && msg.String() != "esc" {
			m.err = nil
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			usedDefault := false
			if value == "" {
				value = m.defaultValue
				if m.defaultValue != "" || !m.field.Required {
					usedDefault = true
				}
			}
			if strings.TrimSpace(value) == "" && m.field.Required {
				m.err = fmt.Errorf("%s is required", m.field.Label)
				return m, nil
			}
			m.field.Value = value
			m.field.UsedDefault = usedDefault
			return m, tea.Quit
		}
	}

	m.input, _ = m.input.Update(msg)
	return m, nil
}

func (m *fieldModel) View() tea.View {
	sections := make([]string, 0, 6)
	if m.title != "" {
		sections = append(sections, m.theme.Title.Render(m.title))
	}

	labelStyle := m.theme.LabelFocused
	label := m.field.Label
	if m.showDefault {
		label = fmt.Sprintf("%s (%s)", label, m.displayDefault)
	}
	if m.field.Required {
		label = label + " " + m.theme.Required.Render("*")
	}
	sections = append(sections, labelStyle.Render(label), m.input.View())

	if m.err != nil {
		sections = append(sections, m.theme.Error.Render(m.err.Error()))
	}

	sections = append(
		sections,
		m.theme.Help.Render("enter to continue | esc to cancel"),
		m.theme.Help.Render("press enter on empty input to accept the default in parentheses"),
	)

	content := strings.Join(sections, "\n\n")
	frame := lipgloss.NewStyle().Padding(1, 2)
	view := tea.NewView(frame.Render(content))
	view.AltScreen = true
	return view
}

func (m *fieldModel) resizeInput() {
	width := m.width - 6
	if width < 20 {
		width = 20
	}
	if width > 72 {
		width = 72
	}
	m.input.SetWidth(width)
}

func (m *fieldModel) applyStyles() {
	styles := m.input.Styles()
	styles.Focused.Prompt = m.theme.PromptFocused
	styles.Focused.Text = m.theme.TextFocused
	styles.Focused.Placeholder = m.theme.Placeholder
	styles.Blurred.Prompt = m.theme.Prompt
	styles.Blurred.Text = m.theme.Text
	styles.Blurred.Placeholder = m.theme.Placeholder
	m.input.SetStyles(styles)
}
