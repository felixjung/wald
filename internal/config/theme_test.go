package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveThemePath(t *testing.T) {
	path, err := ResolveThemePath(
		func(key string) string {
			if key == "XDG_CONFIG_HOME" {
				return "/tmp/xdg"
			}
			return ""
		},
		"/home/test",
		"solarized",
	)
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/tmp/xdg/wald/themes/solarized.toml"), path)
}

func TestResolveThemePathFallsBackToHomeConfig(t *testing.T) {
	path, err := ResolveThemePath(func(string) string { return "" }, "/home/test", "solarized")
	require.NoError(t, err)
	require.Equal(t, filepath.FromSlash("/home/test/.config/wald/themes/solarized.toml"), path)
}

func TestResolveThemePathRejectsInvalidThemeName(t *testing.T) {
	_, err := ResolveThemePath(func(string) string { return "" }, "/home/test", "../oops")
	require.EqualError(t, err, `theme name "../oops" contains invalid characters`)
}

func TestLoadThemeFileValidatesSupportedColorFormats(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	content := `name = "catppuccin"
description = "example"

[variants.latte]
title = "default"
label = "8"
label_focused = "blue"
required = "red"
prompt = "#333"
prompt_focused = "#0055ff"
text = "default"
text_focused = "15"
placeholder = "bright_black"
help = "grey"
error = "1"

[variants.macchiato]
title = "default"
label = "bright_black"
label_focused = "bright_blue"
required = "bright_red"
prompt = "#9ca3af"
prompt_focused = "#60a5fa"
text = "default"
text_focused = "default"
placeholder = "8"
help = "8"
error = "#f87171"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	theme, err := LoadThemeFile(path)
	require.NoError(t, err)
	require.Equal(t, "catppuccin", theme.Name)

	latte, ok := theme.Variants["latte"]
	require.True(t, ok)
	require.Equal(t, "default", latte.Title)
	require.Equal(t, "8", latte.Label)
	require.Equal(t, "blue", latte.LabelFocused)
	require.Equal(t, "bright_black", latte.Help)

	macchiato, ok := theme.Variants["macchiato"]
	require.True(t, ok)
	require.Equal(t, "#f87171", macchiato.Error)
}

func TestLoadThemeFileRejectsInvalidColorValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	content := `name = "broken"

[variants.default]
title = "default"
label = "invalid-color"
label_focused = "blue"
required = "red"
prompt = "default"
prompt_focused = "blue"
text = "default"
text_focused = "default"
placeholder = "8"
help = "8"
error = "1"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := LoadThemeFile(path)
	require.ErrorContains(t, err, `variants.default.label: unsupported color value "invalid-color"`)
}

func TestLoadThemeFileRequiresVariants(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	content := `name = "missing-variants"`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := LoadThemeFile(path)
	require.ErrorContains(t, err, "at least one variant is required")
}

func TestLoadThemeFileRejectsInvalidVariantName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	content := `name = "bad-variant"

[variants."../oops"]
title = "default"
label = "8"
label_focused = "4"
required = "1"
prompt = "8"
prompt_focused = "4"
text = "default"
text_focused = "default"
placeholder = "8"
help = "8"
error = "1"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err := LoadThemeFile(path)
	require.ErrorContains(t, err, `variant name "../oops" contains invalid characters`)
}
