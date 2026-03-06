package tui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

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

// ThemeMode controls how a ThemeProfile chooses light or dark palettes.
type ThemeMode string

const (
	// ThemeModeAuto selects the palette based on the terminal background.
	ThemeModeAuto ThemeMode = "auto"
	// ThemeModeLight always uses the light palette.
	ThemeModeLight ThemeMode = "light"
	// ThemeModeDark always uses the dark palette.
	ThemeModeDark ThemeMode = "dark"
)

// ThemeColors defines semantic colors for one palette.
type ThemeColors struct {
	Title         color.Color
	Label         color.Color
	LabelFocused  color.Color
	Required      color.Color
	Prompt        color.Color
	PromptFocused color.Color
	Text          color.Color
	TextFocused   color.Color
	Placeholder   color.Color
	Help          color.Color
	Error         color.Color
}

// ThemeTokens defines semantic color tokens that are parsed into colors.
type ThemeTokens struct {
	Title         string
	Label         string
	LabelFocused  string
	Required      string
	Prompt        string
	PromptFocused string
	Text          string
	TextFocused   string
	Placeholder   string
	Help          string
	Error         string
}

// ThemeProfile contains light and dark palettes and the selected mode.
type ThemeProfile struct {
	Mode  ThemeMode
	Light ThemeColors
	Dark  ThemeColors
}

// DefaultTheme returns the default TUI theme.
func DefaultTheme() *Theme {
	defaultProfile := DefaultThemeProfile()
	return defaultProfile.Theme(true)
}

// DefaultThemeProfile returns the default adaptive theme profile.
func DefaultThemeProfile() ThemeProfile {
	light, _ := ColorsFromTokens(ThemeTokens{
		Title:         "default",
		Label:         "8",
		LabelFocused:  "4",
		Required:      "1",
		Prompt:        "8",
		PromptFocused: "4",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "1",
	})
	dark, _ := ColorsFromTokens(ThemeTokens{
		Title:         "default",
		Label:         "8",
		LabelFocused:  "12",
		Required:      "9",
		Prompt:        "8",
		PromptFocused: "12",
		Text:          "default",
		TextFocused:   "default",
		Placeholder:   "8",
		Help:          "8",
		Error:         "9",
	})
	return ThemeProfile{
		Mode:  ThemeModeAuto,
		Light: light,
		Dark:  dark,
	}
}

// ColorsFromTokens parses supported color tokens into ThemeColors.
func ColorsFromTokens(tokens ThemeTokens) (ThemeColors, error) {
	title, err := parseColorToken(tokens.Title)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("title: %w", err)
	}
	label, err := parseColorToken(tokens.Label)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("label: %w", err)
	}
	labelFocused, err := parseColorToken(tokens.LabelFocused)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("label_focused: %w", err)
	}
	required, err := parseColorToken(tokens.Required)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("required: %w", err)
	}
	prompt, err := parseColorToken(tokens.Prompt)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("prompt: %w", err)
	}
	promptFocused, err := parseColorToken(tokens.PromptFocused)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("prompt_focused: %w", err)
	}
	text, err := parseColorToken(tokens.Text)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("text: %w", err)
	}
	textFocused, err := parseColorToken(tokens.TextFocused)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("text_focused: %w", err)
	}
	placeholder, err := parseColorToken(tokens.Placeholder)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("placeholder: %w", err)
	}
	help, err := parseColorToken(tokens.Help)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("help: %w", err)
	}
	errorColor, err := parseColorToken(tokens.Error)
	if err != nil {
		return ThemeColors{}, fmt.Errorf("error: %w", err)
	}

	return ThemeColors{
		Title:         title,
		Label:         label,
		LabelFocused:  labelFocused,
		Required:      required,
		Prompt:        prompt,
		PromptFocused: promptFocused,
		Text:          text,
		TextFocused:   textFocused,
		Placeholder:   placeholder,
		Help:          help,
		Error:         errorColor,
	}, nil
}

