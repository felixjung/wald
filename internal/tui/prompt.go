package tui

import (
	"errors"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
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
	input  *os.File
	output io.Writer
	theme  *Theme
}

// Option customizes prompt behavior.
type Option func(*options)

// WithTheme overrides the prompt theme.
func WithTheme(theme *Theme) Option {
	return func(o *options) {
		o.theme = theme
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
	config := options{input: os.Stdin, output: os.Stdout, theme: DefaultTheme()}
	for _, opt := range opts {
		opt(&config)
	}

	if len(fields) == 0 {
		return fields, nil
	}

	for i := range fields {
		model := newFieldModel(title, fields[i], config.theme)
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
