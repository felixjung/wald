package tui

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestThemeProfileThemeHonorsLightMode(t *testing.T) {
	light := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "blue",
		LabelFocused:  "blue",
		Required:      "red",
		Prompt:        "blue",
		PromptFocused: "blue",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "red",
	})
	dark := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "green",
		LabelFocused:  "green",
		Required:      "yellow",
		Prompt:        "green",
		PromptFocused: "green",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "yellow",
	})
	profile := ThemeProfile{Mode: ThemeModeLight, Light: light, Dark: dark}

	require.Equal(t, profile.Theme(false), profile.Theme(true))
}

func TestThemeProfileThemeHonorsDarkMode(t *testing.T) {
	light := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "blue",
		LabelFocused:  "blue",
		Required:      "red",
		Prompt:        "blue",
		PromptFocused: "blue",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "red",
	})
	dark := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "green",
		LabelFocused:  "green",
		Required:      "yellow",
		Prompt:        "green",
		PromptFocused: "green",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "yellow",
	})
	profile := ThemeProfile{Mode: ThemeModeDark, Light: light, Dark: dark}

	require.Equal(t, profile.Theme(false), profile.Theme(true))
}

func TestFieldModelUpdatesThemeOnBackgroundColorMsgInAutoMode(t *testing.T) {
	light := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "blue",
		LabelFocused:  "blue",
		Required:      "red",
		Prompt:        "blue",
		PromptFocused: "blue",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "red",
	})
	dark := mustColorsFromTokens(t, ThemeTokens{
		Title:         "default",
		Label:         "green",
		LabelFocused:  "green",
		Required:      "yellow",
		Prompt:        "green",
		PromptFocused: "green",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "yellow",
	})
	profile := ThemeProfile{Mode: ThemeModeAuto, Light: light, Dark: dark}

	model := newFieldModel(
		"Title",
		Field{ID: "path", Label: "Path", Required: true},
		profile.Theme(false),
		&profile,
		false,
	)

	updated, _ := model.Update(tea.BackgroundColorMsg{Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}})
	finalModel, ok := updated.(*fieldModel)
	require.True(t, ok)
	require.True(t, finalModel.isDarkBG)
	require.Equal(t, profile.Theme(true), finalModel.theme)
}

func mustColorsFromTokens(t *testing.T, tokens ThemeTokens) ThemeColors {
	t.Helper()
	colors, err := ColorsFromTokens(tokens)
	require.NoError(t, err)
	return colors
}
