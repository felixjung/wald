package tui

import (
	"errors"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ErrCanceled indicates the user canceled the prompt.
var ErrCanceled = errors.New("prompt canceled")

// Field describes a prompt field.
type Field struct {
	ID          string
	Label       string
	Placeholder string
	Value       string
	Default     string
	Required    bool
	UsedDefault bool
}

type options struct {
	input        *os.File
	output       io.Writer
	theme        *Theme
	themeProfile *ThemeProfile
}

// Option customizes prompt behavior.
type Option func(*options)

// WithTheme overrides the prompt theme.
func WithTheme(theme *Theme) Option {
	return func(o *options) {
		o.theme = theme
		o.themeProfile = nil
	}
}

// WithThemeProfile sets an adaptive theme profile.
func WithThemeProfile(profile ThemeProfile) Option {
	return func(o *options) {
		copied := profile
		o.themeProfile = &copied
		o.theme = nil
	}
}

// WithInput overrides the prompt input source.
func WithInput(input *os.File) Option {
	return func(o *options) {
		o.input = input
	}
}

// WithOutput overrides the prompt output destination.
func WithOutput(output io.Writer) Option {
	return func(o *options) {
		o.output = output
	}
}

// Prompt displays interactive prompts, one field at a time, and returns updated fields.
func Prompt(title string, fields []Field, opts ...Option) ([]Field, error) {
	defaultProfile := DefaultThemeProfile()
	config := options{input: os.Stdin, output: os.Stdout, themeProfile: &defaultProfile}
	for _, opt := range opts {
		opt(&config)
	}

	if len(fields) == 0 {
		return fields, nil
	}

	initialDark := true
	if dark, ok := detectDarkBackground(config.input, config.output); ok {
		initialDark = dark
	}

	for i := range fields {
		model := newFieldModel(
			title,
			fields[i],
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
			return nil, err
		}
		finalModel, ok := result.(*fieldModel)
		if !ok {
			return nil, errors.New("unexpected prompt model")
		}
		if finalModel.canceled {
			return nil, ErrCanceled
		}
		fields[i] = finalModel.field
	}

	return fields, nil
}

func resolveTheme(config options, isDark bool) *Theme {
	if config.theme != nil {
		return config.theme
	}
	if config.themeProfile != nil {
		return config.themeProfile.Theme(isDark)
	}
	return DefaultTheme()
}

func detectDarkBackground(input *os.File, output io.Writer) (isDark, detected bool) {
	outFile, ok := output.(*os.File)
	if !ok || input == nil || outFile == nil {
		return false, false
	}
	if !IsTerminal(input) || !IsTerminal(outFile) {
		return false, false
	}
	return lipgloss.HasDarkBackground(input, outFile), true
}

// IsTerminal reports whether the given file descriptor is a terminal.
func IsTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
