package main

import (
	"fmt"
	"io"
	"strings"

	internalconfig "github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/tui"
)

func resolveThemeProfile(
	cfg *internalconfig.Config,
	getenv func(string) string,
	homeDir string,
	stderr io.Writer,
) tui.ThemeProfile {
	profile := tui.DefaultThemeProfile()
	mode := internalconfig.ThemeModeAuto

	if cfg != nil && cfg.Theme != nil {
		if trimmed := strings.TrimSpace(cfg.Theme.Mode); trimmed != "" {
			mode = strings.ToLower(trimmed)
		}
	}
	profile.Mode = tui.ThemeMode(mode)

	if cfg == nil || cfg.Theme == nil {
		return profile
	}

	cache := map[string]*internalconfig.ThemeFile{}

	if ref := strings.TrimSpace(cfg.Theme.Light); ref != "" {
		light, err := resolveVariantColors(ref, getenv, homeDir, cache)
		if err != nil {
			warnThemeVariantFallback(stderr, "light", ref, err)
		} else {
			profile.Light = light
		}
	}
	if ref := strings.TrimSpace(cfg.Theme.Dark); ref != "" {
		dark, err := resolveVariantColors(ref, getenv, homeDir, cache)
		if err != nil {
			warnThemeVariantFallback(stderr, "dark", ref, err)
		} else {
			profile.Dark = dark
		}
	}

	return profile
}

func resolveVariantColors(
	ref string,
	getenv func(string) string,
	homeDir string,
	cache map[string]*internalconfig.ThemeFile,
) (tui.ThemeColors, error) {
	themeName, variantName, err := parseThemeVariantRef(ref)
	if err != nil {
		return tui.ThemeColors{}, err
	}
	themeFile, err := loadThemeByName(themeName, getenv, homeDir, cache)
	if err != nil {
		return tui.ThemeColors{}, err
	}
	variant, ok := themeFile.Variants[strings.ToLower(variantName)]
	if !ok {
		return tui.ThemeColors{}, fmt.Errorf("variant %q not found in theme %q", variantName, themeName)
	}
	colors, err := tui.ColorsFromTokens(themeTokensFromConfig(variant))
	if err != nil {
		return tui.ThemeColors{}, err
	}
	return colors, nil
}

func parseThemeVariantRef(ref string) (themeName, variantName string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("theme variant reference is required")
	}
	separator := strings.LastIndex(ref, ".")
	if separator <= 0 || separator >= len(ref)-1 {
		return "", "", fmt.Errorf("invalid theme variant reference %q; expected <theme>.<variant>", ref)
	}
	themeName = strings.TrimSpace(ref[:separator])
	variantName = strings.TrimSpace(ref[separator+1:])
	if themeName == "" || variantName == "" {
		return "", "", fmt.Errorf("invalid theme variant reference %q; expected <theme>.<variant>", ref)
	}
	return themeName, variantName, nil
}

func loadThemeByName(
	themeName string,
	getenv func(string) string,
	homeDir string,
	cache map[string]*internalconfig.ThemeFile,
) (*internalconfig.ThemeFile, error) {
	normalizedTheme := strings.ToLower(strings.TrimSpace(themeName))
	if file, ok := cache[normalizedTheme]; ok {
		return file, nil
	}
	path, err := internalconfig.ResolveThemePath(getenv, homeDir, normalizedTheme)
	if err != nil {
		return nil, err
	}
	file, err := internalconfig.LoadThemeFile(path)
	if err != nil {
		return nil, err
	}
	cache[normalizedTheme] = file
	return file, nil
}

func themeTokensFromConfig(colors internalconfig.ThemeColorSet) tui.ThemeTokens {
	return tui.ThemeTokens{
		Title:         colors.Title,
		Label:         colors.Label,
		LabelFocused:  colors.LabelFocused,
		Required:      colors.Required,
		Prompt:        colors.Prompt,
		PromptFocused: colors.PromptFocused,
		Text:          colors.Text,
		TextFocused:   colors.TextFocused,
		Placeholder:   colors.Placeholder,
		Help:          colors.Help,
		Error:         colors.Error,
	}
}

func warnThemeVariantFallback(stderr io.Writer, side, reference string, err error) {
	if stderr == nil {
		return
	}
	_, _ = fmt.Fprintf(
		stderr,
		"warning: unable to use %s theme variant %q (%v); falling back to default %s variant\n",
		side,
		reference,
		err,
		side,
	)
}

func isConfigInitCommand(args []string) bool {
	first, second := firstTwoNonFlagArgs(args)
	return first == "config" && second == "init"
}

func firstTwoNonFlagArgs(args []string) (first, second string) {
	values := make([]string, 0, 2)
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "" || arg[0] == '-' {
			continue
		}
		values = append(values, arg)
		if len(values) == 2 {
			break
		}
	}
	switch len(values) {
	case 0:
		return "", ""
	case 1:
		return values[0], ""
	default:
		return values[0], values[1]
	}
}
