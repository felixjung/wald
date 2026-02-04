package tui

import "github.com/charmbracelet/lipgloss"

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

// DefaultTheme returns a theme that adapts to light and dark terminals.
func DefaultTheme() Theme {
	accent := lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}
	text := lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#E2E8F0"}
	label := lipgloss.AdaptiveColor{Light: "#334155", Dark: "#94A3B8"}
	muted := lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
	errorColor := lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}

	return Theme{
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
