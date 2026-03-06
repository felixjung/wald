package config

import "github.com/felixjung/wald/internal/tui"

func fieldValue(fields []tui.Field, id string) string {
	for _, field := range fields {
		if field.ID == id {
			return field.Value
		}
	}
	return ""
}

func fieldByID(fields []tui.Field, id string) (tui.Field, bool) {
	for _, field := range fields {
		if field.ID == id {
			return field, true
		}
	}
	return tui.Field{}, false
}

func withThemeOptions(base []tui.Option, extra ...tui.Option) []tui.Option {
	options := make([]tui.Option, 0, len(base)+len(extra))
	options = append(options, base...)
	options = append(options, extra...)
	return options
}