// Theme builds a concrete Theme from this profile using background preference.
func (p ThemeProfile) Theme(isDarkBackground bool) *Theme {
	mode := normalizeThemeMode(p.Mode)
	switch mode {
	case ThemeModeLight:
		return themeFromColors(p.Light)
	case ThemeModeDark:
		return themeFromColors(p.Dark)
	default:
		if isDarkBackground {
			return themeFromColors(p.Dark)
		}
		return themeFromColors(p.Light)
	}
}

// IsAuto reports whether this profile is adaptive.
func (p ThemeProfile) IsAuto() bool {
	return normalizeThemeMode(p.Mode) == ThemeModeAuto
}

func themeFromColors(colors ThemeColors) *Theme {
	return &Theme{
		Title:         lipgloss.NewStyle().Bold(true).Foreground(colors.Title),
		Label:         lipgloss.NewStyle().Foreground(colors.Label),
		LabelFocused:  lipgloss.NewStyle().Bold(true).Foreground(colors.LabelFocused),
		Required:      lipgloss.NewStyle().Bold(true).Foreground(colors.Required),
		Prompt:        lipgloss.NewStyle().Foreground(colors.Prompt),
		PromptFocused: lipgloss.NewStyle().Foreground(colors.PromptFocused),
		Text:          lipgloss.NewStyle().Foreground(colors.Text),
		TextFocused:   lipgloss.NewStyle().Foreground(colors.TextFocused),
		Placeholder:   lipgloss.NewStyle().Foreground(colors.Placeholder),
		Help:          lipgloss.NewStyle().Foreground(colors.Help),
		Error:         lipgloss.NewStyle().Foreground(colors.Error),
	}
}

func parseColorToken(value string) (color.Color, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, fmt.Errorf("color value is required")
	}
	lower := strings.ToLower(normalized)
	if lower == "default" {
		return lipgloss.NoColor{}, nil
	}

	ansiNamed := map[string]color.Color{
		"black":          lipgloss.Black,
		"red":            lipgloss.Red,
		"green":          lipgloss.Green,
		"yellow":         lipgloss.Yellow,
		"blue":           lipgloss.Blue,
		"magenta":        lipgloss.Magenta,
		"cyan":           lipgloss.Cyan,
		"white":          lipgloss.White,
		"bright_black":   lipgloss.BrightBlack,
		"bright_red":     lipgloss.BrightRed,
		"bright_green":   lipgloss.BrightGreen,
		"bright_yellow":  lipgloss.BrightYellow,
		"bright_blue":    lipgloss.BrightBlue,
		"bright_magenta": lipgloss.BrightMagenta,
		"bright_cyan":    lipgloss.BrightCyan,
		"bright_white":   lipgloss.BrightWhite,
		"gray":           lipgloss.BrightBlack,
		"grey":           lipgloss.BrightBlack,
	}

	key := strings.ReplaceAll(strings.ReplaceAll(lower, "-", "_"), " ", "_")
	if namedColor, ok := ansiNamed[key]; ok {
		return namedColor, nil
	}

	if index, err := strconv.Atoi(normalized); err == nil {
		if index < 0 || index > 255 {
			return nil, fmt.Errorf("ansi color index must be between 0 and 255")
		}
		return lipgloss.Color(strconv.Itoa(index)), nil
	}

	if strings.HasPrefix(normalized, "#") {
		colorValue := lipgloss.Color(normalized)
		if _, ok := colorValue.(lipgloss.NoColor); ok {
			return nil, fmt.Errorf("invalid hex color %q", normalized)
		}
		return colorValue, nil
	}

	return nil, fmt.Errorf("unsupported color value %q", value)
}

func normalizeThemeMode(mode ThemeMode) ThemeMode {
	switch ThemeMode(strings.ToLower(strings.TrimSpace(string(mode)))) {
	case ThemeModeLight:
		return ThemeModeLight
	case ThemeModeDark:
		return ThemeModeDark
	default:
		return ThemeModeAuto
	}
}
