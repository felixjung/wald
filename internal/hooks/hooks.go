package hooks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var placeholderPattern = regexp.MustCompile(`\{\{([^{}]*)\}\}`)

// Vars contains template variables for hook command rendering.
type Vars map[string]string

// Runner executes commands for hook operations.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) error
}

// RunAll executes hook commands in order and fails fast on the first error.
func RunAll(ctx context.Context, r Runner, dir, hookName string, commands map[string]string, vars Vars) error {
	if r == nil {
		return errors.New("runner is required")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("hook directory is required")
	}
	names := sortedHookNames(commands)
	for _, name := range names {
		command := strings.TrimSpace(commands[name])
		if command == "" {
			return fmt.Errorf("%s hook %q command is required", hookName, name)
		}
		rendered, err := renderCommand(command, vars)
		if err != nil {
			return fmt.Errorf("%s hook %q template error: %w", hookName, name, err)
		}
		if err := r.Run(ctx, dir, "sh", "-c", rendered); err != nil {
			return fmt.Errorf("%s hook %q failed: %w", hookName, name, err)
		}
	}
	return nil
}

func sortedHookNames(commands map[string]string) []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func renderCommand(command string, vars Vars) (string, error) {
	if len(vars) == 0 {
		vars = Vars{}
	}
	matches := placeholderPattern.FindAllStringSubmatchIndex(command, -1)
	if len(matches) == 0 {
		return command, nil
	}

	var builder strings.Builder
	builder.Grow(len(command))

	last := 0
	for _, match := range matches {
		start := match[0]
		end := match[1]
		keyStart := match[2]
		keyEnd := match[3]

		builder.WriteString(command[last:start])

		key := strings.TrimSpace(command[keyStart:keyEnd])
		value, ok := vars[key]
		if !ok {
			return "", fmt.Errorf("unknown template variable %q", key)
		}
		builder.WriteString(value)
		last = end
	}

	builder.WriteString(command[last:])
	return builder.String(), nil
}
