package main

import (
	"strings"

	"github.com/felixjung/trees/internal/tui"
)

func fieldValue(fields []tui.Field, id string) string {
	for _, field := range fields {
		if field.ID == id {
			return field.Value
		}
	}
	return ""
}

func splitArgs(value string) []string {
	return strings.Fields(value)
}
