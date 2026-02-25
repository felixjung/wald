package tui

import (
	"charm.land/lipgloss/v2"
)

// Theme defines the styles used by the TUI prompt.
type Theme struct {
	Title         lipgloss.Style
	Label         lipgloss.Style
	LabelFocused  lipgloss.Style
	Required      lipgloss.Style
	Prompt        lipgloss.Style
	PromptFocused lipgloss.Style
	Text          lipgloss.Style
	TextFocused   lipgloss.Style
	Placeholder   lipgloss.Style
	Help          lipgloss.Style
	Error         lipgloss.Style
}

// DefaultTheme returns the default TUI theme.
func DefaultTheme() *Theme {
	accent := lipgloss.Color("#60A5FA")
	text := lipgloss.Color("#E2E8F0")
	label := lipgloss.Color("#94A3B8")
	muted := lipgloss.Color("#94A3B8")
	errorColor := lipgloss.Color("#F87171")

	return &Theme{
		Title:         lipgloss.NewStyle().Bold(true).Foreground(text),
		Label:         lipgloss.NewStyle().Foreground(label),
		LabelFocused:  lipgloss.NewStyle().Bold(true).Foreground(accent),
		Required:      lipgloss.NewStyle().Bold(true).Foreground(errorColor),
		Prompt:        lipgloss.NewStyle().Foreground(label),
		PromptFocused: lipgloss.NewStyle().Foreground(accent),
		Text:          lipgloss.NewStyle().Foreground(text),
		TextFocused:   lipgloss.NewStyle().Foreground(text),
		Placeholder:   lipgloss.NewStyle().Foreground(muted),
		Help:          lipgloss.NewStyle().Foreground(muted),
		Error:         lipgloss.NewStyle().Foreground(errorColor),
	}
}
