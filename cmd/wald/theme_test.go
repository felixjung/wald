package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	internalconfig "github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestResolveThemeProfileLoadsConfiguredVariants(t *testing.T) {
	xdg := t.TempDir()
	writeThemeFile(t, xdg, "catppuccin", `name = "catppuccin"

[variants.latte]
title = "#4c4f69"
label = "#6c6f85"
label_focused = "#8839ef"
required = "#d20f39"
prompt = "#6c6f85"
prompt_focused = "#8839ef"
text = "#4c4f69"
text_focused = "#4c4f69"
placeholder = "#8c8fa1"
help = "#8c8fa1"
error = "#d20f39"

[variants.macchiato]
title = "#cad3f5"
label = "#a5adcb"
label_focused = "#c6a0f6"
required = "#ed8796"
prompt = "#a5adcb"
prompt_focused = "#c6a0f6"
text = "#cad3f5"
text_focused = "#cad3f5"
placeholder = "#8087a2"
help = "#8087a2"
error = "#ed8796"
`)

	cfg := &internalconfig.Config{
		Theme: &internalconfig.ThemeConfig{
			Light: "catppuccin.latte",
			Dark:  "catppuccin.macchiato",
			Mode:  internalconfig.ThemeModeAuto,
		},
	}
	var stderr bytes.Buffer

	profile := resolveThemeProfile(cfg, envWithXDG(xdg), "/home/test", &stderr)
	require.Equal(t, tui.ThemeModeAuto, profile.Mode)
	require.Empty(t, stderr.String())
	require.NotEqual(t, tui.DefaultThemeProfile().Light, profile.Light)
	require.NotEqual(t, tui.DefaultThemeProfile().Dark, profile.Dark)
}

func TestResolveThemeProfileFallsBackMissingSideToDefaultVariant(t *testing.T) {
	xdg := t.TempDir()
	writeThemeFile(t, xdg, "catppuccin", `name = "catppuccin"

[variants.latte]
title = "default"
label = "7"
label_focused = "5"
required = "9"
prompt = "7"
prompt_focused = "5"
text = "default"
text_focused = "default"
placeholder = "7"
help = "7"
error = "9"
`)

	cfg := &internalconfig.Config{
		Theme: &internalconfig.ThemeConfig{
			Light: "catppuccin.latte",
			Mode:  internalconfig.ThemeModeAuto,
		},
	}
	defaultProfile := tui.DefaultThemeProfile()

	profile := resolveThemeProfile(cfg, envWithXDG(xdg), "/home/test", nil)
	require.NotEqual(t, defaultProfile.Light, profile.Light)
	require.Equal(t, defaultProfile.Dark, profile.Dark)
}

func TestResolveThemeProfileWarnsAndFallsBackOnInvalidReference(t *testing.T) {
	cfg := &internalconfig.Config{
		Theme: &internalconfig.ThemeConfig{
			Light: "invalid-reference",
			Dark:  "missing.nope",
			Mode:  internalconfig.ThemeModeAuto,
		},
	}
	defaultProfile := tui.DefaultThemeProfile()
	var stderr bytes.Buffer

	profile := resolveThemeProfile(cfg, func(string) string { return "" }, "/home/test", &stderr)
	require.Equal(t, defaultProfile.Light, profile.Light)
	require.Equal(t, defaultProfile.Dark, profile.Dark)
	require.Contains(t, stderr.String(), `warning: unable to use light theme variant "invalid-reference"`)
	require.Contains(t, stderr.String(), `warning: unable to use dark theme variant "missing.nope"`)
}

func writeThemeFile(t *testing.T, xdgRoot, name, content string) {
	t.Helper()
	path := filepath.Join(xdgRoot, "wald", "themes", name+".toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func envWithXDG(root string) func(string) string {
	return func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return root
		}
		return ""
	}
}
