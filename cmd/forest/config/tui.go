package config

import "github.com/felixjung/forest/internal/tui"

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
