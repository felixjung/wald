package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// ThemeModeAuto detects light/dark background at runtime.
	ThemeModeAuto = "auto"
	// ThemeModeLight always uses the light palette.
	ThemeModeLight = "light"
	// ThemeModeDark always uses the dark palette.
	ThemeModeDark = "dark"

	// DefaultThemeName is the built-in theme.
	DefaultThemeName = "default"

	themeDirName = "themes"
)

var (
	themeNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	hexColorPattern  = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	ansiColorNames   = map[string]struct{}{
		"black":          {},
		"red":            {},
		"green":          {},
		"yellow":         {},
		"blue":           {},
		"magenta":        {},
		"cyan":           {},
		"white":          {},
		"bright_black":   {},
		"bright_red":     {},
		"bright_green":   {},
		"bright_yellow":  {},
		"bright_blue":    {},
		"bright_magenta": {},
		"bright_cyan":    {},
		"bright_white":   {},
	}
)

// ThemeFile describes a user-provided TUI theme.
type ThemeFile struct {
	Name        string                   `toml:"name"`
	Description string                   `toml:"description,omitempty"`
	Variants    map[string]ThemeColorSet `toml:"variants"`
}

// ThemeColorSet maps semantic UI tokens to color values.
type ThemeColorSet struct {
	Title         string `toml:"title"`
	Label         string `toml:"label"`
	LabelFocused  string `toml:"label_focused"`
	Required      string `toml:"required"`
	Prompt        string `toml:"prompt"`
	PromptFocused string `toml:"prompt_focused"`
	Text          string `toml:"text"`
	TextFocused   string `toml:"text_focused"`
	Placeholder   string `toml:"placeholder"`
	Help          string `toml:"help"`
	Error         string `toml:"error"`
}

// ResolveThemePath returns the theme file path for a given theme name.
func ResolveThemePath(getenv func(string) string, homeDir, themeName string) (string, error) {
	themeName = strings.TrimSpace(themeName)
	if themeName == "" {
		return "", fmt.Errorf("theme name is required")
	}
	if !themeNamePattern.MatchString(themeName) {
		return "", fmt.Errorf("theme name %q contains invalid characters", themeName)
	}

	xdg := strings.TrimSpace(getenv("XDG_CONFIG_HOME"))
	if xdg == "" {
		xdg = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(xdg, "wald", themeDirName, themeName+".toml"), nil
}

// LoadThemeFile loads a theme file and validates its schema and color values.
func LoadThemeFile(path string) (*ThemeFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read theme file %s: %w", path, err)
	}

	var theme ThemeFile
	if err = toml.Unmarshal(content, &theme); err != nil {
		return nil, fmt.Errorf("parse theme file %s: %w", path, err)
	}
	if err = theme.normalizeAndValidate(); err != nil {
		return nil, fmt.Errorf("validate theme file %s: %w", path, err)
	}
	return &theme, nil
}

func (t *ThemeFile) normalizeAndValidate() error {
	t.Name = strings.TrimSpace(t.Name)
	t.Description = strings.TrimSpace(t.Description)
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(t.Variants) == 0 {
		return fmt.Errorf("at least one variant is required")
	}
	normalizedVariants := make(map[string]ThemeColorSet, len(t.Variants))
	for variantName := range t.Variants {
		colors := t.Variants[variantName]
		normalizedName := strings.ToLower(strings.TrimSpace(variantName))
		if normalizedName == "" {
			return fmt.Errorf("variant name is required")
		}
		if !themeNamePattern.MatchString(normalizedName) {
			return fmt.Errorf("variant name %q contains invalid characters", variantName)
		}
		if err := colors.normalizeAndValidate("variants." + normalizedName); err != nil {
			return err
		}
		normalizedVariants[normalizedName] = colors
	}
	t.Variants = normalizedVariants
	return nil
}

func (c *ThemeColorSet) normalizeAndValidate(prefix string) error {
	values := []struct {
		name  string
		value *string
	}{
		{name: "title", value: &c.Title},
		{name: "label", value: &c.Label},
		{name: "label_focused", value: &c.LabelFocused},
		{name: "required", value: &c.Required},
		{name: "prompt", value: &c.Prompt},
		{name: "prompt_focused", value: &c.PromptFocused},
		{name: "text", value: &c.Text},
		{name: "text_focused", value: &c.TextFocused},
		{name: "placeholder", value: &c.Placeholder},
		{name: "help", value: &c.Help},
		{name: "error", value: &c.Error},
	}

	for _, entry := range values {
		normalized, err := normalizeThemeColorToken(*entry.value)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", prefix, entry.name, err)
		}
		*entry.value = normalized
	}
	return nil
}

func normalizeThemeColorToken(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("color value is required")
	}

	lower := strings.ToLower(value)
	if lower == "default" {
		return "default", nil
	}

	normalizedName := strings.ReplaceAll(strings.ReplaceAll(lower, "-", "_"), " ", "_")
	if _, ok := ansiColorNames[normalizedName]; ok {
		return normalizedName, nil
	}
	if normalizedName == "gray" || normalizedName == "grey" {
		return "bright_black", nil
	}

	if number, err := strconv.Atoi(value); err == nil {
		if number < 0 || number > 255 {
			return "", fmt.Errorf("ansi color index must be between 0 and 255")
		}
		return strconv.Itoa(number), nil
	}

	if hexColorPattern.MatchString(value) {
		return strings.ToLower(value), nil
	}

	return "", fmt.Errorf("unsupported color value %q", value)
}
